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
		TestNamespace(&env)
	})
	// pod service account
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestPodServiceAccountBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		TestPodServiceAccount(&env)
	})
	// pod role bindings
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestPodRoleBindingsBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		TestPodRoleBindings(&env, rbac.NewRoleBindingTester(clientsholder.GetClientsHolder()))
	})
	// pod cluster role bindings
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestPodClusterRoleBindingsBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		TestPodClusterRoleBindings(&env)
	})
	// automount service token
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestPodAutomountServiceAccountIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		TestAutomountServiceToken(&env, rbac.NewAutomountTokenTester(clientsholder.GetClientsHolder()))
	})
	// one process per container
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestOneProcessPerContainerIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		TestOneProcessPerContainer(&env)
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

	invalidCrsNum := namespace.GetInvalidCRsNum(invalidCrs)
	if invalidCrsNum > 0 {
		ginkgo.Fail(fmt.Sprintf("Found %d CRs belonging to invalid namespaces.", invalidCrsNum))
	}
}

// TestPodServiceAccount verifies that the pod utilizes a valid service account
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

// TestPodRoleBindings verifies that the pod utilizes a valid role binding that does not cross namespaces
func TestPodRoleBindings(env *provider.TestEnvironment, testerFuncs rbac.RoleBindingFuncs) {
	ginkgo.By("Should not have RoleBinding in other namespaces")
	failedPods := []string{}

	for _, put := range env.Pods {
		ginkgo.By(fmt.Sprintf("Testing role binding for pod: %s namespace: %s", put.Name, put.Namespace))
		if put.Spec.ServiceAccountName == "" {
			ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
		}

		// Get any rolebindings that do not belong to the pod namespace.
		roleBindings, err := testerFuncs.GetRoleBindings(put.Namespace, put.Spec.ServiceAccountName)
		if err != nil {
			failedPods = append(failedPods, put.Name)
		}

		if len(roleBindings) > 0 {
			logrus.Warnf("Pod: %s/%s has the following role bindings: %s", put.Namespace, put.Name, roleBindings)
			tnf.ClaimFilePrintf("Pod: %s/%s has the following role bindings: %s", put.Namespace, put.Name, roleBindings)
			failedPods = append(failedPods, put.Name)
		}
	}
	if n := len(failedPods); n > 0 {
		logrus.Debugf("Pods with role bindings: %+v", failedPods)
		ginkgo.Fail(fmt.Sprintf("%d pods have role bindings in other namespaces.", n))
	}
}

// TestPodClusterRoleBindings verifies that the pod utilizes a valid cluster role binding that does not cross namespaces
func TestPodClusterRoleBindings(env *provider.TestEnvironment) {
	ginkgo.By("Should not have ClusterRoleBinding in other namespaces")
	failedPods := []string{}

	for _, put := range env.Pods {
		ginkgo.By(fmt.Sprintf("Testing cluster role binding for pod: %s namespace: %s", put.Name, put.Namespace))
		if put.Spec.ServiceAccountName == "" {
			ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
		}

		// Create a new object with the ability to gather clusterrolebinding specs.
		rbTester := rbac.NewClusterRoleBindingTester(put.Spec.ServiceAccountName, put.Namespace, clientsholder.GetClientsHolder())

		// Get any clusterrolebindings that do not belong to the pod namespace.
		clusterRoleBindings, err := rbTester.GetClusterRoleBindings()
		if err != nil {
			failedPods = append(failedPods, put.Name)
		}

		if len(clusterRoleBindings) > 0 {
			logrus.Warnf("Pod: %s/%s has the following cluster role bindings: %s", put.Namespace, put.Name, clusterRoleBindings)
			tnf.ClaimFilePrintf("Pod: %s/%s has the following cluster role bindings: %s", put.Namespace, put.Name, clusterRoleBindings)
			failedPods = append(failedPods, put.Name)
		}
	}
	if n := len(failedPods); n > 0 {
		logrus.Debugf("Pods with cluster role bindings: %+v", failedPods)
		ginkgo.Fail(fmt.Sprintf("%d pods have cluster role bindings in other namespaces.", n))
	}
}

func TestAutomountServiceToken(env *provider.TestEnvironment, testerFuncs rbac.AutomountTokenFuncs) {
	ginkgo.By("Should have automountServiceAccountToken set to false")

	msg := []string{}
	failedPods := []string{}
	for _, put := range env.Pods {
		ginkgo.By(fmt.Sprintf("check the existence of pod service account %s (ns= %s )", put.Namespace, put.Name))
		gomega.Expect(put.Spec.ServiceAccountName).ToNot(gomega.BeEmpty())

		// Evaluate the pod's automount service tokens and any attached service accounts
		podPassed, newMsg := testerFuncs.EvaluateTokens(put)
		if !podPassed {
			failedPods = append(failedPods, put.Name)
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
		}

		ocpContext := clientsholder.Context{
			Namespace:     debugPod.Namespace,
			Podname:       debugPod.Name,
			Containername: debugPod.Spec.Containers[0].Name,
		}

		pid, err := crclient.GetPidFromContainer(cut, ocpContext)
		if err != nil {
			tnf.ClaimFilePrintf("Could not get PID for: %s, error: %s", cut.StringShort(), err)
			badContainers = append(badContainers, cut.Data.Name)
			continue
		}

		nbProcesses, err := getNbOfProcessesInPidNamespace(ocpContext, pid)
		if err != nil {
			tnf.ClaimFilePrintf("Could not get number of processes for: %s, error: %s", cut.StringShort(), err)
			badContainers = append(badContainers, cut.Data.Name)
			continue
		}
		if nbProcesses > 1 {
			tnf.ClaimFilePrintf("Container %s has more than one process running", cut.Data.Name)
			badContainers = append(badContainers, cut.Data.Name)
		}
	}

	if n := len(badContainers); n > 0 {
		errMsg := fmt.Sprintf("Found %d containers with more than one process running", n)
		tnf.ClaimFilePrintf(errMsg)
		ginkgo.Fail(errMsg)
	}
}
