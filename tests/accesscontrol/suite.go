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

// LoadChecks Initialises all access control checks for the test suite
//
// This function registers a group of security checks under the access‑control
// key, attaching pre‑execution logic and a series of check functions that
// validate container capabilities, pod configurations, namespace policies, and
// more. Each check is created with identifiers derived from test metadata, may
// be skipped based on environmental conditions, and logs its progress through a
// shared logger.
//
//nolint:funlen
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

// isContainerCapabilitySet checks if a capability is explicitly added to a container
//
// The function receives a pointer to a capabilities structure and a capability
// name. It returns true when the Add list contains either the specified
// capability or the special ALL value, indicating that the capability has been
// granted. If the capabilities object is nil or its Add list is empty, it
// returns false.
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

// checkForbiddenCapability determines if containers avoid a specific capability
//
// The function iterates over each container, checking whether the specified
// capability is present in its security context. Containers lacking the
// capability are recorded as compliant; those with it are flagged
// non‑compliant and an error logged. The results are returned as two slices
// of report objects for further processing.
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

// testSysAdminCapability Checks containers for the SYS_ADMIN capability
//
// This routine examines each container in the test environment, looking for the
// SYS_ADMIN capability in its security context. Containers lacking this
// capability are recorded as compliant; those that include it are flagged
// non‑compliant with an error log entry. The results are then stored back
// into the check object.
func testSysAdminCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(env.Containers, "SYS_ADMIN", check.GetLogger())
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testNetAdminCapability Verifies that containers do not use the NET_ADMIN capability
//
// The function queries all test containers for the presence of the NET_ADMIN
// capability using a helper routine, collecting compliant and non‑compliant
// reports. It then records these results in the provided check object. This
// ensures that network administration privileges are not granted to container
// processes.
func testNetAdminCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(env.Containers, "NET_ADMIN", check.GetLogger())
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testNetRawCapability Validates that containers do not use the NET_RAW capability
//
// The function examines all container security contexts in the test environment
// and reports any instance where the NET_RAW capability is granted. It records
// compliant and non‑compliant findings, attaching relevant details to each
// report object. Finally, it stores the results in the provided check for later
// aggregation.
func testNetRawCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(env.Containers, "NET_RAW", check.GetLogger())
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testIpcLockCapability Verifies containers lack the IPC_LOCK capability
//
// This function inspects each container in the test environment, checking
// whether the IPC_LOCK capability is present in its security context.
// Containers without the capability are marked compliant; those with it are
// flagged non‑compliant and reported accordingly. The results are stored back
// into the check object for later aggregation.
func testIpcLockCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(env.Containers, "IPC_LOCK", check.GetLogger())
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testBpfCapability Verifies that containers do not use the BPF kernel capability
//
// The function iterates over all test containers, checking whether the "BPF"
// capability is present in their security context. It records compliant
// containers where the capability is absent and non‑compliant ones where it
// appears, attaching appropriate log messages. Finally, it stores the results
// within the provided check object for reporting.
func testBpfCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(env.Containers, "BPF", check.GetLogger())
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testSecConRunAsNonRoot checks that pods do not run containers as root
//
// The routine iterates over all test pods, determines which containers are
// configured to run as root, and records compliance results. For each pod it
// logs a message, then calls a helper to retrieve non‑compliant containers.
// If none exist the pod is marked compliant; otherwise each offending container
// is logged with an error and added to the non‑compliant list. Finally, the
// check’s result is set with both compliant and non‑compliant report
// objects.
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

// testSecConPrivilegeEscalation Verifies that containers do not allow privilege escalation
//
// The function iterates over each container in the test environment, checking
// if the SecurityContext's AllowPrivilegeEscalation flag is explicitly set to
// true. Containers with this setting are marked non‑compliant and logged as
// errors; those without the flag or with it false are considered compliant.
// Results are collected into report objects and stored via SetResult.
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

// testSecConReadOnlyFilesystem verifies that each container mounts a read‑only root filesystem
//
// The function iterates over every pod and its containers in the test
// environment, checking if the container’s root filesystem is set to
// read‑only using a helper method. Containers that satisfy the requirement
// are recorded as compliant; those that do not are logged as errors and marked
// non‑compliant. Finally, the results are aggregated into report objects and
// passed back via the check’s result setter.
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

// testContainerHostPort Verifies that containers do not expose host ports
//
// The function iterates over all containers in the test environment, checking
// each declared port for a non‑zero HostPort value. If such a port is found,
// it records a non‑compliant report object with details of the offending
// container and port number; otherwise it logs compliance and records a
// compliant object. Finally, it sets the check result with the collected
// reports.
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

// testPodHostNetwork Checks whether pods enable host networking
//
// The routine iterates over all pod objects in the test environment, logging
// each one. If a pod’s HostNetwork flag is true it records a non‑compliant
// report object and logs an error; otherwise it records a compliant report
// object and logs success. Finally, it stores both lists of results on the
// check instance.
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

// testPodHostPath Verifies that pods do not use host path volumes
//
// The function iterates over all pods in the test environment, checking each
// volume for a non-empty HostPath field. If such a path is found, the pod is
// marked non‑compliant and recorded with details; otherwise it is considered
// compliant. Results are logged and reported via the check object.
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

// testPodHostIPC Ensures Pod HostIPC is disabled
//
// The function iterates over all pods in the test environment, logging each
// pod’s name. For every pod it checks whether the HostIPC flag is true; if
// so, it records a non‑compliant report object and logs an error, otherwise
// it records a compliant report object. Finally, it sets the check result with
// both lists of objects.
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

// testPodHostPID Checks that no pod uses the host PID namespace
//
// The function iterates over all pods in the test environment, logging each
// one’s status. If a pod has HostPID enabled it records a non‑compliant
// report object and logs an error; otherwise it records a compliant object and
// logs informational text. Finally, it sets the check result with the lists of
// compliant and non‑compliant objects.
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

// testNamespace Checks namespace names for disallowed prefixes and validates CR placement
//
// The function iterates over all namespaces in the test environment, logging
// each one. For every namespace it verifies that none of the predefined invalid
// prefixes are present; if an invalid prefix is found a non‑compliant report
// object is created. After checking prefixes, it calls helper routines to
// examine whether custom resources exist in unauthorized namespaces and records
// any failures as non‑compliant objects. Finally, results for both compliant
// and non‑compliant cases are set on the check.
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

// testPodServiceAccount checks that pods use non‑default service accounts
//
// The function iterates over all pods in the test environment, logging each
// pod’s name. For every pod it verifies whether the ServiceAccountName equals
// the default account; if so, it records a non‑compliant report object and
// logs an error, otherwise it records a compliant report object and logs
// success. Finally, it sets the check result with the collected compliant and
// non‑compliant objects.
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

// testPodRoleBindings Checks pod role bindings against CNF namespace rules
//
// The routine iterates over each pod in the test environment, verifying that
// its service account is not default and that any role binding it relies on
// resides within an approved CNF namespace. If a pod references a role binding
// outside these namespaces, it is marked non‑compliant and detailed
// information about the offending binding is recorded. Pods passing all checks
// are logged as compliant, and results are reported back to the test framework.
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

// testPodClusterRoleBindings Verifies that pods do not use cluster role bindings
//
// The function iterates over all pods in the test environment, checking each
// for usage of a cluster role binding. If a pod uses one but is owned by a
// cluster‑wide operator, it is considered compliant; otherwise any usage
// flags the pod as non‑compliant. Results are recorded in report objects and
// returned via the check result.
//
//nolint:dupl
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

// isCSVAndClusterWide determines if a CSV belongs to a cluster‑wide operator
//
// The function inspects all operators in the test environment, checking whether
// any have a CSV matching the given namespace and name. If the CSV is found and
// its operator is marked as cluster wide or supports all namespaces, it returns
// true; otherwise it returns false.
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

// isInstallModeMultiNamespace determines whether a CSV install mode includes all namespaces
//
// The function iterates over the provided slice of install modes and returns
// true if any entry indicates an all‑namespace deployment. If none match, it
// returns false. This result is used to identify cluster‑wide operators in
// subsequent logic.
func isInstallModeMultiNamespace(installModes []v1alpha1.InstallMode) bool {
	for i := 0; i < len(installModes); i++ {
		if installModes[i].Type == v1alpha1.InstallModeTypeAllNamespaces {
			return true
		}
	}
	return false
}

// ownedByClusterWideOperator Determines if any top owner is a cluster‑wide CSV
//
// The function examines each top owner of an object, checking whether the owner
// is a ClusterServiceVersion that is installed by a cluster‑wide operator. It
// returns the namespace and name of the matching CSV along with a boolean flag
// indicating a match. If no such owner exists, empty strings and false are
// returned.
func ownedByClusterWideOperator(topOwners map[string]podhelper.TopOwner, env *provider.TestEnvironment) (aNamespace, name string, found bool) {
	for _, owner := range topOwners {
		if isCSVAndClusterWide(owner.Namespace, owner.Name, env) {
			return owner.Namespace, owner.Name, true
		}
	}
	return "", "", false
}

// testAutomountServiceToken Verifies pod service account usage and automount token settings
//
// The function iterates over all pods in the test environment, checking whether
// a pod uses the default service account name and evaluating its
// automountServiceAccountToken configuration via an external helper. It records
// non-compliant objects when defaults are used or tokens are set to true, and
// compliant objects otherwise. Results are stored in the check’s report for
// later reporting.
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

// testOneProcessPerContainer verifies that each container runs only a single process
//
// The function iterates over all containers in the test environment, excluding
// Istio proxy sidecars. For each container it obtains the main PID via the
// probe pod and counts processes in its namespace. Containers with more than
// one process are flagged as non‑compliant; otherwise they are marked
// compliant. The results are aggregated into report objects and set on the
// check.
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

// testSYSNiceRealtimeCapability Verifies SYS_NICE capability on containers in realtime kernel nodes
//
// The function iterates over all test containers, checks if their node uses a
// realtime kernel, and then determines whether the container has the SYS_NICE
// capability set. Containers running on non‑realtime nodes are automatically
// considered compliant. Results are collected into compliant and
// non‑compliant lists that are reported back to the test framework.
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

// testSysPtraceCapability verifies pods with shared process namespaces contain a container granting SYS_PTRACE
//
// The function iterates over all pods that enable shared process namespaces,
// examining each container’s security context for the SYS_PTRACE capability.
// If at least one container has this capability it records the pod as
// compliant; otherwise it logs an error and marks it non‑compliant. Finally,
// it sets the check result with lists of compliant and non‑compliant report
// objects.
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

// testNamespaceResourceQuota evaluates pod placement against namespace resource quotas
//
// The function iterates through all pods in the test environment, checking
// whether each pod's namespace has an associated ResourceQuota object. If a
// matching quota is found, the pod is marked compliant; otherwise it is
// recorded as non‑compliant and logged with an error message. After
// processing all pods, the compliance results are stored via the check’s
// SetResult method.
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

// testNoSSHDaemonsAllowed Verifies pods do not run SSH daemons
//
// The function iterates over each pod in the test environment, attempting to
// locate an SSH listening port within its first container. If a port is found,
// it checks whether that port is actively listening; presence indicates a
// running SSH daemon and marks the pod non‑compliant. Pods without an SSH
// port or with no active listener are considered compliant. Results are
// aggregated into report objects and set as the check outcome.
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

// testPodRequests Verifies that every container has defined CPU and memory requests
//
// The function iterates over all containers in the test environment, checking
// whether each one specifies resource requests using a helper routine.
// Containers lacking any request or with zero CPU or memory values are logged
// as errors and collected into a non‑compliant list; those that pass are
// recorded as compliant. Finally, it records both lists as the result of the
// compliance check.
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

// test1337UIDs Checks whether pods run with UID 1337
//
// The function iterates over all pods in the test environment, logging each
// pod’s status. It records a non‑compliant report object for any pod whose
// securityContext RunAsUser is set to 1337 and a compliant one otherwise.
// Finally, it sets the check result using these lists.
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

// testContainerSCC Determines container privilege levels based on SCC analysis
//
// The function iterates over all pods in the test environment, applying a
// security context check to each pod's containers. Containers are classified
// into categories, with only those in the lowest privileged category considered
// compliant. It records both compliant and non‑compliant containers, tracks
// the highest privilege level found, and reports an overall CNF category before
// setting the test result.
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

// testNodePort Checks services for disallowed nodePort usage
//
// The function iterates over all services in the test environment, logging each
// one. If a service is of type NodePort, it records a non‑compliant report
// object and logs an error; otherwise it records a compliant report object.
// Finally, it sets the check result with the two lists.
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

// testCrdRoles Evaluates whether role rules target only CRDs under test
//
// The function iterates over all roles in the environment, filtering by
// namespace, then extracts each role's API rules. It compares these rules
// against the list of CRD resources under test to separate matching and
// non‑matching rules. For each rule it records a report object indicating
// compliance, and if any role contains mixed rule types it logs an error and
// marks the role as non‑compliant.
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
