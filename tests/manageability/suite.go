// Copyright (C) 2020-2024 Red Hat, Inc.
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

package manageability

import (
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksadapter"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	checksfn "github.com/redhat-best-practices-for-k8s/checks/manageability"
)

var (
	env provider.TestEnvironment

	beforeEachFn = func(check *checksdb.Check) error {
		env = provider.GetTestEnvironment()
		return nil
	}

	skipIfNoContainersFn = func() (bool, string) {
		if len(env.Containers) == 0 {
			log.Warn("No containers to check...")
			return true, "There are no containers to check. Please check under test labels."
		}
		return false, ""
	}
)

// LoadChecks loads all the checks.
func LoadChecks() {
	log.Debug("Loading %s suite checks", common.ManageabilityTestKey)

	checksGroup := checksdb.NewChecksGroup(common.ManageabilityTestKey).
		WithBeforeEachFn(beforeEachFn)

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("manageability-containers-image-tag")).
		WithSkipCheckFn(skipIfNoContainersFn).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckImageTag).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("manageability-container-port-name-format")).
		WithSkipCheckFn(skipIfNoContainersFn).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckPortNameFormat).MakeCheckFn(&env)))
}

// The name field in the ContainerPort section must be of the form <protocol>[-<suffix>] where <protocol> is one of the following,
// and the optional <suffix> can be chosen by the application. Allowed protocol names: grpc, grpc-web, http, http2, tcp, udp.
var allowedProtocolNames = map[string]bool{"grpc": true, "http": true, "http2": true, "tcp": true, "udp": true}

// containerPortNameFormatCheck is a function that checks if the format of a container port name is valid.
// Return:
//   - bool: true if the format of a container port name is valid, otherwise return false.
func containerPortNameFormatCheck(portName string) bool {
	res := strings.Split(portName, "-")
	return allowedProtocolNames[res[0]]
}
