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
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
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
//
// It registers a series of checks that validate various aspects of
// container manageability. The function sets up before‑each hooks,
// creates check groups, and adds individual checks with their
// corresponding test functions and skip conditions. No parameters are
// required and it returns nothing.
func LoadChecks() {
	log.Debug("Loading %s suite checks", common.ManageabilityTestKey)

	checksGroup := checksdb.NewChecksGroup(common.ManageabilityTestKey).
		WithBeforeEachFn(beforeEachFn)

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestContainersImageTag)).
		WithSkipCheckFn(skipIfNoContainersFn).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainersImageTag(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestContainerPortNameFormat)).
		WithSkipCheckFn(skipIfNoContainersFn).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainerPortNameFormat(c, &env)
			return nil
		}))
}

// testContainersImageTag checks if any container is missing an image tag and sets the compliance result accordingly.
//
// It iterates over the containers in the provided environment, determines whether each container's image tag is empty,
// categorises them into compliant or non-compliant lists, creates report objects for each case, logs relevant information,
// and finally updates the check result with a pass if all containers are compliant or fail otherwise.
func testContainersImageTag(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		check.LogDebug("Testing Container %q", cut)
		if cut.IsTagEmpty() {
			check.LogError("Container %q is missing image tag(s)", cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container is missing image tag(s)", false))
		} else {
			check.LogInfo("Container %q is tagged with %q", cut, cut.ContainerImageIdentifier.Tag)
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container is tagged", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// The name field in the ContainerPort section must be of the form <protocol>[-<suffix>] where <protocol> is one of the following,
// and the optional <suffix> can be chosen by the application. Allowed protocol names: grpc, grpc-web, http, http2, tcp, udp.
var allowedProtocolNames = map[string]bool{"grpc": true, "http": true, "http2": true, "tcp": true, "udp": true}

// containerPortNameFormatCheck checks if a container port name follows the required format.
//
// It validates that the given string conforms to the pattern <protocol>[-<suffix>],
// where <protocol> must be one of the allowed protocol names such as grpc, grpc-web,
// http, http2, tcp, or udp. The function returns true when the format is valid
// and false otherwise. This check is used to ensure container port names meet
// the specifications before further processing.
func containerPortNameFormatCheck(portName string) bool {
	res := strings.Split(portName, "-")
	return allowedProtocolNames[res[0]]
}

// testContainerPortNameFormat checks that each container declares ports with names following partner naming conventions.
// It analyzes the container port lists, categorizes them into compliant and non-compliant,
// updates the compliance check result, and records detailed report objects for each container.
func testContainerPortNameFormat(check *checksdb.Check, env *provider.TestEnvironment) {
	for _, newProtocol := range env.ValidProtocolNames {
		allowedProtocolNames[newProtocol] = true
	}
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		check.LogDebug("Testing Container %q", cut)
		for _, port := range cut.Ports {
			if !containerPortNameFormatCheck(port.Name) {
				check.LogError("Container %q declares port %q that does not follow the partner naming conventions", cut, port.Name)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "ContainerPort does not follow the partner naming conventions", false).
					AddField(testhelper.ContainerPort, port.Name))
			} else {
				check.LogInfo("Container %q declares port %q that does follow the partner naming conventions", cut, port.Name)
				compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "ContainerPort follows the partner naming conventions", true).
					AddField(testhelper.ContainerPort, port.Name))
			}
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}
