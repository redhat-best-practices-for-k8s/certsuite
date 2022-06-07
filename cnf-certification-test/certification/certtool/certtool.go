// Copyright (C) 2020-2021 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package certtool

import (
	"time"

	"github.com/test-network-function/cnf-certification-test/internal/api"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
)

const (
	// timeout for eventually call
	CertifiedOperator = "certified-operators"
)

var (
	CertAPIClient api.CertificationValidator
)

// getContainerCertificationRequestFunction returns function that will try to get the certification status (CCP) for a container.
func GetContainerCertificationRequestFunction(id configuration.ContainerImageIdentifier) func() bool {
	return func() bool {
		return CertAPIClient.IsContainerCertified(id.Repository, id.Name, id.Tag, id.Digest)
	}
}

// getOperatorCertificationRequestFunction returns function that will try to get the certification status (OCP) for an operator.
func GetOperatorCertificationRequestFunction(organization, operatorName, ocpversion string) func() bool {
	return func() bool {
		return CertAPIClient.IsOperatorCertified(organization, operatorName, ocpversion)
	}
}

// waitForCertificationRequestToSuccess calls to certificationRequestFunc until it returns true.
func WaitForCertificationRequestToSuccess(certificationRequestFunc func() (interface{}, error), timeout time.Duration) interface{} {
	const pollingPeriod = 1 * time.Second
	var elapsed time.Duration
	var err error
	var result interface{}

	for elapsed < timeout {
		result, err = certificationRequestFunc()

		if err == nil {
			break
		}
		time.Sleep(pollingPeriod)
		elapsed += pollingPeriod
	}
	return result
}

func GetContainersToQuery(env *provider.TestEnvironment) map[configuration.ContainerImageIdentifier]bool {
	containersToQuery := make(map[configuration.ContainerImageIdentifier]bool)
	for _, c := range env.Config.CertifiedContainerInfo {
		containersToQuery[c] = true
	}
	if env.Config.CheckDiscoveredContainerCertificationStatus {
		for _, cut := range env.Containers {
			containersToQuery[cut.ContainerImageIdentifier] = true
		}
	}
	return containersToQuery
}
