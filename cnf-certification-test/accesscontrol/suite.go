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

package accesscontrol

import (
	"fmt"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/namespace"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/rbac"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/resources"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/securitycontextcontainer"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/netutil"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/internal/crclient"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/stringhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

var (
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
	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSecContextIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		testContainerSCC(&env)
	})

	// Security Context: non-compliant capabilities (SYS_ADMIN)
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSysAdminIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		TestSysAdminCapability(&env)
	})

	// Security Context: non-compliant capabilities (NET_ADMIN)
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNetAdminIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		TestNetAdminCapability(&env)
	})

	// Security Context: non-compliant capabilities (NET_RAW)
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNetRawIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		TestNetRawCapability(&env)
	})

	// Security Context: non-compliant capabilities (IPC_LOCK)
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestIpcLockIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		TestIpcLockCapability(&env)
	})

	// container security context: non-root user
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSecConNonRootUserIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		TestSecConRootUser(&env)
	})
	// container security context: privileged escalation
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSecConPrivilegeEscalation)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		TestSecConPrivilegeEscalation(&env)
	})
	// container security context: host port
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestContainerHostPort)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		TestContainerHostPort(&env)
	})
	// container security context: host network
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodHostNetwork)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestPodHostNetwork(&env)
	})
	// pod host path
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodHostPath)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestPodHostPath(&env)
	})
	// pod host ipc
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodHostIPC)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestPodHostIPC(&env)
	})
	// pod host pid
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodHostPID)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestPodHostPID(&env)
	})
	// Namespace
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNamespaceBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Namespaces)
		TestNamespace(&env)
	})
	// pod service account
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodServiceAccountBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestPodServiceAccount(&env)
	})
	// pod role bindings
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodRoleBindingsBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestPodRoleBindings(&env)
	})
	// pod cluster role bindings
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodClusterRoleBindingsBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestPodClusterRoleBindings(&env)
	})
	// automount service token
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodAutomountServiceAccountIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestAutomountServiceToken(&env)
	})
	// one process per container
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestOneProcessPerContainerIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		TestOneProcessPerContainer(&env)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSYSNiceRealtimeCapabilityIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestSYSNiceRealtimeCapability(&env)
	})
	// SYS_PTRACE capability
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSysPtraceCapabilityIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		shareProcessPods := env.GetShareProcessNamespacePods()
		testhelper.SkipIfEmptyAny(ginkgo.Skip, shareProcessPods)
		TestSysPtraceCapability(shareProcessPods)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNamespaceResourceQuotaIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		TestNamespaceResourceQuota(&env)
	})
	// ssh daemons
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNoSSHDaemonsAllowedIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		TestNoSSHDaemonsAllowed(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodRequestsAndLimitsIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		testPodRequestsAndLimits(&env)
	})

	// no 1337 UID's being used by pods
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.Test1337UIDIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		Test1337UIDs(&env)
	})
})

func TestSysAdminCapability(env *provider.TestEnvironment) {
	var badContainers []string
	for _, cut := range env.Containers {
		if cut.SecurityContext != nil && cut.SecurityContext.Capabilities != nil {
			if strings.Contains(cut.SecurityContext.Capabilities.String(), "SYS_ADMIN") {
				tnf.ClaimFilePrintf("Non compliant SYS_ADMIN capability detected in container %s. All container caps: %s", cut.String(), cut.SecurityContext.Capabilities.String())
				badContainers = append(badContainers, cut.String())
			}
		}
	}

	testhelper.AddTestResultLog("Non-compliant", badContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func TestNetAdminCapability(env *provider.TestEnvironment) {
	var badContainers []string
	for _, cut := range env.Containers {
		if cut.SecurityContext != nil && cut.SecurityContext.Capabilities != nil {
			if strings.Contains(cut.SecurityContext.Capabilities.String(), "NET_ADMIN") {
				tnf.ClaimFilePrintf("Non compliant NET_ADMIN capability detected in container %s. All container caps: %s", cut.String(), cut.SecurityContext.Capabilities.String())
				badContainers = append(badContainers, cut.String())
			}
		}
	}

	testhelper.AddTestResultLog("Non-compliant", badContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func TestNetRawCapability(env *provider.TestEnvironment) {
	var badContainers []string
	for _, cut := range env.Containers {
		if cut.SecurityContext != nil && cut.SecurityContext.Capabilities != nil {
			if strings.Contains(cut.SecurityContext.Capabilities.String(), "NET_RAW") {
				tnf.ClaimFilePrintf("Non compliant NET_RAW capability detected in container %s. All container caps: %s", cut.String(), cut.SecurityContext.Capabilities.String())
				badContainers = append(badContainers, cut.String())
			}
		}
	}

	testhelper.AddTestResultLog("Non-compliant", badContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func TestIpcLockCapability(env *provider.TestEnvironment) {
	var badContainers []string
	for _, cut := range env.Containers {
		if cut.SecurityContext != nil && cut.SecurityContext.Capabilities != nil {
			if strings.Contains(cut.SecurityContext.Capabilities.String(), "IPC_LOCK") {
				tnf.ClaimFilePrintf("Non compliant IPC_LOCK capability detected in container %s. All container caps: %s", cut.String(), cut.SecurityContext.Capabilities.String())
				badContainers = append(badContainers, cut.String())
			}
		}
	}

	testhelper.AddTestResultLog("Non-compliant", badContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// TestSecConRootUser verifies that the container is not running as root
func TestSecConRootUser(env *provider.TestEnvironment) {
	var badContainers, badPods []string
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

	testhelper.AddTestResultLog("Non-compliant", badPods, tnf.ClaimFilePrintf, ginkgo.Fail)
	testhelper.AddTestResultLog("Non-compliant", badContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// TestSecConPrivilegeEscalation verifies that the container is not allowed privilege escalation
func TestSecConPrivilegeEscalation(env *provider.TestEnvironment) {
	var badContainers []string
	for _, cut := range env.Containers {
		if cut.SecurityContext != nil && cut.SecurityContext.AllowPrivilegeEscalation != nil {
			if *(cut.SecurityContext.AllowPrivilegeEscalation) {
				tnf.ClaimFilePrintf("AllowPrivilegeEscalation is set to true in container %s.", cut.Podname+"."+cut.Name)
				badContainers = append(badContainers, cut.Podname+"."+cut.Name)
			}
		}
	}

	testhelper.AddTestResultLog("Non-compliant", badContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// TestContainerHostPort tests that containers are not configured with host port privileges
func TestContainerHostPort(env *provider.TestEnvironment) {
	var badContainers []string
	for _, cut := range env.Containers {
		for _, aPort := range cut.Ports {
			if aPort.HostPort != 0 {
				tnf.ClaimFilePrintf("Host port %d is configured in container %s.", aPort.HostPort, cut.String())
				badContainers = append(badContainers, cut.String())
			}
		}
	}

	testhelper.AddTestResultLog("Non-compliant", badContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// TestPodHostNetwork verifies that the pod hostNetwork parameter is not set to true
func TestPodHostNetwork(env *provider.TestEnvironment) {
	var badPods []string
	for _, put := range env.Pods {
		if put.Spec.HostNetwork {
			tnf.ClaimFilePrintf("Host network is set to true in pod %s.", put.Namespace+"."+put.Name)
			badPods = append(badPods, put.Namespace+"."+put.Name)
		}
	}

	testhelper.AddTestResultLog("Non-compliant", badPods, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// TestPodHostPath verifies that the pod hostpath parameter is not set to true
func TestPodHostPath(env *provider.TestEnvironment) {
	var badPods []string
	for _, put := range env.Pods {
		for idx := range put.Spec.Volumes {
			vol := &put.Spec.Volumes[idx]
			if vol.HostPath != nil && vol.HostPath.Path != "" {
				tnf.ClaimFilePrintf("Hostpath path: %s is set in pod %s.", vol.HostPath.Path, put.Namespace+"."+put.Name)
				badPods = append(badPods, put.Namespace+"."+put.Name)
			}
		}
	}

	testhelper.AddTestResultLog("Non-compliant", badPods, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// TestPodHostIPC verifies that the pod hostIpc parameter is not set to true
func TestPodHostIPC(env *provider.TestEnvironment) {
	var badPods []string
	for _, put := range env.Pods {
		if put.Spec.HostIPC {
			tnf.ClaimFilePrintf("HostIpc is set in pod %s.", put.Namespace+"."+put.Name)
			badPods = append(badPods, put.Namespace+"."+put.Name)
		}
	}

	testhelper.AddTestResultLog("Non-compliant", badPods, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// TestPodHostPID verifies that the pod hostPid parameter is not set to true
func TestPodHostPID(env *provider.TestEnvironment) {
	var badPods []string
	for _, put := range env.Pods {
		if put.Spec.HostPID {
			tnf.ClaimFilePrintf("HostPid is set in pod %s.", put.Namespace+"."+put.Name)
			badPods = append(badPods, put.Namespace+"."+put.Name)
		}
	}

	testhelper.AddTestResultLog("Non-compliant", badPods, tnf.ClaimFilePrintf, ginkgo.Fail)
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
	ginkgo.By(fmt.Sprintf("CNF pods should belong to any of the configured Namespaces: %v", env.Namespaces))
	ginkgo.By(fmt.Sprintf("CRs from autodiscovered CRDs should belong only to the configured Namespaces: %v", env.Namespaces))
	invalidCrs, err := namespace.TestCrsNamespaces(env.Crds, env.Namespaces)
	if err != nil {
		ginkgo.Fail("error retrieving CRs")
	}

	invalidCrsNum, claimsLog := namespace.GetInvalidCRsNum(invalidCrs)
	if invalidCrsNum > 0 && len(claimsLog.GetLogLines()) > 0 {
		tnf.ClaimFilePrintf("%s", claimsLog.GetLogLines())
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
	testhelper.AddTestResultLog("Non-compliant", failedPods, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// TestPodRoleBindings verifies that the pod utilizes a valid role binding that does not cross namespaces
//
//nolint:dupl
func TestPodRoleBindings(env *provider.TestEnvironment) {
	ginkgo.By("Should not have RoleBinding in other namespaces")
	failedPods := []string{}

	for _, put := range env.Pods {
		ginkgo.By(fmt.Sprintf("Testing role binding for pod: %s namespace: %s", put.Name, put.Namespace))
		if put.Spec.ServiceAccountName == "" {
			ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
		}

		// Get any rolebindings that do not belong to the pod namespace.
		roleBindings, err := rbac.GetRoleBindings(put.Namespace, put.Spec.ServiceAccountName)
		if err != nil {
			failedPods = append(failedPods, put.Name)
		}

		if len(roleBindings) > 0 {
			logrus.Warnf("Pod: %s/%s has the following role bindings: %s", put.Namespace, put.Name, roleBindings)
			tnf.ClaimFilePrintf("Pod: %s/%s has the following role bindings: %s", put.Namespace, put.Name, roleBindings)
			failedPods = append(failedPods, put.Name)
		}
	}
	testhelper.AddTestResultLog("Non-compliant", failedPods, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// TestPodClusterRoleBindings verifies that the pod utilizes a valid cluster role binding that does not cross namespaces
//
//nolint:dupl
func TestPodClusterRoleBindings(env *provider.TestEnvironment) {
	ginkgo.By("Should not have ClusterRoleBinding in other namespaces")
	failedPods := []string{}

	for _, put := range env.Pods {
		ginkgo.By(fmt.Sprintf("Testing cluster role binding for pod: %s namespace: %s", put.Name, put.Namespace))
		if put.Spec.ServiceAccountName == "" {
			ginkgo.Skip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
		}

		// Get any clusterrolebindings that do not belong to the pod namespace.
		clusterRoleBindings, err := rbac.GetClusterRoleBindings(put.Namespace, put.Spec.ServiceAccountName)
		if err != nil {
			failedPods = append(failedPods, put.Name)
		}

		if len(clusterRoleBindings) > 0 {
			logrus.Warnf("Pod: %s/%s has the following cluster role bindings: %s", put.Namespace, put.Name, clusterRoleBindings)
			tnf.ClaimFilePrintf("Pod: %s/%s has the following cluster role bindings: %s", put.Namespace, put.Name, clusterRoleBindings)
			failedPods = append(failedPods, put.Name)
		}
	}
	testhelper.AddTestResultLog("Non-compliant", failedPods, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func TestAutomountServiceToken(env *provider.TestEnvironment) {
	ginkgo.By("Should have automountServiceAccountToken set to false")

	msg := []string{}
	failedPods := []string{}
	for _, put := range env.Pods {
		ginkgo.By(fmt.Sprintf("check the existence of pod service account %s (ns= %s )", put.Namespace, put.Name))
		if put.Spec.ServiceAccountName == "" {
			tnf.ClaimFilePrintf("Pod %s has been found with an empty service account name.", put.Name)
			ginkgo.Fail("Pod has been found with an empty service account name.")
		}

		// Evaluate the pod's automount service tokens and any attached service accounts
		podPassed, newMsg := rbac.EvaluateAutomountTokens(put.Pod)
		if !podPassed {
			failedPods = append(failedPods, put.Name)
			msg = append(msg, newMsg)
		}
	}

	if len(msg) > 0 {
		tnf.ClaimFilePrintf(strings.Join(msg, ""))
	}

	testhelper.AddTestResultLog("Non-compliant", failedPods, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func TestOneProcessPerContainer(env *provider.TestEnvironment) {
	var badContainers []string

	for _, cut := range env.Containers {
		debugPod := env.DebugPods[cut.NodeName]
		if debugPod == nil {
			ginkgo.Fail(fmt.Sprintf("Debug pod not found on Node: %s", cut.NodeName))
		}
		ocpContext := clientsholder.NewContext(debugPod.Namespace, debugPod.Name, debugPod.Spec.Containers[0].Name)
		pid, err := crclient.GetPidFromContainer(cut, ocpContext)
		if err != nil {
			tnf.ClaimFilePrintf("Could not get PID for: %s, error: %v", cut, err)
			badContainers = append(badContainers, cut.String())
			continue
		}

		nbProcesses, err := getNbOfProcessesInPidNamespace(ocpContext, pid, clientsholder.GetClientsHolder())
		if err != nil {
			tnf.ClaimFilePrintf("Could not get number of processes for: %s, error: %v", cut, err)
			badContainers = append(badContainers, cut.String())
			continue
		}
		if nbProcesses > 1 {
			tnf.ClaimFilePrintf("%s has more than one process running", cut.String())
			badContainers = append(badContainers, cut.String())
		}
	}

	testhelper.AddTestResultLog("Non-compliant", badContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func TestSYSNiceRealtimeCapability(env *provider.TestEnvironment) {
	var containersWithoutSysNice []string

	// Loop through all of the labeled containers and compare their security context capabilities and whether
	// or not the node's kernel is realtime enabled.
	for _, cut := range env.Containers {
		n := env.Nodes[cut.NodeName]
		if n.IsRTKernel() && !strings.Contains(cut.SecurityContext.Capabilities.String(), "SYS_NICE") {
			tnf.ClaimFilePrintf("%s has been found running on a realtime kernel enabled node without SYS_NICE capability.", cut.String())
			containersWithoutSysNice = append(containersWithoutSysNice, cut.String())
		}
	}

	testhelper.AddTestResultLog("Non-compliant", containersWithoutSysNice, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func TestSysPtraceCapability(shareProcessPods []*provider.Pod) {
	var podsWithoutSysPtrace []string
	for _, put := range shareProcessPods {
		sysPtraceEnabled := false
		for _, cut := range put.Containers {
			if cut.SecurityContext == nil ||
				cut.SecurityContext.Capabilities == nil ||
				len(cut.SecurityContext.Capabilities.Add) == 0 {
				continue
			}
			if stringhelper.StringInSlice(cut.SecurityContext.Capabilities.Add, "SYS_PTRACE", false) {
				sysPtraceEnabled = true
				break
			}
		}
		if !sysPtraceEnabled {
			tnf.ClaimFilePrintf("Pod %s has process namespace sharing enabled but no container allowing the SYS_PTRACE capability.", put.String())
			podsWithoutSysPtrace = append(podsWithoutSysPtrace, put.String())
		}
	}
	testhelper.AddTestResultLog("Non-compliant", podsWithoutSysPtrace, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func TestNamespaceResourceQuota(env *provider.TestEnvironment) {
	var namespacesMissingQuotas []string
	ginkgo.By("Testing namespace resource quotas")

	for _, put := range env.Pods {
		// Look through all of the pods and compare their namespace to any potential
		// resource quotas
		foundPodNamespaceRQ := false
		for index := range env.ResourceQuotas {
			// We are just checking for the existence of the resource quota as of right now.
			// Read more about the resource quota object here:
			// https://kubernetes.io/docs/concepts/policy/resource-quotas/
			if put.Namespace == env.ResourceQuotas[index].Namespace {
				foundPodNamespaceRQ = true
				break
			}
		}

		if !foundPodNamespaceRQ {
			namespacesMissingQuotas = append(namespacesMissingQuotas, put.String())
			tnf.ClaimFilePrintf("Pod %s is running in a namespace that does not have a ResourceQuota applied.", put.String())
		}
	}

	testhelper.AddTestResultLog("Non-compliant", namespacesMissingQuotas, tnf.ClaimFilePrintf, ginkgo.Fail)
}

const (
	sshServicePortNumber   = 22
	sshServicePortProtocol = "TCP"
)

func TestNoSSHDaemonsAllowed(env *provider.TestEnvironment) {
	var badPods []string
	var errPods []string

	sshPortInfo := netutil.PortInfo{PortNumber: sshServicePortNumber, Protocol: sshServicePortProtocol}

	for _, put := range env.Pods {
		cut := put.Containers[0]

		listeningPorts, err := netutil.GetListeningPorts(cut)
		if err != nil {
			tnf.ClaimFilePrintf("Failed to get the listening ports on %s, err: %v", cut, err)
			errPods = append(errPods, put.String())
			continue
		}

		if _, ok := listeningPorts[sshPortInfo]; ok {
			tnf.ClaimFilePrintf("Pod %s is running an SSH daemon", put)
			badPods = append(badPods, put.String())
		}
	}

	testhelper.AddTestResultLog("Non-compliant", badPods, tnf.ClaimFilePrintf, ginkgo.Fail)
	testhelper.AddTestResultLog("Error", errPods, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testPodRequestsAndLimits(env *provider.TestEnvironment) {
	var containersMissingRequestsOrLimits []string
	ginkgo.By("Testing container resource requests and limits")

	// Loop through the containers, looking for containers that are missing requests or limits.
	// These need to be defined in order to pass.
	for _, cut := range env.Containers {
		if !resources.HasRequestsAndLimitsSet(cut) {
			containersMissingRequestsOrLimits = append(containersMissingRequestsOrLimits, cut.String())
		}
	}

	testhelper.AddTestResultLog("Non-compliant", containersMissingRequestsOrLimits, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func Test1337UIDs(env *provider.TestEnvironment) {
	// Note this test is only ran as part of the 'extended' test suite.
	ginkgo.By("Testing pods to ensure none are using UID 1337")
	const leetNum = 1337
	var badPods []string
	for _, put := range env.Pods {
		ginkgo.By(fmt.Sprintf("checking if pod %s has a securityContext RunAsUser 1337 (ns= %s)", put.Name, put.Namespace))
		if put.Spec.SecurityContext.RunAsUser != nil && *put.Spec.SecurityContext.RunAsUser == int64(leetNum) {
			tnf.ClaimFilePrintf("Pod: %s/%s is found to use securityContext RunAsUser 1337", put.Namespace, put.Name)
			badPods = append(badPods, put.Name)
		}
	}

	testhelper.AddTestResultLog("Non-compliant", badPods, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// a test for security context that are allowed from the documentation of the cnf
// an allowed one will pass the test

func testContainerSCC(env *provider.TestEnvironment) {
	var badContainer []securitycontextcontainer.PodListcategory
	var goodContainer []securitycontextcontainer.PodListcategory
	highLevelCat := securitycontextcontainer.CategoryID1
	for _, pod := range env.Pods {
		listCategory := securitycontextcontainer.CheckPod(pod)
		for _, cat := range listCategory {
			if cat.Category > securitycontextcontainer.CategoryID1NoUID0 {
				badContainer = append(badContainer, cat)
			} else {
				goodContainer = append(goodContainer, cat)
			}
			if cat.Category > highLevelCat {
				highLevelCat = cat.Category
			}
		}
	}
	logrus.Infof("CNF category (highest container category across all containers):  %s \n", highLevelCat)
	tnf.ClaimFilePrintf("CNF category (highest container category across all containers):  %s \n", highLevelCat)
	logrus.Infof("List of containers that are Category1 or CategoryNoUID0 %+v \n", goodContainer)
	tnf.ClaimFilePrintf("List of containers that are Category1 or CategoryNoUID0 %+v \n", goodContainer)
	logrus.Infof("List of non-compliant containers that are not from Category1 or CategoryNoUID0 - %+v", badContainer)
	testhelper.AddTestResultLog("List of non-compliant containers that are not from Category1 or CategoryNoUID0 - ", badContainer, tnf.ClaimFilePrintf, ginkgo.Fail)
}
