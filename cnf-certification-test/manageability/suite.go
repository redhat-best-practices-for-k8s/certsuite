// Copyright (C) 2020-2022 Red Hat, Inc.
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

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

// All actual test code belongs below here.  Utilities belong above.
var _ = ginkgo.Describe(common.ManageabilityTestKey, func() {
	logrus.Debugf("Entering %s suite", common.ManageabilityTestKey)
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})
	ginkgo.ReportAfterEach(results.RecordResult)

	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestContainersImageTag)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAll(ginkgo.Skip, env.Containers)
		testContainersImageTag(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestContainerPortNameFormat)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAll(ginkgo.Skip, env.Containers)
		testContainerPortNameFormat(&env)
	})
})

func testContainersImageTag(env *provider.TestEnvironment) {
	badContainers := []string{}
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " image should be tagged ")
		if cut.ContainerImageIdentifier.Tag == "" {
			badContainers = append(badContainers, cut.String())
			tnf.ClaimFilePrintf("Container %s is missing image tag(s)", cut.String())
		}
	}
	testhelper.AddTestResultLog("Non-compliant", badContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// The name field in the ContainerPort section must be of the form <protocol>[-<suffix>] where <protocol> is one of the following,
// and the optional <suffix> can be chosen by the application. Allowed protocol names: grpc, grpc-web, http, http2, tcp, udp.
var allowedProtocolNames = map[string]bool{"grpc": true, "http": true, "http2": true, "tcp": true, "udp": true}

func containerPortNameFormatCheck(portName string) bool {
	res := strings.Split(portName, "-")
	return allowedProtocolNames[res[0]]
}

func testContainerPortNameFormat(env *provider.TestEnvironment) {
	for _, newProtocol := range env.ValidProtocolNames {
		allowedProtocolNames[newProtocol] = true
	}
	badContainers := []string{}
	for _, cut := range env.Containers {
		for _, port := range cut.Ports {
			if !containerPortNameFormatCheck(port.Name) {
				badContainers = append(badContainers, cut.String())
				tnf.ClaimFilePrintf("%s: ContainerPort %s does not follow the partner naming conventions", cut, port.Name)
			}
		}
	}

	if len(badContainers) > 0 {
		tnf.ClaimFilePrintf("Containers declaring ports whose names do not follow the partner naming conventions: %v", badContainers)
		ginkgo.Fail("Number of containers with port names that do not follow the partner naming conventions: %d", len(badContainers))
	}
}
