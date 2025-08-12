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

package accesscontrol

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/crclient"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/podhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/stringhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/accesscontrol/namespace"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/accesscontrol/resources"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/accesscontrol/securitycontextcontainer"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common/rbac"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/networking/netutil"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/networking/services"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	nodePort              = "NodePort"
	defaultServiceAccount = "default"
)

var (
	invalidNamespacePrefixes = []string{
		"default",
		"openshift-",
		"istio-",
		"aspenmesh-",
	}

	knownContainersToSkip = map[string]bool{"kube-rbac-proxy": true}
)

var (
	env provider.TestEnvironment

	beforeEachFn = func(check *checksdb.Check) error {
		env = provider.GetTestEnvironment()
		return nil
	}
)

// LoadChecks registers all access control checks to the test suite.
//
// It creates check groups and adds individual checks, each with its
// corresponding execution and skip functions. The function returns a
// teardown function that will be called after all tests have run. No
// parameters are required; it operates on package‑level state.
func LoadChecks() {
	log.Debug("Loading %s suite checks", common.AccessControlTestKey)

	checksGroup := checksdb.NewChecksGroup(common.AccessControlTestKey).
		WithBeforeEachFn(beforeEachFn)

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestSecContextIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainerSCC(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestSysAdminIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testSysAdminCapability(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestNetAdminIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNetAdminCapability(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestNetRawIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNetRawCapability(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestIpcLockIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testIpcLockCapability(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestBpfIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testBpfCapability(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestSecConNonRootUserIDIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testSecConRunAsNonRoot(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestSecConPrivilegeEscalation)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testSecConPrivilegeEscalation(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestSecConReadOnlyFilesystem)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testSecConReadOnlyFilesystem(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestContainerHostPort)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainerHostPort(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestPodHostNetwork)).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodHostNetwork(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestPodHostPath)).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodHostPath(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestPodHostIPC)).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodHostIPC(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestPodHostPID)).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodHostPID(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestNamespaceBestPracticesIdentifier)).
		WithSkipCheckFn(testhelper.GetNoNamespacesSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNamespace(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestPodServiceAccountBestPracticesIdentifier)).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodServiceAccount(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestPodRoleBindingsBestPracticesIdentifier)).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodRoleBindings(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestPodClusterRoleBindingsBestPracticesIdentifier)).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodClusterRoleBindings(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestPodAutomountServiceAccountIdentifier)).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testAutomountServiceToken(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestOneProcessPerContainerIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOneProcessPerContainer(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestSYSNiceRealtimeCapabilityIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithSkipCheckFn(testhelper.GetNoNodesWithRealtimeKernelSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testSYSNiceRealtimeCapability(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestSysPtraceCapabilityIdentifier)).
		WithSkipCheckFn(testhelper.GetSharedProcessNamespacePodsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testSysPtraceCapability(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestNamespaceResourceQuotaIdentifier)).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNamespaceResourceQuota(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestNoSSHDaemonsAllowedIdentifier)).
		WithSkipCheckFn(testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNoSSHDaemonsAllowed(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestPodRequestsIdentifier)).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodRequests(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.Test1337UIDIdentifier)).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			test1337UIDs(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestServicesDoNotUseNodeportsIdentifier)).
		WithSkipCheckFn(testhelper.GetNoServicesUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNodePort(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestCrdRoleIdentifier)).
		WithSkipCheckFn(testhelper.GetNoCrdsUnderTestSkipFn(&env), testhelper.GetNoNamespacesSkipFn(&env), testhelper.GetNoRolesSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testCrdRoles(c, &env)
			return nil
		}))
}

// isContainerCapabilitySet checks if a capability was explicitly added to a container's security context.
//
// It examines the SecurityContext.Capabilities.Add slice and returns true
// if the specified capability name appears in that list, otherwise it returns false.
func isContainerCapabilitySet(containerCapabilities *corev1.Capabilities, capability string) bool {
	if containerCapabilities == nil {
		return false
	}

	if len(containerCapabilities.Add) == 0 {
		return false
	}

	if stringhelper.StringInSlice(containerCapabilities.Add, corev1.Capability("ALL"), true) ||
		stringhelper.StringInSlice(containerCapabilities.Add, corev1.Capability(capability), true) {
		return true
	}

	return false
}

// checkForbiddenCapability verifies that containers do not use forbidden capabilities.
//
// It examines each container in the provided slice, determines if a disallowed
// capability is set using isContainerCapabilitySet, and logs relevant information.
// Containers that comply with the capability restrictions are added to compliantObjects,
// while those that violate the restrictions are added to nonCompliantObjects. The function
// returns two slices of report objects representing compliant and non‑compliant containers.
func checkForbiddenCapability(containers []*provider.Container, capability string, logger *log.Logger) (compliantObjects, nonCompliantObjects []*testhelper.ReportObject) {
	for _, cut := range containers {
		logger.Info("Testing Container %q", cut)
		compliant := true

		switch {
		case cut.SecurityContext == nil:
		case cut.SecurityContext.Capabilities == nil:
		case isContainerCapabilitySet(cut.SecurityContext.Capabilities, capability):
			compliant = false
		}

		if compliant {
			logger.Info("Container %q does not use non-compliant capability %q", cut, capability)
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "No forbidden capability "+capability+" detected in container", true))
		} else {
			logger.Error("Non compliant %q capability detected in container %q. All container caps: %q", capability, cut, cut.SecurityContext.Capabilities)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Non compliant capability "+capability+" in container", false).AddField(testhelper.SCCCapability, capability))
		}
	}
	return compliantObjects, nonCompliantObjects
}

// testSysAdminCapability checks that a system administrator does not possess forbidden capabilities in the given environment.
//
// It receives a check object and a test environment.
// The function logs its progress, verifies that the sysadmin role lacks prohibited capabilities,
// and records the result of the check using SetResult.
func testSysAdminCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(env.Containers, "SYS_ADMIN", check.GetLogger())
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testNetAdminCapability verifies that a process lacking network admin capability is correctly rejected by the system.
//
// It receives a checksdb.Check and a TestEnvironment, then attempts to create a privileged operation.
// The function logs its actions, calls checkForbiddenCapability to ensure the capability is denied,
// and records the test result using SetResult. No value is returned.
func testNetAdminCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(env.Containers, "NET_ADMIN", check.GetLogger())
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testNetRawCapability checks that the system denies raw network capabilities in the test environment.
//
// It takes a Check object and a TestEnvironment, performs capability verification by invoking
// checkForbiddenCapability on the relevant process identifiers, logs the result, and sets the
// outcome of the check using SetResult. No value is returned.
func testNetRawCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(env.Containers, "NET_RAW", check.GetLogger())
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testIpcLockCapability checks whether the IPC_LOCK capability is correctly
// forbidden in the container environment.
//
// It receives a Check object and a TestEnvironment, logs the check outcome,
// verifies that the IPC_LOCK capability is denied, and records the result.
// The function uses helper functions to perform the verification and to set
// the test status.
func testIpcLockCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(env.Containers, "IPC_LOCK", check.GetLogger())
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testBpfCapability verifies that a pod does not have forbidden BPF capabilities.
//
// It receives the current check and the test environment, logs relevant
// information, calls checkForbiddenCapability to perform the validation,
// and records the result using SetResult. The function is used as a
// helper in the access control test suite.
func testBpfCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(env.Containers, "BPF", check.GetLogger())
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testSecConRunAsNonRoot verifies that containers are not allowed to run as root.
//
// It examines the provided check and environment, logs informational messages,
// retrieves containers configured to run as root, and records results in
// pod and container report objects. If any such containers are found,
// it marks them with a failure result; otherwise the test passes.
func testSecConRunAsNonRoot(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, put := range env.Pods {
		check.LogInfo("Testing pod %s/%s", put.Namespace, put.Name)
		nonCompliantContainers, nonComplianceReason := put.GetRunAsNonRootFalseContainers(knownContainersToSkip)
		if len(nonCompliantContainers) == 0 {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is configured with RunAsNonRoot=true or RunAsUser!=0 at pod or container level.", true))
		} else {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "One or more containers of the pod are running with root user", false))
			for index := range nonCompliantContainers {
				check.LogError("Pod %s/%s, container %q is not compliant: %s", put.Namespace, put.Name, nonCompliantContainers[index].Name, nonComplianceReason[index])

				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(put.Namespace, put.Name, nonCompliantContainers[index].Name,
					nonComplianceReason[index], false))
			}
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testSecConPrivilegeEscalation verifies that the container is not allowed privilege escalation.
//
// It receives a Check object and a TestEnvironment, performs security context checks
// to ensure the container does not have privilege escalation enabled,
// logs relevant information or errors, updates the report objects,
// and sets the result status accordingly.
func testSecConPrivilegeEscalation(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		check.LogInfo("Testing Container %q", cut)
		privEscFound := false
		if cut.SecurityContext != nil && cut.SecurityContext.AllowPrivilegeEscalation != nil {
			if *(cut.SecurityContext.AllowPrivilegeEscalation) {
				check.LogError("AllowPrivilegeEscalation is set to true in Container %q.", cut)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "AllowPrivilegeEscalation is set to true", false))
				privEscFound = true
			}
		}

		if !privEscFound {
			check.LogInfo("AllowPrivilegeEscalation is set to false in Container %q.", cut)
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "AllowPrivilegeEscalation is not set to true", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testSecConReadOnlyFilesystem verifies that a container has a readonly file system access.
//
// testSecConReadOnlyFilesystem checks whether the root filesystem of the tested container is set to read‑only.
// It logs the result, creates pod reports for success or failure, and updates the check status accordingly.
func testSecConReadOnlyFilesystem(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, pod := range env.Pods {
		check.LogInfo("Testing Pod %q in namespace %q", pod.Name, pod.Namespace)
		for _, cut := range pod.Containers {
			check.LogInfo("Testing Container %q in Pod %q", cut.Name, pod.Name)
			if cut.IsReadOnlyRootFilesystem(check.GetLogger()) {
				check.LogInfo("Container %q in Pod %q has a read-only root filesystem.", cut.Name, pod.Name)
				compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(pod.Namespace, pod.Name, "Container has a read-only root filesystem", true))
			} else {
				check.LogError("Container %q in Pod %q does not have a read-only root filesystem.", cut.Name, pod.Name)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(pod.Namespace, pod.Name, "Container does not have a read-only root filesystem", false))
			}
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testContainerHostPort verifies that containers are not configured with host port privileges.
//
// It examines the container specifications in the provided check and environment,
// ensuring that no container exposes or binds to host ports. For each container
// inspected, it records a report object indicating success or failure.
// The function logs progress and errors using the test environment's logging
// facilities. No value is returned; results are recorded through side effects on
// the supplied *checksdb.Check structure.
func testContainerHostPort(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		check.LogInfo("Testing Container %q", cut)
		hostPortFound := false
		for _, aPort := range cut.Ports {
			if aPort.HostPort != 0 {
				check.LogError("Host port %d is configured in Container %q.", aPort.HostPort, cut)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Host port is configured", false).
					SetType(testhelper.HostPortType).
					AddField(testhelper.PortNumber, strconv.Itoa(int(aPort.HostPort))))
				hostPortFound = true
			}
		}

		if !hostPortFound {
			check.LogInfo("Host port not configured in Container %q.", cut)
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Host port is not configured", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodHostNetwork checks that a pod does not enable host networking.
//
// It examines the pod specification in the provided test environment and verifies
// that the hostNetwork flag is not set to true. If the flag is enabled, the function
// logs an error and records a failure result in the check report. Successful pods
// are logged with an informational message. The function takes a checksdb.Check
// object for reporting and a provider.TestEnvironment containing the pod data.
func testPodHostNetwork(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		check.LogInfo("Testing Pod %q", put)
		if put.Spec.HostNetwork {
			check.LogError("Host network is set to true in Pod %q.", put)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Host network is set to true", false))
		} else {
			check.LogInfo("Host network is set to false in Pod %q.", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Host network is not set to true", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodHostPath verifies that the pod hostpath parameter is not set to true.
//
// It receives a checksdb.Check and a provider.TestEnvironment, logs information about the check, inspects the pod specification for any hostPath usage, records findings in a PodReportObject, and sets the result accordingly. If hostPath is enabled, it logs an error and marks the test as failed; otherwise it reports success.
func testPodHostPath(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		check.LogInfo("Testing Pod %q", put)
		podIsCompliant := true
		for idx := range put.Spec.Volumes {
			vol := &put.Spec.Volumes[idx]
			if vol.HostPath != nil && vol.HostPath.Path != "" {
				check.LogError("Hostpath path: %q is set in Pod %q.", vol.HostPath.Path, put)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Hostpath path is set", false).
					SetType(testhelper.HostPathType).
					AddField(testhelper.Path, vol.HostPath.Path))
				podIsCompliant = false
			}
		}
		if podIsCompliant {
			check.LogError("Hostpath path not set in Pod %q.", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Hostpath path is not set", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodHostIPC verifies that a Pod’s HostIPC field is not enabled.
//
// It receives a Check object and a TestEnvironment, iterates over all Pods
// in the environment, and records a report for each Pod where the HostIPC
// flag is true. The function logs informational messages during processing
// and sets the result status on the check based on whether any disallowed
// Pods were found. No value is returned; the outcome is communicated via
// side‑effects on the supplied Check object.
func testPodHostIPC(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		check.LogInfo("Testing Pod %q", put)
		if put.Spec.HostIPC {
			check.LogError("HostIpc is set in Pod %q.", put)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "HostIpc is set to true", false))
		} else {
			check.LogInfo("HostIpc not set in Pod %q.", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "HostIpc is not set to true", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodHostPID verifies that a pod does not set the hostPid parameter to true.
//
// It accepts a checksdb.Check and a provider.TestEnvironment, creates report objects for each pod,
// logs information about the test execution, and records the result of the check.
// The function iterates over the pods in the environment, ensuring that none have hostPid enabled,
// and updates the check with success or failure accordingly.
func testPodHostPID(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		check.LogInfo("Testing Pod %q", put)
		if put.Spec.HostPID {
			check.LogError("HostPid is set in Pod %q.", put)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "HostPid is set to true", false))
		} else {
			check.LogInfo("HostPid not set in Pod %q.", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "HostPid is not set to true", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testNamespace tests namespace validity and custom resource alignment.
//
// It checks that all namespaces used in the environment do not have
// disallowed prefixes and that any Custom Resources (CRs) are defined
// within a valid namespace. The function logs progress, records any
// failures, and sets the test result accordingly. It operates on a
// *checksdb.Check object to record outcomes and uses a *provider.TestEnvironment
// for context information about the test run.
func testNamespace(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, namespace := range env.Namespaces {
		check.LogInfo("Testing namespace %q", namespace)
		namespaceCompliant := true
		for _, invalidPrefix := range invalidNamespacePrefixes {
			if strings.HasPrefix(namespace, invalidPrefix) {
				check.LogError("Namespace %q has invalid prefix %q", namespace, invalidPrefix)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNamespacedReportObject("Namespace has invalid prefix", testhelper.Namespace, false, namespace))
				namespaceCompliant = false
				break // Break out of the loop if we find an invalid prefix
			}
		}
		if namespaceCompliant {
			check.LogInfo("Namespace %q has valid prefix", namespace)
			compliantObjects = append(compliantObjects, testhelper.NewNamespacedReportObject("Namespace has valid prefix", testhelper.Namespace, true, namespace))
		}
	}
	if failedNamespacesNum := len(nonCompliantObjects); failedNamespacesNum > 0 {
		check.SetResult(compliantObjects, nonCompliantObjects)
	}

	invalidCrs, err := namespace.TestCrsNamespaces(env.Crds, env.Namespaces, check.GetLogger())
	if err != nil {
		check.LogError("Error while testing CRs namespaces, err=%v", err)
		return
	}

	invalidCrsNum := namespace.GetInvalidCRsNum(invalidCrs, check.GetLogger())
	if invalidCrsNum > 0 {
		nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject("CRs are not in the configured namespaces", testhelper.Namespace, false))
	} else {
		compliantObjects = append(compliantObjects, testhelper.NewReportObject("CRs are in the configured namespaces", testhelper.Namespace, true))
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodServiceAccount verifies that the pod utilizes a valid service account.
//
// It inspects each pod in the test environment, checks the service account
// assigned to the pod against expected values, and records the result.
// The function logs informational messages, appends report objects,
// and sets the overall test result based on whether any pods use an
// invalid or missing service account.
func testPodServiceAccount(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		check.LogInfo("Testing Pod %q", put)
		if put.Spec.ServiceAccountName == defaultServiceAccount {
			check.LogError("Pod %q does not have a valid service account name (uses the default service account instead).", put)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod does not have a valid service account name", false))
		} else {
			check.LogInfo("Pod %q has a valid service account name", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has a service account name", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodRoleBindings verifies that a pod uses a valid role binding that does not cross non‑CNF namespaces.
//
// It examines the role bindings associated with the pod in the given test environment,
// ensuring they reference only allowed namespaces and do not grant permissions beyond
// the CNF scope. The function logs detailed information about each check, records
// any violations, and updates the checks database record accordingly. No return value is provided.
func testPodRoleBindings(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, put := range env.Pods {
		check.LogInfo("Testing Pod %q", put)
		podIsCompliant := true
		if put.Spec.ServiceAccountName == defaultServiceAccount {
			check.LogError("Pod %q has an empty or default serviceAccountName", put)
			// Add the pod to the non-compliant list
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewPodReportObject(put.Namespace, put.Name,
					"The serviceAccountName is either empty or default", false))
			podIsCompliant = false
		} else {
			check.LogInfo("Pod %q has a serviceAccountName %q, checking role bindings.", put, put.Spec.ServiceAccountName)
			// Loop through the rolebindings and check if they are from another namespace
			for rbIndex := range env.RoleBindings {
				// Short circuit if the role binding and the pod are in the same namespace.
				if env.RoleBindings[rbIndex].Namespace == put.Namespace {
					check.LogInfo("Pod %q and the role binding are in the same namespace", put)
					continue
				}
				// If we make it to this point, the role binding and the pod are in different namespaces.
				// We must check if the pod's service account is in the role binding's subjects.
				found := false
				for _, subject := range env.RoleBindings[rbIndex].Subjects {
					// If the subject is a service account and the service account is in the same namespace as one of the CNF's namespaces, then continue, this is allowed
					if subject.Kind == rbacv1.ServiceAccountKind &&
						subject.Namespace == put.Namespace &&
						subject.Name == put.Spec.ServiceAccountName &&
						stringhelper.StringInSlice[string](env.Namespaces, env.RoleBindings[rbIndex].Namespace, false) {
						continue
					}

					// Finally, if the subject is a service account and the service account is in the same namespace as the pod, then we have a failure
					if subject.Kind == rbacv1.ServiceAccountKind &&
						subject.Namespace == put.Namespace &&
						subject.Name == put.Spec.ServiceAccountName {
						check.LogError("Pod %q has the following role bindings that do not live in one of the CNF namespaces: %q", put, env.RoleBindings[rbIndex].Name)

						// Add the pod to the non-compliant list
						nonCompliantObjects = append(nonCompliantObjects,
							testhelper.NewPodReportObject(put.Namespace, put.Name,
								"The role bindings used by this pod do not live in one of the CNF namespaces", false).
								AddField(testhelper.RoleBindingName, env.RoleBindings[rbIndex].Name).
								AddField(testhelper.RoleBindingNamespace, env.RoleBindings[rbIndex].Namespace).
								AddField(testhelper.ServiceAccountName, put.Spec.ServiceAccountName).
								SetType(testhelper.PodRoleBinding))
						found = true
						podIsCompliant = false
						break
					}
				}
				// Break of out the loop if we found a role binding that is out of namespace
				if found {
					break
				}
			}
		}
		// Add pod to the compliant object list
		if podIsCompliant {
			check.LogInfo("All the role bindings used by Pod %q (applied by the service accounts) live in one of the CNF namespaces", put)
			compliantObjects = append(compliantObjects,
				testhelper.NewPodReportObject(put.Namespace, put.Name, "All the role bindings used by this pod (applied by the service accounts) live in one of the CNF namespaces", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodClusterRoleBindings verifies that a pod does not use a cluster role binding.
//
// testPodClusterRoleBindings checks whether the specified pod has any cluster‑wide role bindings.
// It logs relevant information, determines if the pod is using such bindings, and records
// the result in a report object. The function takes a check context and a test environment,
// performs the verification logic, and updates the pod's report accordingly.
func testPodClusterRoleBindings(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, put := range env.Pods {
		check.LogInfo("Testing Pod %q", put)
		result, roleRefName, err := put.IsUsingClusterRoleBinding(env.ClusterRoleBindings, check.GetLogger())
		if err != nil {
			check.LogError("Failed to determine if Pod %q is using a cluster role binding, err=%v", put, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, fmt.Sprintf("failed to determine if pod is using a cluster role binding: %v", err), false).
				AddField(testhelper.ClusterRoleName, roleRefName))
			continue
		}

		topOwners, err := put.GetTopOwner()
		if err != nil {
			check.LogError("Could not get top owners of Pod %q, err=%v", put, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, fmt.Sprintf("Error getting top owners of this pod, err=%s", err), false).
				AddField(testhelper.ClusterRoleName, roleRefName))
			continue
		}

		csvNamespace, csvName, isOwnedByClusterWideOperator := ownedByClusterWideOperator(topOwners, env)
		// Pod is using a cluster role binding but is owned by a cluster wide operator, so it is ok
		if isOwnedByClusterWideOperator && result {
			check.LogInfo("Pod %q is using a cluster role binding but is owned by a cluster-wide operator (Csv %q, namespace %q)", put, csvName, csvNamespace)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is using a cluster role binding but owned by a cluster-wide operator", true))
			continue
		}
		if result {
			// Pod was found to be using a cluster role binding.  This is not allowed.
			// Flagging this pod as a failed pod.
			check.LogError("Pod %q is using a cluster role binding (roleRefName=%q)", put, roleRefName)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is using a cluster role binding", false).
				AddField(testhelper.ClusterRoleName, roleRefName))
			continue
		}
		check.LogInfo("Pod %q is not using a cluster role binding", put)
		compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is not using a cluster role binding", true))
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// isCSVAndClusterWide reports whether a CSV object in the given namespace and name was created by a cluster‑wide operator.
//
// It checks if the specified CustomResourceDefinition (CSV) exists in the supplied namespace and name,
// then determines if that CSV originates from a cluster‑wide operator by inspecting its install mode.
// The function returns true when the CSV is associated with a cluster‑wide operator; otherwise it returns false.
func isCSVAndClusterWide(aNamespace, name string, env *provider.TestEnvironment) bool {
	for _, op := range env.Operators {
		if op.Csv != nil &&
			op.Csv.Namespace == aNamespace &&
			op.Csv.Name == name &&
			(op.IsClusterWide || isInstallModeMultiNamespace(op.Csv.Spec.InstallModes)) {
			return true
		}
	}
	return false
}

// isInstallModeMultiNamespace reports whether any install mode indicates multi or all namespaces.
//
// It examines a slice of InstallMode values and returns true if at least one element
// specifies either MultiNamespace or AllNamespaces, otherwise it returns false. The function
// simply iterates over the provided slice and checks each mode's type field.
func isInstallModeMultiNamespace(installModes []v1alpha1.InstallMode) bool {
	for i := 0; i < len(installModes); i++ {
		if installModes[i].Type == v1alpha1.InstallModeTypeAllNamespaces {
			return true
		}
	}
	return false
}

// ownedByClusterWideOperator checks whether any of the provided top owners is a CSV installed by a cluster‑wide operator.
//
// It examines the map of podhelper.TopOwner values, looking for an entry that represents a ClusterServiceVersion
// installed as a cluster‑wide resource. If such an owner is found, the function returns its name and namespace along with true.
// If no matching owner exists, it returns false and empty strings for name and namespace.
func ownedByClusterWideOperator(topOwners map[string]podhelper.TopOwner, env *provider.TestEnvironment) (aNamespace, name string, found bool) {
	for _, owner := range topOwners {
		if isCSVAndClusterWide(owner.Namespace, owner.Name, env) {
			return owner.Namespace, owner.Name, true
		}
	}
	return "", "", false
}

// testAutomountServiceToken verifies whether pods use the default service account and if the service token is explicitly set or inherited.
//
// It analyzes all pods in the cluster, checking their spec for a custom service token setting.
// For each pod it records compliance status based on the presence of an explicit token or
// inheritance from the associated ServiceAccount. The function logs progress and errors,
// updates the check result, and appends detailed report objects for compliant and non‑compliant pods.
func testAutomountServiceToken(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		check.LogInfo("Testing Pod %q", put)
		if put.Spec.ServiceAccountName == defaultServiceAccount {
			check.LogError("Pod %q uses the default service account name.", put.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has been found with default service account name", false))
			break
		}

		// Evaluate the pod's automount service tokens and any attached service accounts
		client := clientsholder.GetClientsHolder()
		podPassed, newMsg := rbac.EvaluateAutomountTokens(client.K8sClient.CoreV1(), put)
		if !podPassed {
			check.LogError("%s", newMsg)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, newMsg, false))
		} else {
			check.LogInfo("Pod %q does not have automount service tokens set to true", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod does not have automount service tokens set to true", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testOneProcessPerContainer checks that each non-Istio proxy container runs a single process.
//
// It iterates over the containers in the test environment, skips Istio proxy containers,
// retrieves their PID information, and counts the number of processes running in each
// container's namespace. Containers with more than one process are recorded as non‑compliant,
// while those with exactly one process are marked compliant. The function updates the
// compliance check result accordingly, logging progress and any errors encountered during
// analysis.
func testOneProcessPerContainer(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, cut := range env.Containers {
		check.LogInfo("Testing Container %q", cut)
		// the Istio sidecar container "istio-proxy" launches two processes: "pilot-agent" and "envoy"
		if cut.IsIstioProxy() {
			check.LogInfo("Skipping \"istio-proxy\" container")
			continue
		}
		probePod := env.ProbePods[cut.NodeName]
		if probePod == nil {
			check.LogError("Debug pod not found for node %q", cut.NodeName)
			return
		}
		ocpContext := clientsholder.NewContext(probePod.Namespace, probePod.Name, probePod.Spec.Containers[0].Name)
		pid, err := crclient.GetPidFromContainer(cut, ocpContext)
		if err != nil {
			check.LogError("Could not get PID for Container %q, error: %v", cut, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, err.Error(), false))
			continue
		}

		nbProcesses, err := getNbOfProcessesInPidNamespace(ocpContext, pid, clientsholder.GetClientsHolder())
		if err != nil {
			check.LogError("Could not get number of processes for Container %q, error: %v", cut, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, err.Error(), false))
			continue
		}
		if nbProcesses > 1 {
			check.LogError("Container %q has more than one process running", cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has more than one process running", false))
		} else {
			check.LogInfo("Container %q has only one process running", cut)
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has only one process running", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testSYSNiceRealtimeCapability checks that every container running on a node
// with a realtime kernel has the SYS_NICE capability set.
//
// It receives a Check object and a TestEnvironment, scans all containers in the
// environment, and for each container on an RT-enabled node it verifies whether
// the SYS_NICE capability is present.  Containers are reported as compliant or
// non‑compliant using report objects.  The function logs progress and errors,
// updates the check result accordingly, and records any failures in the test
// output.
func testSYSNiceRealtimeCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	// Loop through all of the labeled containers and compare their security context capabilities and whether
	// or not the node's kernel is realtime enabled.
	for _, cut := range env.Containers {
		check.LogInfo("Testing Container %q", cut)
		n := env.Nodes[cut.NodeName]
		if !n.IsRTKernel() {
			check.LogInfo("Container is not running on a realtime kernel enabled node")
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container is not running on a realtime kernel enabled node", true))
			continue
		}
		if !isContainerCapabilitySet(cut.SecurityContext.Capabilities, "SYS_NICE") {
			check.LogError("Container %q has been found running on a realtime kernel enabled node without SYS_NICE capability.", cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container is running on a realtime kernel enabled node without SYS_NICE capability", false))
		} else {
			check.LogInfo("Container is running on a realtime kernel enabled node with the SYS_NICE capability")
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container is running on a realtime kernel enabled node with the SYS_NICE capability", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testSysPtraceCapability checks pod compliance with SYS_PTRACE capability requirements.
//
// It evaluates whether each pod has the process namespace shared and contains at least one container
// that allows the SYS_PTRACE capability. The function collects lists of compliant and non-compliant pods,
// logs relevant information, and sets the result of the compliance check accordingly. The function takes
// a *checksdb.Check object to record results and a *provider.TestEnvironment for environment context.
func testSysPtraceCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.GetShareProcessNamespacePods() {
		check.LogInfo("Testing Pod %q", put)
		sysPtraceEnabled := false
		for _, cut := range put.Containers {
			if cut.SecurityContext == nil ||
				cut.SecurityContext.Capabilities == nil ||
				len(cut.SecurityContext.Capabilities.Add) == 0 {
				continue
			}
			if stringhelper.StringInSlice(cut.SecurityContext.Capabilities.Add, "SYS_PTRACE", false) {
				check.LogInfo("Container %q defines the SYS_PTRACE capability", cut)
				sysPtraceEnabled = true
				break
			}
		}
		if !sysPtraceEnabled {
			check.LogError("Pod %q has process namespace sharing enabled but no container allowing the SYS_PTRACE capability.", put)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has process namespace sharing enabled but no container allowing the SYS_PTRACE capability", false))
		} else {
			check.LogInfo("Pod %q has process namespace sharing enabled and at least one container allowing the SYS_PTRACE capability", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has process namespace sharing enabled and at least one container allowing the SYS_PTRACE capability", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testNamespaceResourceQuota checks that each pod runs in a namespace with an applied ResourceQuota and sets the compliance check result accordingly.
//
// It iterates over pods discovered by the environment, classifying them into compliant or non‑compliant lists based on whether their namespace has an active ResourceQuota.
// For each pod it creates a report object indicating compliance status. After processing all pods, it records the overall result of the check using SetResult.
// The function logs informational messages during execution and reports errors if any occur while retrieving or evaluating resources.
func testNamespaceResourceQuota(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, put := range env.Pods {
		check.LogInfo("Testing Pod %q", put)
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
			check.LogError("Pod %q is running in a namespace that does not have a ResourceQuota applied.", put)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is running in a namespace that does not have a ResourceQuota applied", false))
		} else {
			check.LogInfo("Pod %q is running in a namespace that has a ResourceQuota applied.", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is running in a namespace that has a ResourceQuota applied", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

const (
	sshServicePortProtocol = "TCP"
)

// testNoSSHDaemonsAllowed checks whether any pod in the test environment is running an SSH daemon and records compliance results.
//
// It iterates over all pods, inspects their listening ports for the SSH service port,
// and classifies each pod as compliant or non‑compliant based on the presence of that port.
// The function logs progress, handles errors from port parsing, and updates the check result
// with lists of compliant and non‑compliant pod report objects.
func testNoSSHDaemonsAllowed(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, put := range env.Pods {
		check.LogInfo("Testing Pod %q", put)
		cut := put.Containers[0]

		// 1. Find SSH port
		port, err := netutil.GetSSHDaemonPort(cut)
		if err != nil {
			check.LogError("Could not get ssh daemon port on %q, err: %v", cut, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Failed to get the ssh port for pod", false))
			continue
		}

		if port == "" {
			check.LogInfo("Pod %q is not running an SSH daemon", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is not running an SSH daemon", true))
			continue
		}

		sshServicePortNumber, err := strconv.ParseInt(port, 10, 32)
		if err != nil {
			check.LogError("Could not convert port %q from string to integer on Container %q", port, cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Failed to get the listening ports for pod", false))
			continue
		}

		// 2. Check if SSH port is listening
		sshPortInfo := netutil.PortInfo{PortNumber: int32(sshServicePortNumber), Protocol: sshServicePortProtocol}
		listeningPorts, err := netutil.GetListeningPorts(cut)
		if err != nil {
			check.LogError("Failed to get the listening ports for Pod %q, err: %v", put, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Failed to get the listening ports for pod", false))
			continue
		}

		if _, ok := listeningPorts[sshPortInfo]; ok {
			check.LogError("Pod %q is running an SSH daemon", put)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is running an SSH daemon", false))
		} else {
			check.LogInfo("Pod %q is not running an SSH daemon", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is not running an SSH daemon", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodRequests checks whether every container in the tested pods has resource requests set and records a compliance report.
//
// It examines each pod returned by the environment, determines if containers have requested resources,
// and classifies them as compliant or non‑compliant. The function logs information about the process,
// builds report objects for each container, and finally sets the result of the associated check
// based on the collected data. No values are returned; instead the check's state is updated through SetResult.
func testPodRequests(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	// Loop through the containers, looking for containers that are missing requests.
	// These need to be defined in order to pass.
	for _, cut := range env.Containers {
		check.LogInfo("Testing Container %q", cut)
		if !resources.HasRequestsSet(cut, check.GetLogger()) {
			check.LogError("Container %q is missing resource requests", cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container is missing resource requests", false))
		} else {
			check.LogInfo("Container %q has resource requests", cut)
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has resource requests", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// test1337UIDs checks whether all pods in a test environment run with UID 1337 and records compliance results.
//
// It iterates over the list of pods, evaluates each pod’s securityContext RunAsUser value,
// and classifies them as compliant or non‑compliant.
// For every pod it creates a report object indicating success or failure
// and aggregates these reports into the overall check result.
func test1337UIDs(check *checksdb.Check, env *provider.TestEnvironment) {
	// Note this test is only ran as part of the 'extended' test suite.
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	const leetNum = 1337
	for _, put := range env.Pods {
		check.LogInfo("Testing Pod %q", put)
		if put.IsRunAsUserID(leetNum) {
			check.LogError("Pod %q is using securityContext RunAsUser 1337", put)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is using securityContext RunAsUser 1337", false))
		} else {
			check.LogInfo("Pod %q is not using securityContext RunAsUser 1337", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is not using securityContext RunAsUser 1337", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testContainerSCC categorizes the containers under test into several categories of increasing privileges based on their SCC.
//
// It analyzes the list of compliant and non‑compliant objects, logs relevant information,
// and sets the compliance check result accordingly. The function receives a checksdb.Check
// pointer to record the outcome and a provider.TestEnvironment pointer for accessing
// test context and utilities. No return value is produced; results are stored in the
// provided Check object.
func testContainerSCC(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	highLevelCat := securitycontextcontainer.CategoryID1
	for _, pod := range env.Pods {
		check.LogInfo("Testing Pod %q", pod)
		listCategory := securitycontextcontainer.CheckPod(pod)
		for _, cat := range listCategory {
			if cat.Category > securitycontextcontainer.CategoryID1NoUID0 {
				check.LogError("Category %q is NOT category 1 or category NoUID0", cat)
				aContainerOut := testhelper.NewContainerReportObject(cat.NameSpace, cat.Podname, cat.Containername, "container category is NOT category 1 or category NoUID0", false).
					SetType(testhelper.ContainerCategory).
					AddField(testhelper.Category, cat.Category.String())
				nonCompliantObjects = append(nonCompliantObjects, aContainerOut)
			} else {
				check.LogInfo("Category %q is category 1 or category NoUID0", cat)
				aContainerOut := testhelper.NewContainerReportObject(cat.NameSpace, cat.Podname, cat.Containername, "container category is category 1 or category NoUID0", true).
					SetType(testhelper.ContainerCategory).
					AddField(testhelper.Category, cat.Category.String())
				compliantObjects = append(compliantObjects, aContainerOut)
			}
			if cat.Category > highLevelCat {
				highLevelCat = cat.Category
			}
		}
	}
	aCNFOut := testhelper.NewReportObject("Overall CNF category", testhelper.CnfType, false).AddField(testhelper.Category, highLevelCat.String())
	compliantObjects = append(compliantObjects, aCNFOut)
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testNodePort checks that each service of a given type uses NodePort and records compliance results.
//
// It examines the list of services, separating those that expose a node port from those that do not. For each group it creates a report object containing the relevant service details, logs the outcome, and updates the check result accordingly. The function operates on a Check instance and a TestEnvironment to access cluster information and reporting utilities.
func testNodePort(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, s := range env.Services {
		check.LogInfo("Testing %q", services.ToString(s))

		if s.Spec.Type == nodePort {
			check.LogError("Service %q (ns %q) type is nodePort", s.Name, s.Namespace)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject("Service is type NodePort", testhelper.ServiceType, false).
				AddField(testhelper.Namespace, s.Namespace).
				AddField(testhelper.ServiceName, s.Name).
				AddField(testhelper.ServiceMode, string(s.Spec.Type)))
		} else {
			check.LogInfo("Service %q (ns %q) type is not nodePort (type=%q)", s.Name, s.Namespace, s.Spec.Type)
			compliantObjects = append(compliantObjects, testhelper.NewReportObject("Service is not type NodePort", testhelper.ServiceType, true).
				AddField(testhelper.Namespace, s.Namespace).
				AddField(testhelper.ServiceName, s.Name).
				AddField(testhelper.ServiceMode, string(s.Spec.Type)))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testCrdRoles checks that each role applies only to the CRDs under test.
//
// It retrieves the list of CRD resources and all RBAC rules,
// then filters out rules that do not match any of those resources.
// For each rule, it records compliant or non‑compliant objects in a report.
// The function updates the compliance check result based on the analysis
// and logs relevant information during processing.
func testCrdRoles(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	crdResources := rbac.GetCrdResources(env.Crds)
	for roleIndex := range env.Roles {
		if !stringhelper.StringInSlice[string](env.Namespaces, env.Roles[roleIndex].Namespace, false) {
			continue
		}

		allRules := rbac.GetAllRules(&env.Roles[roleIndex])

		matchingRules, nonMatchingRules := rbac.FilterRulesNonMatchingResources(allRules, crdResources)
		if len(matchingRules) == 0 {
			continue
		}
		for _, aRule := range matchingRules {
			check.LogInfo("Rule (resource-name=%q, resource-group=%q, verb=%q, role-name=%q) applies to CRDs under test",
				aRule.Resource.Name, aRule.Resource.Group, aRule.Verb, env.Roles[roleIndex].Name)
			compliantObjects = append(compliantObjects, testhelper.NewNamespacedReportObject("This applies to CRDs under test", testhelper.RoleRuleType, true, env.Roles[roleIndex].Namespace).
				AddField(testhelper.RoleName, env.Roles[roleIndex].Name).
				AddField(testhelper.Group, aRule.Resource.Group).
				AddField(testhelper.ResourceName, aRule.Resource.Name).
				AddField(testhelper.Verb, aRule.Verb))
		}
		for _, aRule := range nonMatchingRules {
			check.LogInfo("Rule (resource-name=%q, resource-group=%q, verb=%q, role-name=%q) does not apply to CRDs under test",
				aRule.Resource.Name, aRule.Resource.Group, aRule.Verb, env.Roles[roleIndex].Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNamespacedReportObject("This rule does not apply to CRDs under test", testhelper.RoleRuleType, false, env.Roles[roleIndex].Namespace).
				AddField(testhelper.RoleName, env.Roles[roleIndex].Name).
				AddField(testhelper.Group, aRule.Resource.Group).
				AddField(testhelper.ResourceName, aRule.Resource.Name).
				AddField(testhelper.Verb, aRule.Verb))
		}

		if len(nonMatchingRules) == 0 {
			check.LogInfo("Role %q rules only apply to CRDs under test", env.Roles[roleIndex].Name)
			compliantObjects = append(compliantObjects, testhelper.NewNamespacedNamedReportObject("This role's rules only apply to CRDs under test",
				testhelper.RoleType, true, env.Roles[roleIndex].Namespace, env.Roles[roleIndex].Name))
		} else {
			check.LogError("Role %q rules apply to a mix of CRDs under test and others.", env.Roles[roleIndex].Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNamespacedNamedReportObject("This role's rules apply to a mix of CRDs under test and others. See non compliant role rule objects.",
				testhelper.RoleType, false, env.Roles[roleIndex].Namespace, env.Roles[roleIndex].Name))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}
