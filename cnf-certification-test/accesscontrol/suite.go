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
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/namespace"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/rbac"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
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

})

// TestSecConCapabilities verifies that non compliant capabilities are not present
func TestSecConCapabilities(env *provider.TestEnvironment) {
	var badContainers []string
	if len(env.Containers) == 0 {
		tnf.GinkgoSkip("No containers to perform test, skipping")
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
	tnf.GomegaExpectSliceBeNil(badContainers)
}

// TestSecConRootUser verifies that the container is not running as root
func TestSecConRootUser(env *provider.TestEnvironment) {
	var badContainers, badPods []string
	if len(env.Containers) == 0 {
		tnf.GinkgoSkip("No containers to perform test, skipping")
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
	tnf.GomegaExpectSliceBeNil(badContainers)
	tnf.GomegaExpectSliceBeNil(badPods)
}

// TestSecConPrivilegeEscalation verifies that the container is not allowed privilege escalation
func TestSecConPrivilegeEscalation(env *provider.TestEnvironment) {
	var badContainers []string
	if len(env.Containers) == 0 {
		tnf.GinkgoSkip("No containers to perform test, skipping")
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
	tnf.GomegaExpectSliceBeNil(badContainers)
}

// TestContainerHostPort tests that containers are not configured with host port privileges
func TestContainerHostPort(env *provider.TestEnvironment) {
	var badContainers []string
	if len(env.Containers) == 0 {
		tnf.GinkgoSkip("No containers to perform test, skipping")
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
	tnf.GomegaExpectSliceBeNil(badContainers)
}

// TestPodHostNetwork verifies that the pod hostNetwork parameter is not set to true
func TestPodHostNetwork(env *provider.TestEnvironment) {
	var badPods []string
	if len(env.Pods) == 0 {
		tnf.GinkgoSkip("No Pods to run test, skipping")
	}
	for _, put := range env.Pods {
		if put.Spec.HostNetwork {
			tnf.ClaimFilePrintf("Host network is set to true in pod %s.", put.Namespace+"."+put.Name)
			badPods = append(badPods, put.Namespace+"."+put.Name)
		}
	}
	tnf.ClaimFilePrintf("bad pods: %v", badPods)
	tnf.GomegaExpectSliceBeNil(badPods)
}

// TestPodHostPath verifies that the pod hostpath parameter is not set to true
func TestPodHostPath(env *provider.TestEnvironment) {
	var badPods []string
	if len(env.Pods) == 0 {
		tnf.GinkgoSkip("No Pods to run test, skipping")
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
	tnf.GomegaExpectSliceBeNil(badPods)
}

// TestPodHostIPC verifies that the pod hostIpc parameter is not set to true
func TestPodHostIPC(env *provider.TestEnvironment) {
	var badPods []string
	if len(env.Pods) == 0 {
		tnf.GinkgoSkip("No Pods to run test, skipping")
	}
	for _, put := range env.Pods {
		if put.Spec.HostIPC {
			tnf.ClaimFilePrintf("HostIpc is set in pod %s.", put.Namespace+"."+put.Name)
			badPods = append(badPods, put.Namespace+"."+put.Name)
		}
	}
	tnf.ClaimFilePrintf("bad pods: %v", badPods)
	tnf.GomegaExpectSliceBeNil(badPods)
}

// TestPodHostPID verifies that the pod hostPid parameter is not set to true
func TestPodHostPID(env *provider.TestEnvironment) {
	var badPods []string
	if len(env.Pods) == 0 {
		tnf.GinkgoSkip("No Pods to run test, skipping")
	}
	for _, put := range env.Pods {
		if put.Spec.HostPID {
			tnf.ClaimFilePrintf("HostPid is set in pod %s.", put.Namespace+"."+put.Name)
			badPods = append(badPods, put.Namespace+"."+put.Name)
		}
	}
	tnf.ClaimFilePrintf("bad pods: %v", badPods)
	tnf.GomegaExpectSliceBeNil(badPods)
}

// Tests namespaces for invalid prefixed and CRs are not defined in namespaces not under test with CRDs under test
func TestNamespace(env *provider.TestEnvironment) {
	tnf.GinkgoBy(fmt.Sprintf("CNF resources' Namespaces should not have any of the following prefixes: %v", invalidNamespacePrefixes))
	var failedNamespaces []string
	for _, namespace := range env.Namespaces {
		tnf.GinkgoBy(fmt.Sprintf("Checking namespace %s", namespace))
		for _, invalidPrefix := range invalidNamespacePrefixes {
			if strings.HasPrefix(namespace, invalidPrefix) {
				tnf.ClaimFilePrintf("Namespace %s has invalid prefix %s", namespace, invalidPrefix)
				failedNamespaces = append(failedNamespaces, namespace)
			}
		}
	}
	if failedNamespacesNum := len(failedNamespaces); failedNamespacesNum > 0 {
		tnf.GinkgoFail(fmt.Sprintf("Found %d Namespaces with an invalid prefix.", failedNamespacesNum))
	}
	tnf.GinkgoBy(fmt.Sprintf("CNF pods' should belong to any of the configured Namespaces: %v", env.Namespaces))
	tnf.GinkgoBy(fmt.Sprintf("CRs from autodiscovered CRDs should belong only to the configured Namespaces: %v", env.Namespaces))
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
		tnf.GinkgoFail(fmt.Sprintf("Found %d CRs belonging to invalid Namespaces.", invalidCrsNum))
	}
}

// TestPodServiceAccount verifies that the pod utilizes a valid service account
func TestPodServiceAccount(env *provider.TestEnvironment) {
	tnf.GinkgoBy("Tests that each pod utilizes a valid service account")
	failedPods := []string{}
	for _, put := range env.Pods {
		tnf.GinkgoBy(fmt.Sprintf("Testing service account for pod %s (ns: %s)", put.Name, put.Namespace))
		if put.Spec.ServiceAccountName == "" {
			tnf.ClaimFilePrintf("Pod %s (ns: %s) doesn't have a service account name.", put.Name, put.Namespace)
			failedPods = append(failedPods, put.Name)
		}
	}
	if n := len(failedPods); n > 0 {
		logrus.Debugf("Pods without service account: %+v", failedPods)
		tnf.GinkgoFail(fmt.Sprintf("%d pods don't have a service account name.", n))
	}
}

// TestPodRoleBindings verifies that the pod utilizes a valid role binding that does not cross namespaces
func TestPodRoleBindings(env *provider.TestEnvironment, testerFuncs rbac.RoleBindingFuncs) {
	tnf.GinkgoBy("Should not have RoleBinding in other namespaces")
	failedPods := []string{}

	for _, put := range env.Pods {
		tnf.GinkgoBy(fmt.Sprintf("Testing role binding for pod: %s namespace: %s", put.Name, put.Namespace))
		if put.Spec.ServiceAccountName == "" {
			tnf.GinkgoSkip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
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
		tnf.GinkgoFail(fmt.Sprintf("%d pods have role bindings in other namespaces.", n))
	}

	if tnf.IsUnitTest() {
		testerFuncs.SetTestingResult(len(failedPods) == 0)
	}
}

// TestPodClusterRoleBindings verifies that the pod utilizes a valid cluster role binding that does not cross namespaces
func TestPodClusterRoleBindings(env *provider.TestEnvironment) {
	tnf.GinkgoBy("Should not have ClusterRoleBinding in other namespaces")
	failedPods := []string{}

	for _, put := range env.Pods {
		tnf.GinkgoBy(fmt.Sprintf("Testing cluster role binding for pod: %s namespace: %s", put.Name, put.Namespace))
		if put.Spec.ServiceAccountName == "" {
			tnf.GinkgoSkip("Can not test when serviceAccountName is empty. Please check previous tests for failures")
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
		tnf.GinkgoFail(fmt.Sprintf("%d pods have cluster role bindings in other namespaces.", n))
	}
}

//nolint:funlen
func TestAutomountServiceToken(env *provider.TestEnvironment, testerFuncs rbac.AutomountTokenFuncs) {
	tnf.GinkgoBy("Should have automountServiceAccountToken set to false")

	msg := []string{}
	failedPods := []string{}
	for _, put := range env.Pods {
		tnf.GinkgoBy(fmt.Sprintf("check the existence of pod service account %s (ns= %s )", put.Namespace, put.Name))
		tnf.GomegaExpectStringNotEmpty(put.Spec.ServiceAccountName)

		// The token can be specified in the pod directly
		// or it can be specified in the service account of the pod
		// if no service account is configured, then the pod will use the configuration
		// of the default service account in that namespace
		// the token defined in the pod has takes precedence
		// the test would pass iif token is explicitly set to false
		// if the token is set to true in the pod, the test would fail right away
		if put.Spec.AutomountServiceAccountToken != nil && *put.Spec.AutomountServiceAccountToken {
			msg = append(msg, fmt.Sprintf("Pod %s:%s is configured with automountServiceAccountToken set to true ", put.Namespace, put.Name))
			failedPods = append(failedPods, put.Name)
			continue
		}

		// Collect information about the service account attached to the pod.
		saAutomountServiceAccountToken, err := testerFuncs.AutomountServiceAccountSetOnSA(put.Spec.ServiceAccountName, put.Namespace)
		if err != nil {
			failedPods = append(failedPods, put.Name)
			continue
		}

		// The pod token is false means the pod is configured properly
		// The pod is not configured and the service account is configured with false means
		// the pod will inherit the behavior `false` and the test would pass
		if (put.Spec.AutomountServiceAccountToken != nil && !*put.Spec.AutomountServiceAccountToken) || (saAutomountServiceAccountToken != nil && !*saAutomountServiceAccountToken) {
			continue
		}

		// the service account is configured with true means all the pods
		// using this service account are not configured properly, register the error
		// message and fail
		if saAutomountServiceAccountToken != nil && *saAutomountServiceAccountToken {
			msg = append(msg, fmt.Sprintf("serviceaccount %s:%s is configured with automountServiceAccountToken set to true, impacting pod %s ", put.Namespace, put.Spec.ServiceAccountName, put.Name))
			failedPods = append(failedPods, put.Name)
		}

		// the token should be set explicitly to false, otherwise, it's a failure
		// register the error message and check the next pod
		if saAutomountServiceAccountToken == nil {
			msg = append(msg, fmt.Sprintf("serviceaccount %s:%s is not configured with automountServiceAccountToken set to false, impacting pod %s ", put.Namespace, put.Spec.ServiceAccountName, put.Name))
			failedPods = append(failedPods, put.Name)
		}
	}

	if len(msg) > 0 {
		tnf.ClaimFilePrintf(strings.Join(msg, ""))
	}

	if n := len(failedPods); n > 0 {
		logrus.Debugf("Pods that failed automount test: %+v", failedPods)
		tnf.ClaimFilePrintf("Pods that failed automount test: %+v", failedPods)
		tnf.GinkgoFail(fmt.Sprintf("% d pods that failed automount test", n))
	}

	if tnf.IsUnitTest() {
		testerFuncs.SetTestingResult(len(failedPods) == 0)
	}
}
