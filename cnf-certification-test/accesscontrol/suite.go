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
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/rbac"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/internal/crclient"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
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
		env = provider.GetTestEnvironment()
	})
	ginkgo.ReportAfterEach(results.RecordResult)

	// Security Context: non-compliant capabilities
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestSecConCapabilitiesIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		TestSecConCapabilities(&env)
	})
	// container security context: non-root user
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestSecConNonRootUserIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		TestSecConRootUser(&env)
	})
	// container security context: privileged escalation
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestSecConPrivilegeEscalation)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		TestSecConPrivilegeEscalation(&env)
	})
	// container security context: host port
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestContainerHostPort)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		TestContainerHostPort(&env)
	})
	// container security context: host network
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestPodHostNetwork)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestPodHostNetwork(&env)
	})
	// pod host path
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestPodHostPath)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestPodHostPath(&env)
	})
	// pod host ipc
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestPodHostIPC)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestPodHostIPC(&env)
	})
	// pod host pid
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestPodHostPID)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestPodHostPID(&env)
	})
	// Namespace
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestNamespaceBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Namespaces)
		TestNamespace(&env)
	})
	// pod service account
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestPodServiceAccountBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestPodServiceAccount(&env)
	})
	// pod role bindings
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestPodRoleBindingsBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestPodRoleBindings(&env)
	})
	// pod cluster role bindings
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestPodClusterRoleBindingsBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestPodClusterRoleBindings(&env)
	})
	// automount service token
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestPodAutomountServiceAccountIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestAutomountServiceToken(&env)
	})
	// one process per container
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestOneProcessPerContainerIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		TestOneProcessPerContainer(&env)
	})
})

// TestSecConCapabilities verifies that non compliant capabilities are not present
func TestSecConCapabilities(env *provider.TestEnvironment) {
	var badContainers []string
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
	for _, put := range env.Pods {
		if put.Data.Spec.SecurityContext != nil && put.Data.Spec.SecurityContext.RunAsUser != nil {
			// Check the pod level RunAsUser parameter
			if *(put.Data.Spec.SecurityContext.RunAsUser) == 0 {
				tnf.ClaimFilePrintf("Non compliant run as Root User detected (RunAsUser uid=0) in pod %s", put.Data.Namespace+"."+put.Data.Name)
				badPods = append(badPods, put.Data.Namespace+"."+put.Data.Name)
			}
		}
		for idx := range put.Data.Spec.Containers {
			cut := &(put.Data.Spec.Containers[idx])
			// Check the container level RunAsUser parameter
			if cut.SecurityContext != nil && cut.SecurityContext.RunAsUser != nil {
				if *(cut.SecurityContext.RunAsUser) == 0 {
					tnf.ClaimFilePrintf("Non compliant run as Root User detected (RunAsUser uid=0) in container %s", put.Data.Namespace+"."+put.Data.Name+"."+cut.Name)
					badContainers = append(badContainers, put.Data.Namespace+"."+put.Data.Name+"."+cut.Name)
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
	for _, put := range env.Pods {
		if put.Data.Spec.HostNetwork {
			tnf.ClaimFilePrintf("Host network is set to true in pod %s.", put.Data.Namespace+"."+put.Data.Name)
			badPods = append(badPods, put.Data.Namespace+"."+put.Data.Name)
		}
	}
	tnf.ClaimFilePrintf("bad pods: %v", badPods)
	gomega.Expect(badPods).To(gomega.BeNil())
}

// TestPodHostPath verifies that the pod hostpath parameter is not set to true
func TestPodHostPath(env *provider.TestEnvironment) {
	var badPods []string
	for _, put := range env.Pods {
		for idx := range put.Data.Spec.Volumes {
			vol := &put.Data.Spec.Volumes[idx]
			if vol.HostPath != nil && vol.HostPath.Path != "" {
				tnf.ClaimFilePrintf("An Hostpath path: %s is set in pod %s.", vol.HostPath.Path, put.Data.Namespace+"."+put.Data.Name)
				badPods = append(badPods, put.Data.Namespace+"."+put.Data.Name)
			}
		}
	}
	tnf.ClaimFilePrintf("bad pods: %v", badPods)
	gomega.Expect(badPods).To(gomega.BeNil())
}

// TestPodHostIPC verifies that the pod hostIpc parameter is not set to true
func TestPodHostIPC(env *provider.TestEnvironment) {
	var badPods []string
	for _, put := range env.Pods {
		if put.Data.Spec.HostIPC {
			tnf.ClaimFilePrintf("HostIpc is set in pod %s.", put.Data.Namespace+"."+put.Data.Name)
			badPods = append(badPods, put.Data.Namespace+"."+put.Data.Name)
		}
	}
	tnf.ClaimFilePrintf("bad pods: %v", badPods)
	gomega.Expect(badPods).To(gomega.BeNil())
}

// TestPodHostPID verifies that the pod hostPid parameter is not set to true
func TestPodHostPID(env *provider.TestEnvironment) {
	var badPods []string
	for _, put := range env.Pods {
		if put.Data.Spec.HostPID {
			tnf.ClaimFilePrintf("HostPid is set in pod %s.", put.Data.Namespace+"."+put.Data.Name)
			badPods = append(badPods, put.Data.Namespace+"."+put.Data.Name)
		}
	}
	tnf.ClaimFilePrintf("bad pods: %v", badPods)
	gomega.Expect(badPods).To(gomega.BeNil())
}

// Tests namespaces for invalid prefixed and CRs are not defined in namespaces not under test with CRDs under test
func TestNamespace(env *provider.TestEnvironment) {
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
	invalidCrs, err := namespace.TestCrsNamespaces(env.Crds, env.Namespaces)
	if err != nil {
		ginkgo.Fail("error retrieving CRs")
	}

	invalidCrsNum, claimsLog := namespace.GetInvalidCRsNum(invalidCrs)
	if invalidCrsNum > 0 {
		ginkgo.Fail(fmt.Sprintf("Found %d CRs belonging to invalid namespaces.", invalidCrsNum))
		tnf.ClaimFilePrintf("%s", claimsLog)
	}
}

// TestPodServiceAccount verifies that the pod utilizes a valid service account
func TestPodServiceAccount(env *provider.TestEnvironment) {
	ginkgo.By("Tests that each pod utilizes a valid service account")
	failedPods := []string{}
	for _, put := range env.Pods {
		ginkgo.By(fmt.Sprintf("Testing service account for pod %s (ns: %s)", put.Data.Name, put.Data.Namespace))
		if put.Data.Spec.ServiceAccountName == "" {
			tnf.ClaimFilePrintf("Pod %s (ns: %s) doesn't have a service account name.", put.Data.Name, put.Data.Namespace)
			failedPods = append(failedPods, put.Data.Name)
		}
	}
	if n := len(failedPods); n > 0 {
		logrus.Debugf("Pods without service account: %+v", failedPods)
		ginkgo.Fail(fmt.Sprintf("%d pods don't have a service account name.", n))
	}
}

// TestPodRoleBindings verifies that the pod utilizes a valid role binding that does not cross namespaces
//nolint:dupl
func TestPodRoleBindings(env *provider.TestEnvironment) {
	ginkgo.By("Should not have RoleBinding in other namespaces")
	failedPods := []string{}

	for _, put := range env.Pods {
		ginkgo.By(fmt.Sprintf("Testing role binding for pod: %s namespace: %s", put.Data.Name, put.Data.Namespace))
		if put.Data.Spec.ServiceAccountName == "" {
			ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
		}

		// Get any rolebindings that do not belong to the pod namespace.
		roleBindings, err := rbac.GetRoleBindings(put.Data.Namespace, put.Data.Spec.ServiceAccountName)
		if err != nil {
			failedPods = append(failedPods, put.Data.Name)
		}

		if len(roleBindings) > 0 {
			logrus.Warnf("Pod: %s/%s has the following role bindings: %s", put.Data.Namespace, put.Data.Name, roleBindings)
			tnf.ClaimFilePrintf("Pod: %s/%s has the following role bindings: %s", put.Data.Namespace, put.Data.Name, roleBindings)
			failedPods = append(failedPods, put.Data.Name)
		}
	}
	if n := len(failedPods); n > 0 {
		logrus.Debugf("Pods with role bindings: %+v", failedPods)
		ginkgo.Fail(fmt.Sprintf("%d pods have role bindings in other namespaces.", n))
	}
}

// TestPodClusterRoleBindings verifies that the pod utilizes a valid cluster role binding that does not cross namespaces
//nolint:dupl
func TestPodClusterRoleBindings(env *provider.TestEnvironment) {
	ginkgo.By("Should not have ClusterRoleBinding in other namespaces")
	failedPods := []string{}

	for _, put := range env.Pods {
		ginkgo.By(fmt.Sprintf("Testing cluster role binding for pod: %s namespace: %s", put.Data.Name, put.Data.Namespace))
		if put.Data.Spec.ServiceAccountName == "" {
			ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
		}

		// Get any clusterrolebindings that do not belong to the pod namespace.
		clusterRoleBindings, err := rbac.GetClusterRoleBindings(put.Data.Namespace, put.Data.Spec.ServiceAccountName)
		if err != nil {
			failedPods = append(failedPods, put.Data.Name)
		}

		if len(clusterRoleBindings) > 0 {
			logrus.Warnf("Pod: %s/%s has the following cluster role bindings: %s", put.Data.Namespace, put.Data.Name, clusterRoleBindings)
			tnf.ClaimFilePrintf("Pod: %s/%s has the following cluster role bindings: %s", put.Data.Namespace, put.Data.Name, clusterRoleBindings)
			failedPods = append(failedPods, put.Data.Name)
		}
	}
	if n := len(failedPods); n > 0 {
		logrus.Debugf("Pods with cluster role bindings: %+v", failedPods)
		ginkgo.Fail(fmt.Sprintf("%d pods have cluster role bindings in other namespaces.", n))
	}
}

func TestAutomountServiceToken(env *provider.TestEnvironment) {
	ginkgo.By("Should have automountServiceAccountToken set to false")

	msg := []string{}
	failedPods := []string{}
	for _, put := range env.Pods {
		ginkgo.By(fmt.Sprintf("check the existence of pod service account %s (ns= %s )", put.Data.Namespace, put.Data.Name))
		gomega.Expect(put.Data.Spec.ServiceAccountName).ToNot(gomega.BeEmpty())

		// Evaluate the pod's automount service tokens and any attached service accounts
		podPassed, newMsg := rbac.EvaluateAutomountTokens(put.Data)
		if !podPassed {
			failedPods = append(failedPods, put.Data.Name)
			msg = append(msg, newMsg)
		}
	}

	if len(msg) > 0 {
		tnf.ClaimFilePrintf(strings.Join(msg, ""))
	}

	if n := len(failedPods); n > 0 {
		logrus.Debugf("Pods that failed automount test: %+v", failedPods)
		tnf.ClaimFilePrintf("Pods that failed automount test: %+v", failedPods)
		ginkgo.Fail(fmt.Sprintf("% d pods that failed automount test", n))
	}
}

func TestOneProcessPerContainer(env *provider.TestEnvironment) {
	var badContainers []string

	for _, cut := range env.Containers {
		debugPod := env.DebugPods[cut.NodeName]
		if debugPod == nil {
			ginkgo.Fail(fmt.Sprintf("Debug pod not found on Node: %s", cut.NodeName))
			continue
		}

		ocpContext := clientsholder.Context{
			Namespace:     debugPod.Namespace,
			Podname:       debugPod.Name,
			Containername: debugPod.Spec.Containers[0].Name,
		}

		pid, err := crclient.GetPidFromContainer(cut, ocpContext)
		if err != nil {
			tnf.ClaimFilePrintf("Could not get PID for: %s, error: %s", cut, err)
			badContainers = append(badContainers, cut.String())
			continue
		}

		nbProcesses, err := getNbOfProcessesInPidNamespace(ocpContext, pid, clientsholder.GetClientsHolder())
		if err != nil {
			tnf.ClaimFilePrintf("Could not get number of processes for: %s, error: %s", cut, err)
			badContainers = append(badContainers, cut.String())
			continue
		}
		if nbProcesses > 1 {
			tnf.ClaimFilePrintf("Container %s has more than one process running", cut.String())
			badContainers = append(badContainers, cut.String())
		}
	}

	if n := len(badContainers); n > 0 {
		errMsg := fmt.Sprintf("Number of faulty containers found: %d", n)
		tnf.ClaimFilePrintf(errMsg)
		ginkgo.Fail(errMsg)
	}
}
