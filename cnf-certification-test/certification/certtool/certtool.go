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
	"strings"
	"time"

	version "github.com/hashicorp/go-version"
	"github.com/test-network-function/cnf-certification-test/internal/api"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"helm.sh/helm/v3/pkg/release"
)

const (
	// timeout for eventually call
	CertifiedOperator = "certified-operators"
)

var (
	CertAPIClient api.CertAPIClientFuncs
)

// getContainerCertificationRequestFunction returns function that will try to get the certification status (CCP) for a container.
func GetContainerCertificationRequestFunction(id configuration.ContainerImageIdentifier) func() (interface{}, error) {
	return func() (interface{}, error) {
		return CertAPIClient.GetContainerCatalogEntry(id)
	}
}

// getOperatorCertificationRequestFunction returns function that will try to get the certification status (OCP) for an operator.
func GetOperatorCertificationRequestFunction(organization, operatorName, ocpversion string) func() (interface{}, error) {
	return func() (interface{}, error) {
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

// CompareVersion compare between versions
func CompareVersion(ver1, ver2 string) bool {
	ourKubeVersion, _ := version.NewVersion(ver1)
	kubeVersion := strings.ReplaceAll(ver2, " ", "")[2:]
	if strings.Contains(kubeVersion, "<") {
		kubever := strings.Split(kubeVersion, "<")
		minVersion, _ := version.NewVersion(kubever[0])
		maxVersion, _ := version.NewVersion(kubever[1])
		if ourKubeVersion.GreaterThanOrEqual(minVersion) && ourKubeVersion.LessThan(maxVersion) {
			return true
		}
	} else {
		kubever := strings.Split(kubeVersion, "-")
		minVersion, _ := version.NewVersion(kubever[0])
		if ourKubeVersion.GreaterThanOrEqual(minVersion) {
			return true
		}
	}
	return false
}

func IsReleaseCertified(helm *release.Release, ourKubeVersion string, out api.ChartStruct) bool {
	for _, entryList := range out.Entries {
		for _, entry := range entryList {
			if entry.Name == helm.Chart.Metadata.Name && entry.Version == helm.Chart.Metadata.Version {
				if entry.KubeVersion != "" {
					if CompareVersion(ourKubeVersion, entry.KubeVersion) {
						return true
					}
				} else {
					return true
				}
			}
		}
	}
	return false
}
