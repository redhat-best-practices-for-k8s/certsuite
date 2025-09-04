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

// LoadChecks Initializes the manageability checks group and registers test functions
//
// The function creates a new checks group for manageability, logs the loading
// action, and adds two checks: one verifying container image tags and another
// validating container port naming conventions. Each check is configured with a
// skip condition that bypasses it if no containers are present and supplies the
// appropriate test logic via closures. The checks are then registered in the
// global checks database for execution during testing.
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

// testContainersImageTag Verifies that each container has a non-empty image tag
//
// The function iterates over all containers in the test environment, checking
// whether their image tags are present. Containers lacking tags are logged as
// errors and added to a non‑compliant list; those with tags are logged as
// info and added to a compliant list. Finally, it records both lists as the
// result of the compliance check.
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

// containerPortNameFormatCheck Verifies that a container port name starts with an allowed protocol
//
// The function splits the provided name on hyphens, extracts the first segment,
// and checks whether this segment is present in the global map of permitted
// protocols. It returns true if the protocol prefix is valid; otherwise it
// returns false.
func containerPortNameFormatCheck(portName string) bool {
	res := strings.Split(portName, "-")
	return allowedProtocolNames[res[0]]
}

// testContainerPortNameFormat Verifies that container port names match partner naming conventions
//
// The function iterates over all containers in the test environment, checking
// each declared port name against a list of allowed protocol prefixes. It logs
// errors for non‑compliant ports and records both compliant and
// non‑compliant objects. Finally, it sets the check result with these lists.
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
