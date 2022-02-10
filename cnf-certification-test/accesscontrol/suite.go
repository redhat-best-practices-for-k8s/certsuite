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
	// Security Context: non-compliant capabilities
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestSecConCapabilitiesIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		TestSecConCapabilities(&env)
	})
	// Security context: non-root user
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestSecConNonRootUserIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		TestSecConRootUser(&env)
	})
	// Security context: privileged escalation
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestSecConPrivilegeEscalation)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		TestSecConPrivilegeEscalation(&env)
	})
})

// TestSecConCapabilities verrifies that non compliant capabilities are not present
func TestSecConCapabilities(env *provider.TestEnvironment) {
	var badContainers []string
	var errContainers []string
	if len(env.Containers) == 0 {
		ginkgo.Skip("No containers to perform test, skipping")
	}
	for _, cut := range env.Containers {
		if cut == nil {
			errContainers = append(errContainers, cut.Podname+"."+cut.Data.Name)
			continue
		}
		if cut.Data.SecurityContext != nil && cut.Data.SecurityContext.Capabilities != nil {
			for _, ncc := range nonCompliantCapabilites {
				if strings.Contains(cut.Data.SecurityContext.Capabilities.String(), ncc) {
					tnf.ClaimFilePrintf("Non compliant %s capability detected in container %s. All container caps: %s", ncc, cut.Data.Name, cut.Data.SecurityContext.Capabilities.String())
					badContainers = append(badContainers, cut.Podname+"."+cut.Data.Name)
				}
			}
			logrus.Infof("test %s", cut.Data.SecurityContext.Capabilities.String())
		}
	}
	tnf.ClaimFilePrintf("bad containers: %v", badContainers)
	tnf.ClaimFilePrintf("err containers: %v", errContainers)
	gomega.Expect(badContainers).To(gomega.BeNil())
	gomega.Expect(errContainers).To(gomega.BeNil())
}

// TestSecConRootUser verifies that the container is not running as root
func TestSecConRootUser(env *provider.TestEnvironment) {
	var badContainers, badPods, errPods []string
	if len(env.Containers) == 0 {
		ginkgo.Skip("No containers to perform test, skipping")
	}
	for _, put := range env.Pods {
		if put == nil {
			errPods = append(errPods, put.Name)
			continue
		}
		if put.Spec.SecurityContext != nil && put.Spec.SecurityContext.RunAsUser != nil {
			// Check the pod level RunAsUser parameter
			if *(put.Spec.SecurityContext.RunAsUser) == 0 {
				tnf.ClaimFilePrintf("Non compliant run as Root User detected (RunAsUser uid=0) in pod %s", put.Name)
				badPods = append(badPods, put.Name)
			}
		}
		for idx := range put.Spec.Containers {
			cut := &(put.Spec.Containers[idx])
			// Check the container level RunAsUser parameter
			if cut.SecurityContext != nil && cut.SecurityContext.RunAsUser != nil {
				if *(cut.SecurityContext.RunAsUser) == 0 {
					tnf.ClaimFilePrintf("Non compliant run as Root User detected (RunAsUser uid=0) in container %s.%s", put.Name, cut.Name)
					badContainers = append(badContainers, put.Name+"."+cut.Name)
				}
			}
		}
	}
	tnf.ClaimFilePrintf("bad pods: %v", badPods)
	tnf.ClaimFilePrintf("err pods: %v", errPods)
	tnf.ClaimFilePrintf("bad containers: %v", badContainers)

	gomega.Expect(badContainers).To(gomega.BeNil())
	gomega.Expect(badPods).To(gomega.BeNil())
	gomega.Expect(errPods).To(gomega.BeNil())
}

// TestSecConPrivilegeEscalation verifies that the container is not allowed privilege escalation
func TestSecConPrivilegeEscalation(env *provider.TestEnvironment) {
	var badContainers []string
	var errContainers []string
	if len(env.Containers) == 0 {
		ginkgo.Skip("No containers to perform test, skipping")
	}
	for _, cut := range env.Containers {
		if cut == nil {
			errContainers = append(errContainers, cut.Podname+"."+cut.Data.Name)
			continue
		}
		if cut.Data.SecurityContext != nil && cut.Data.SecurityContext.AllowPrivilegeEscalation != nil {
			if *(cut.Data.SecurityContext.AllowPrivilegeEscalation) {
				tnf.ClaimFilePrintf("AllowPrivilegeEscalation is set to true in container %s.", cut.Data.Name)
				badContainers = append(badContainers, cut.Podname+"."+cut.Data.Name)
			}
		}
	}
	tnf.ClaimFilePrintf("bad containers: %v", badContainers)
	tnf.ClaimFilePrintf("err containers: %v", errContainers)
	gomega.Expect(badContainers).To(gomega.BeNil())
	gomega.Expect(errContainers).To(gomega.BeNil())
}
