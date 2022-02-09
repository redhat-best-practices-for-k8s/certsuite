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

package accesscontrol

import (
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

var (
	nonCompliantCapabilites = []string{"NET_ADMIN", "SYS_ADMIN", "NET_RAW", "IPC_LOCK"}
)

var _ = ginkgo.Describe(common.AccessControlTestKey, func() {

	logrus.Debugf("Entering %s suite", common.AccessControlTestKey)
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		provider.BuildTestEnvironment()
		env = provider.GetTestEnvironment()
	})
	//
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestSecConCapabilitiesIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		TestSecConCapabilities(&env)
	})
})

// TestSecConCapabilities
func TestSecConCapabilities(env *provider.TestEnvironment) {
	var badContainers []string
	var errContainers []string
	for _, cut := range env.Containers {
		if cut == nil {
			errContainers = append(errContainers, cut.Data.Name)
			continue
		}
		if cut.Data.SecurityContext != nil && cut.Data.SecurityContext.Capabilities != nil {
			for _, ncc := range nonCompliantCapabilites {
				if strings.Contains(cut.Data.SecurityContext.Capabilities.String(), ncc) {
					tnf.ClaimFilePrintf("Non compliant %s capability detected in container %s. All container caps: %s", ncc, cut.Data.Name, cut.Data.SecurityContext.Capabilities.String())
					badContainers = append(badContainers, cut.Data.Name)
				}
			}
			logrus.Infof("test %s", cut.Data.SecurityContext.Capabilities.String())
		}
	}
	if len(badContainers) > 0 {
		tnf.ClaimFilePrintf("bad containers: %v", badContainers)
	}
	if len(errContainers) > 0 {
		tnf.ClaimFilePrintf("err containers: %v", errContainers)
	}
	gomega.Expect(badContainers).To(gomega.BeNil())
	gomega.Expect(errContainers).To(gomega.BeNil())
}
