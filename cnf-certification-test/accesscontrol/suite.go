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
	"fmt"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/namespace"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

var (
	nonCompliantCapabilities = []string{"NET_ADMIN", "SYS_ADMIN", "NET_RAW", "IPC_LOCK"}
	invalidNamespacePrefixes = []string{
		"default",
		"openshift-",
		"istio-",
		"aspenmesh-",
	}
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
	// container security context: non-root user
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestSecConNonRootUserIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		TestSecConRootUser(&env)
	})
	// container security context: privileged escalation
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestSecConPrivilegeEscalation)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		TestSecConPrivilegeEscalation(&env)
	})
	// container security context: host port
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestContainerHostPort)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		TestContainerHostPort(&env)
	})
	// container security context: host network
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestPodHostNetwork)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		TestPodHostNetwork(&env)
	})
	// pod host path
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestPodHostPath)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		TestPodHostPath(&env)
	})
	// pod host ipc
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestPodHostIPC)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		TestPodHostIPC(&env)
	})
	// pod host pid
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestPodHostPID)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		TestPodHostPID(&env)
	})
	// Namespace
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestNamespaceBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testNamespace(&env)
	})
	// pod service account
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestPodServiceAccountBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		TestPodServiceAccount(&env)
	})
})

// TestSecConCapabilities verifies that non compliant capabilities are not present
func TestSecConCapabilities(env *provider.TestEnvironment) {
	var badContainers []string
	if len(env.Containers) == 0 {
		ginkgo.Skip("No containers to perform test, skipping")
	}
	for _, cut := range env.Containers {
		if cut.Data.SecurityContext != nil && cut.Data.SecurityContext.Capabilities != nil {
			for _, ncc := range nonCompliantCapabilities {
				if strings.Contains(cut.Data.SecurityContext.Capabilities.String(), ncc) {
					tnf.ClaimFilePrintf("Non compliant %s capability detected in container %s. All container caps: %s", ncc, cut.Namespace+"."+cut.Podname+"."+cut.Data.Name, cut.Data.SecurityContext.Capabilities.String())
					badContainers = append(badContainers, cut.Namespace+"."+cut.Podname+"."+cut.Data.Name)
				}
			}
		}
	}
	tnf.ClaimFilePrintf("bad containers: %v", badContainers)
	gomega.Expect(badContainers).To(gomega.BeNil())
}

// TestSecConRootUser verifies that the container is not running as root
func TestSecConRootUser(env *provider.TestEnvironment) {
	var badContainers, badPods []string
	if len(env.Containers) == 0 {
		ginkgo.Skip("No containers to perform test, skipping")
	}
	for _, put := range env.Pods {
		if put.Spec.SecurityContext != nil && put.Spec.SecurityContext.RunAsUser != nil {
			// Check the pod level RunAsUser parameter
			if *(put.Spec.SecurityContext.RunAsUser) == 0 {
				tnf.ClaimFilePrintf("Non compliant run as Root User detected (RunAsUser uid=0) in pod %s", put.Namespace+"."+put.Name)
				badPods = append(badPods, put.Namespace+"."+put.Name)
			}
		}
		for idx := range put.Spec.Containers {
			cut := &(put.Spec.Containers[idx])
			// Check the container level RunAsUser parameter
			if cut.SecurityContext != nil && cut.SecurityContext.RunAsUser != nil {
				if *(cut.SecurityContext.RunAsUser) == 0 {
					tnf.ClaimFilePrintf("Non compliant run as Root User detected (RunAsUser uid=0) in container %s", put.Namespace+"."+put.Name+"."+cut.Name)
					badContainers = append(badContainers, put.Namespace+"."+put.Name+"."+cut.Name)
				}
			}
		}
	}
	tnf.ClaimFilePrintf("bad pods: %v", badPods)
	tnf.ClaimFilePrintf("bad containers: %v", badContainers)
	gomega.Expect(badContainers).To(gomega.BeNil())
	gomega.Expect(badPods).To(gomega.BeNil())
}

// TestSecConPrivilegeEscalation verifies that the container is not allowed privilege escalation
func TestSecConPrivilegeEscalation(env *provider.TestEnvironment) {
	var badContainers []string
	if len(env.Containers) == 0 {
		ginkgo.Skip("No containers to perform test, skipping")
	}
	for _, cut := range env.Containers {
		if cut.Data.SecurityContext != nil && cut.Data.SecurityContext.AllowPrivilegeEscalation != nil {
			if *(cut.Data.SecurityContext.AllowPrivilegeEscalation) {
				tnf.ClaimFilePrintf("AllowPrivilegeEscalation is set to true in container %s.", cut.Podname+"."+cut.Data.Name)
				badContainers = append(badContainers, cut.Podname+"."+cut.Data.Name)
			}
		}
	}
	tnf.ClaimFilePrintf("bad containers: %v", badContainers)
	gomega.Expect(badContainers).To(gomega.BeNil())
}

// TestContainerHostPort tests that containers are not configured with host port privileges
func TestContainerHostPort(env *provider.TestEnvironment) {
	var badContainers []string
	if len(env.Containers) == 0 {
		ginkgo.Skip("No containers to perform test, skipping")
	}
	for _, cut := range env.Containers {
		for _, aPort := range cut.Data.Ports {
			if aPort.HostPort != 0 {
				tnf.ClaimFilePrintf("Host port %d is configured in container %s.", aPort.HostPort, cut.Namespace+"."+cut.Podname+"."+cut.Data.Name)
				badContainers = append(badContainers, cut.Namespace+"."+cut.Podname+"."+cut.Data.Name)
			}
		}
	}
	tnf.ClaimFilePrintf("bad containers: %v", badContainers)
	gomega.Expect(badContainers).To(gomega.BeNil())
}

// TestPodHostNetwork verifies that the pod hostNetwork parameter is not set to true
func TestPodHostNetwork(env *provider.TestEnvironment) {
	var badPods []string
	if len(env.Pods) == 0 {
		ginkgo.Skip("No Pods to run test, skipping")
	}
	for _, put := range env.Pods {
		if put.Spec.HostNetwork {
			tnf.ClaimFilePrintf("Host network is set to true in pod %s.", put.Namespace+"."+put.Name)
			badPods = append(badPods, put.Namespace+"."+put.Name)
		}
	}
	tnf.ClaimFilePrintf("bad pods: %v", badPods)
	gomega.Expect(badPods).To(gomega.BeNil())
}

// TestPodHostPath verifies that the pod hostpath parameter is not set to true
func TestPodHostPath(env *provider.TestEnvironment) {
	var badPods []string
	if len(env.Pods) == 0 {
		ginkgo.Skip("No Pods to run test, skipping")
	}
	for _, put := range env.Pods {
		for idx := range put.Spec.Volumes {
			vol := &put.Spec.Volumes[idx]
			if vol.HostPath != nil && vol.HostPath.Path != "" {
				tnf.ClaimFilePrintf("An Hostpath path: %s is set in pod %s.", vol.HostPath.Path, put.Namespace+"."+put.Name)
				badPods = append(badPods, put.Namespace+"."+put.Name)
			}
		}
	}
	tnf.ClaimFilePrintf("bad pods: %v", badPods)
	gomega.Expect(badPods).To(gomega.BeNil())
}

// TestPodHostIPC verifies that the pod hostIpc parameter is not set to true
func TestPodHostIPC(env *provider.TestEnvironment) {
	var badPods []string
	if len(env.Pods) == 0 {
		ginkgo.Skip("No Pods to run test, skipping")
	}
	for _, put := range env.Pods {
		if put.Spec.HostIPC {
			tnf.ClaimFilePrintf("HostIpc is set in pod %s.", put.Namespace+"."+put.Name)
			badPods = append(badPods, put.Namespace+"."+put.Name)
		}
	}
	tnf.ClaimFilePrintf("bad pods: %v", badPods)
	gomega.Expect(badPods).To(gomega.BeNil())
}

// TestPodHostPID verifies that the pod hostPid parameter is not set to true
func TestPodHostPID(env *provider.TestEnvironment) {
	var badPods []string
	if len(env.Pods) == 0 {
		ginkgo.Skip("No Pods to run test, skipping")
	}
	for _, put := range env.Pods {
		if put.Spec.HostPID {
			tnf.ClaimFilePrintf("HostPid is set in pod %s.", put.Namespace+"."+put.Name)
			badPods = append(badPods, put.Namespace+"."+put.Name)
		}
	}
	tnf.ClaimFilePrintf("bad pods: %v", badPods)
	gomega.Expect(badPods).To(gomega.BeNil())
}

// Tests namespaces for invalid prefixed and CRs are not defined in namespaces not under test with CRDs under test
func testNamespace(env *provider.TestEnvironment) {
	ginkgo.By(fmt.Sprintf("CNF resources' Namespaces should not have any of the following prefixes: %v", invalidNamespacePrefixes))
	var failedNamespaces []string
	for _, namespace := range env.Namespaces {
		ginkgo.By(fmt.Sprintf("Checking namespace %s", namespace))
		for _, invalidPrefix := range invalidNamespacePrefixes {
			if strings.HasPrefix(namespace, invalidPrefix) {
				tnf.ClaimFilePrintf("Namespace %s has invalid prefix %s", namespace, invalidPrefix)
				failedNamespaces = append(failedNamespaces, namespace)
			}
		}
	}
	if failedNamespacesNum := len(failedNamespaces); failedNamespacesNum > 0 {
		ginkgo.Fail(fmt.Sprintf("Found %d Namespaces with an invalid prefix.", failedNamespacesNum))
	}
	ginkgo.By(fmt.Sprintf("CNF pods' should belong to any of the configured Namespaces: %v", env.Namespaces))
	ginkgo.By(fmt.Sprintf("CRs from autodiscovered CRDs should belong only to the configured Namespaces: %v", env.Namespaces))
	invalidCrs, _ := namespace.TestCrsNamespaces(env.Crds, env.Namespaces)

	invalidCrsNum := 0
	if invalidCrdsNum := len(invalidCrs); invalidCrdsNum > 0 {
		for crdName, namespaces := range invalidCrs {
			for namespace, crNames := range namespaces {
				for _, crName := range crNames {
					tnf.ClaimFilePrintf("crName=%s namespace=%s is invalid (crd=%s)", crName, namespace, crdName)
					invalidCrsNum++
				}
			}
		}
		ginkgo.Fail(fmt.Sprintf("Found %d CRs belonging to invalid Namespaces.", invalidCrsNum))
	}
}

func TestPodServiceAccount(env *provider.TestEnvironment) {
	ginkgo.By("Tests that each pod utilizes a valid service account")
	failedPods := []string{}
	for _, put := range env.Pods {
		ginkgo.By(fmt.Sprintf("Testing service account for pod %s (ns: %s)", put.Name, put.Namespace))
		if put.Spec.ServiceAccountName == "" {
			tnf.ClaimFilePrintf("Pod %s (ns: %s) doesn't have a service account name.", put.Name, put.Namespace)
			failedPods = append(failedPods, put.Name)
		}
	}
	if n := len(failedPods); n > 0 {
		logrus.Debugf("Pods without service account: %+v", failedPods)
		ginkgo.Fail(fmt.Sprintf("%d pods don't have a service account name.", n))
	}
}
