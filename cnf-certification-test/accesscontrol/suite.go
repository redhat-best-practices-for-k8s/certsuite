// Copyright (C) 2020-2023 Red Hat, Inc.
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
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/namespace"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/rbac"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/resources"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/securitycontextcontainer"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/netutil"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/services"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/internal/crclient"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/checksdb"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/stringhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
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
)

var (
	env provider.TestEnvironment

	beforeEachFn = func(check *checksdb.Check) error {
		check.LogInfo("Check %s: getting test environment.", check.ID)
		env = provider.GetTestEnvironment()
		return nil
	}
)

//nolint:funlen
func LoadChecks() {
	log.Debug("Loading %s checks", common.AccessControlTestKey)

	checksGroup := checksdb.NewChecksGroup(common.AccessControlTestKey).
		WithBeforeEachFn(beforeEachFn)

	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSecContextIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainerSCC(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSysAdminIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testSysAdminCapability(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNetAdminIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNetAdminCapability(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNetRawIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNetRawCapability(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestIpcLockIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testIpcLockCapability(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestBpfIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testBpfCapability(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSecConNonRootUserIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testSecConRootUser(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSecConPrivilegeEscalation)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testSecConPrivilegeEscalation(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestContainerHostPort)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainerHostPort(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodHostNetwork)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodHostNetwork(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodHostPath)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodHostPath(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodHostIPC)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodHostIPC(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodHostPID)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodHostPID(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNamespaceBestPracticesIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoNamespacesSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNamespace(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodServiceAccountBestPracticesIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodServiceAccount(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodRoleBindingsBestPracticesIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodRoleBindings(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodClusterRoleBindingsBestPracticesIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodClusterRoleBindings(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodAutomountServiceAccountIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testAutomountServiceToken(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestOneProcessPerContainerIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOneProcessPerContainer(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSYSNiceRealtimeCapabilityIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testSYSNiceRealtimeCapability(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSysPtraceCapabilityIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetSharedProcessNamespacePodsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testSysPtraceCapability(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNamespaceResourceQuotaIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNamespaceResourceQuota(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNoSSHDaemonsAllowedIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNoSSHDaemonsAllowed(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodRequestsAndLimitsIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodRequestsAndLimits(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.Test1337UIDIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			test1337UIDs(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestServicesDoNotUseNodeportsIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoServicesUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNodePort(c, &env)
			return nil
		}))

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestCrdRoleIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoCrdsUnderTestSkipFn(&env), testhelper.GetNoNamespacesSkipFn(&env), testhelper.GetNoRolesSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testCrdRoles(c, &env)
			return nil
		}))
}

func checkForbiddenCapability(check *checksdb.Check, containers []*provider.Container, capability string) (compliantObjects, nonCompliantObjects []*testhelper.ReportObject) {
	for _, cut := range containers {
		compliant := true

		switch {
		case cut.SecurityContext == nil:
		case cut.SecurityContext.Capabilities == nil:
		case strings.Contains(cut.SecurityContext.Capabilities.String(), capability):
			compliant = false
		}

		if compliant {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "No forbidden capability "+capability+" detected in container", true))
		} else {
			check.LogDebug("Non compliant %s capability detected in container %s. All container caps: %s", capability, cut.String(), cut.SecurityContext.Capabilities.String())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Non compliant capability "+capability+" in container", false).AddField(testhelper.SCCCapability, capability))
		}
	}
	return compliantObjects, nonCompliantObjects
}

func testSysAdminCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(check, env.Containers, "SYS_ADMIN")
	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testNetAdminCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(check, env.Containers, "NET_ADMIN")
	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testNetRawCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(check, env.Containers, "NET_RAW")
	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testIpcLockCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(check, env.Containers, "IPC_LOCK")
	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testBpfCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(check, env.Containers, "BPF")
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testSecConRootUser verifies that the container is not running as root
func testSecConRootUser(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		if put.IsRunAsUserID(0) {
			check.LogDebug("Non compliant run as Root User detected (RunAsUser uid=0) in pod %s", put.Namespace+"."+put.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Root User detected (RunAsUser uid=0)", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Root User not detected (RunAsUser uid=0)", true))
		}

		for idx := range put.Spec.Containers {
			cut := &(put.Spec.Containers[idx])
			// Check the container level RunAsUser parameter
			if cut.SecurityContext != nil && cut.SecurityContext.RunAsUser != nil {
				if *(cut.SecurityContext.RunAsUser) == 0 {
					nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(put.Namespace, put.Name, cut.Name, "Root User detected (RunAsUser uid=0)", false))
				} else {
					compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(put.Namespace, put.Name, cut.Name, "Root User not detected (RunAsUser uid=0)", true))
				}
			}
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testSecConPrivilegeEscalation verifies that the container is not allowed privilege escalation
func testSecConPrivilegeEscalation(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		privEscFound := false
		if cut.SecurityContext != nil && cut.SecurityContext.AllowPrivilegeEscalation != nil {
			if *(cut.SecurityContext.AllowPrivilegeEscalation) {
				check.LogDebug("AllowPrivilegeEscalation is set to true in container %s.", cut.Podname+"."+cut.Name)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "AllowPrivilegeEscalation is set to true", false))
				privEscFound = true
			}
		}

		if !privEscFound {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "AllowPrivilegeEscalation is not set to true", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testContainerHostPort tests that containers are not configured with host port privileges
func testContainerHostPort(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		hostPortFound := false
		for _, aPort := range cut.Ports {
			if aPort.HostPort != 0 {
				check.LogDebug("Host port %d is configured in container %s.", aPort.HostPort, cut.String())
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Host port is configured", false).
					SetType(testhelper.HostPortType).
					AddField(testhelper.PortNumber, strconv.Itoa(int(aPort.HostPort))))
				hostPortFound = true
			}
		}

		if !hostPortFound {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Host port is not configured", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodHostNetwork verifies that the pod hostNetwork parameter is not set to true
func testPodHostNetwork(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		if put.Spec.HostNetwork {
			check.LogDebug("Host network is set to true in pod %s.", put.Namespace+"."+put.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Host network is set to true", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Host network is not set to true", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodHostPath verifies that the pod hostpath parameter is not set to true
func testPodHostPath(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		podIsCompliant := true
		for idx := range put.Spec.Volumes {
			vol := &put.Spec.Volumes[idx]
			if vol.HostPath != nil && vol.HostPath.Path != "" {
				check.LogDebug("Hostpath path: %s is set in pod %s.", vol.HostPath.Path, put.Namespace+"."+put.Name)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Hostpath path is set", false).
					SetType(testhelper.HostPathType).
					AddField(testhelper.Path, vol.HostPath.Path))
				podIsCompliant = false
			}
		}
		if podIsCompliant {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Hostpath path is not set", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodHostIPC verifies that the pod hostIpc parameter is not set to true
func testPodHostIPC(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		if put.Spec.HostIPC {
			check.LogDebug("HostIpc is set in pod %s.", put.Namespace+"."+put.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "HostIpc is set to true", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "HostIpc is not set to true", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodHostPID verifies that the pod hostPid parameter is not set to true
func testPodHostPID(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		if put.Spec.HostPID {
			check.LogDebug("HostPid is set in pod %s.", put.Namespace+"."+put.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "HostPid is set to true", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "HostPid is not set to true", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// Tests namespaces for invalid prefixed and CRs are not defined in namespaces not under test with CRDs under test
func testNamespace(check *checksdb.Check, env *provider.TestEnvironment) {
	check.LogInfo("CNF resources' Namespaces should not have any of the following prefixes: %v", invalidNamespacePrefixes)
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, namespace := range env.Namespaces {
		namespaceCompliant := true
		log.Info("Checking namespace %s", namespace)
		for _, invalidPrefix := range invalidNamespacePrefixes {
			if strings.HasPrefix(namespace, invalidPrefix) {
				check.LogDebug("Namespace %s has invalid prefix %s", namespace, invalidPrefix)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNamespacedReportObject("Namespace has invalid prefix", testhelper.Namespace, false, namespace))
				namespaceCompliant = false
				break // Break out of the loop if we find an invalid prefix
			}
		}
		if namespaceCompliant {
			compliantObjects = append(compliantObjects, testhelper.NewNamespacedReportObject("Namespace has valid prefix", testhelper.Namespace, true, namespace))
		}
	}
	if failedNamespacesNum := len(nonCompliantObjects); failedNamespacesNum > 0 {
		check.SetResult(compliantObjects, nonCompliantObjects)
	}
	check.LogInfo("CNF pods should belong to any of the configured Namespaces: %v", env.Namespaces)
	check.LogInfo("CRs from autodiscovered CRDs should belong only to the configured Namespaces: %v", env.Namespaces)
	invalidCrs, err := namespace.TestCrsNamespaces(env.Crds, env.Namespaces)
	if err != nil {
		check.LogError("Error while testing CRs namespaces: %v", err)
		return
	}

	invalidCrsNum := namespace.GetInvalidCRsNum(invalidCrs, check.GetLoggger())
	if invalidCrsNum > 0 {
		nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject("CRs are not in the configured namespaces", testhelper.Namespace, false))
	} else {
		compliantObjects = append(compliantObjects, testhelper.NewReportObject("CRs are in the configured namespaces", testhelper.Namespace, true))
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodServiceAccount verifies that the pod utilizes a valid service account
func testPodServiceAccount(check *checksdb.Check, env *provider.TestEnvironment) {
	check.LogInfo("Tests that each pod utilizes a valid service account")
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		check.LogInfo("Testing service account for pod %s (ns: %s)", put.Name, put.Namespace)
		if put.Spec.ServiceAccountName == defaultServiceAccount {
			check.LogDebug("Pod %s (ns: %s) does not have a valid service account name.", put.Name, put.Namespace)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod does not have a valid service account name", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has a service account name", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodRoleBindings verifies that the pod utilizes a valid role binding that does not cross non-CNF namespaces
//
//nolint:funlen
func testPodRoleBindings(check *checksdb.Check, env *provider.TestEnvironment) {
	check.LogInfo("Should not have RoleBinding in other namespaces")
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, put := range env.Pods {
		podIsCompliant := true
		check.LogInfo("Testing role binding for pod: %s namespace: %s", put.Name, put.Namespace)
		if put.Pod.Spec.ServiceAccountName == defaultServiceAccount {
			log.Info("%s has an empty or default serviceAccountName, skipping.", put.String())
			// Add the pod to the non-compliant list
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewPodReportObject(put.Namespace, put.Name,
					"The serviceAccountName is either empty or default", false))
			podIsCompliant = false
		} else {
			log.Info("%s has a serviceAccountName: %s, checking role bindings.", put.String(), put.Spec.ServiceAccountName)
			// Loop through the rolebindings and check if they are from another namespace
			for rbIndex := range env.RoleBindings {
				// Short circuit if the role binding and the pod are in the same namespace.
				if env.RoleBindings[rbIndex].Namespace == put.Namespace {
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
						check.LogWarn("Pod: %s has the following role bindings that do not live in one of the CNF namespaces: %s", put, env.RoleBindings[rbIndex].Name)

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
			compliantObjects = append(compliantObjects,
				testhelper.NewPodReportObject(put.Namespace, put.Name, "All the role bindings used by this pod (applied by the service accounts) live in one of the CNF namespaces", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodClusterRoleBindings verifies that the pod does not use a cluster role binding
//
//nolint:dupl
func testPodClusterRoleBindings(check *checksdb.Check, env *provider.TestEnvironment) {
	check.LogInfo("Pods should not have ClusterRoleBindings")
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	log.Info("There were %d cluster role bindings found in the cluster.", len(env.ClusterRoleBindings))

	for _, put := range env.Pods {
		podIsCompliant := true
		check.LogInfo("Testing cluster role binding for pod: %s namespace: %s", put.Name, put.Namespace)
		result, roleRefName, err := put.IsUsingClusterRoleBinding(env.ClusterRoleBindings)
		if err != nil {
			log.Error("failed to determine if pod %s/%s is using a cluster role binding: %v", put.Namespace, put.Name, err)
			podIsCompliant = false
		}

		// Pod was found to be using a cluster role binding.  This is not allowed.
		// Flagging this pod as a failed pod.
		if result {
			check.LogWarn("%s is using a cluster role binding", put.String())
			podIsCompliant = false
		}

		if podIsCompliant {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is not using a cluster role binding", true))
		} else {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is using a cluster role binding", false).
				AddField(testhelper.ClusterRoleName, roleRefName))
		}

		topOwners, err := put.GetTopOwner()

		if err != nil {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, fmt.Sprintf("Error getting top owners of this pod, err=%s", err), false).
				AddField(testhelper.ClusterRoleName, roleRefName))
			continue
		}

		csvNamespace, csvName, isOwnedByClusterWideOperator := OwnedByClusterWideOperator(topOwners, env)
		// Pod is using a cluster role binding but is owned by a cluster wide operator, so it is ok
		if isOwnedByClusterWideOperator && result {
			log.Info("%s is using a cluster role binding but is owned by CSV namespace=%s, name=%s", put.String(), csvNamespace, csvName)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is using a cluster role binding but owned by a cluster-wide operator", true))
			continue
		}
		if result {
			// Pod was found to be using a cluster role binding.  This is not allowed.
			// Flagging this pod as a failed pod.
			log.Info("%s is using a cluster role binding", put.String())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is using a cluster role binding", false).
				AddField(testhelper.ClusterRoleName, roleRefName))
			continue
		}
		compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is not using a cluster role binding", true))
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// Returns true if object identified by namespace and name is a CSV created by a cluster-wide operator
func IsCSVAndClusterWide(aNamespace, name string, env *provider.TestEnvironment) bool {
	for _, op := range env.Operators {
		if op.Csv != nil &&
			op.Csv.Namespace == aNamespace &&
			op.Csv.Name == name &&
			(op.IsClusterWide || IsInstallModeMultiNamespace(op.Csv.Spec.InstallModes)) {
			return true
		}
	}
	return false
}

// return true if CSV install mode contains multi namespaces or all namespaces
func IsInstallModeMultiNamespace(installModes []v1alpha1.InstallMode) bool {
	for i := 0; i < len(installModes); i++ {
		if installModes[i].Type == v1alpha1.InstallModeTypeAllNamespaces {
			return true
		}
	}
	return false
}

// Return true if one of the passed topOwners is a CSV that is installed by a cluster-wide operator
func OwnedByClusterWideOperator(topOwners map[string]provider.TopOwner, env *provider.TestEnvironment) (aNamespace, name string, found bool) {
	for _, owner := range topOwners {
		if IsCSVAndClusterWide(owner.Namespace, owner.Name, env) {
			return owner.Namespace, owner.Name, true
		}
	}
	return "", "", false
}

func testAutomountServiceToken(check *checksdb.Check, env *provider.TestEnvironment) {
	check.LogInfo("Should have automountServiceAccountToken set to false")

	msg := []string{}
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		check.LogInfo("check the existence of pod service account %s (ns= %s )", put.Namespace, put.Name)
		if put.Spec.ServiceAccountName == defaultServiceAccount {
			check.LogDebug("Pod %s has been found with default service account name.", put.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has been found with default service account name", false))
			break
		}

		// Evaluate the pod's automount service tokens and any attached service accounts
		podPassed, newMsg := rbac.EvaluateAutomountTokens(put.Pod)
		if !podPassed {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, newMsg, false))
			msg = append(msg, newMsg)
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod does not have automount service tokens set to true", true))
		}
	}

	if len(msg) > 0 {
		check.LogDebug(strings.Join(msg, ""))
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testOneProcessPerContainer(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, cut := range env.Containers {
		// the Istio sidecar container "istio-proxy" launches two processes: "pilot-agent" and "envoy"
		if cut.IsIstioProxy() {
			continue
		}
		debugPod := env.DebugPods[cut.NodeName]
		if debugPod == nil {
			check.LogError("Debug pod not found for node %s", cut.NodeName)
			return
		}
		ocpContext := clientsholder.NewContext(debugPod.Namespace, debugPod.Name, debugPod.Spec.Containers[0].Name)
		pid, err := crclient.GetPidFromContainer(cut, ocpContext)
		if err != nil {
			check.LogDebug("Could not get PID for: %s, error: %v", cut, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, err.Error(), false))
			continue
		}

		nbProcesses, err := getNbOfProcessesInPidNamespace(ocpContext, pid, clientsholder.GetClientsHolder())
		if err != nil {
			check.LogDebug("Could not get number of processes for: %s, error: %v", cut, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, err.Error(), false))
			continue
		}
		if nbProcesses > 1 {
			check.LogDebug("%s has more than one process running", cut.String())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has more than one process running", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has only one process running", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testSYSNiceRealtimeCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	// Loop through all of the labeled containers and compare their security context capabilities and whether
	// or not the node's kernel is realtime enabled.
	for _, cut := range env.Containers {
		n := env.Nodes[cut.NodeName]
		if n.IsRTKernel() && !strings.Contains(cut.SecurityContext.Capabilities.String(), "SYS_NICE") {
			check.LogDebug("%s has been found running on a realtime kernel enabled node without SYS_NICE capability.", cut.String())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container is running on a realtime kernel enabled node without SYS_NICE capability", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container is not running on a realtime kernel enabled node", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testSysPtraceCapability(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.GetShareProcessNamespacePods() {
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
			check.LogDebug("Pod %s has process namespace sharing enabled but no container allowing the SYS_PTRACE capability.", put.String())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has process namespace sharing enabled but no container allowing the SYS_PTRACE capability", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has process namespace sharing enabled and at least one container allowing the SYS_PTRACE capability", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testNamespaceResourceQuota(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	check.LogInfo("Testing namespace resource quotas")

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
			check.LogDebug("Pod %s is running in a namespace that does not have a ResourceQuota applied.", put.String())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is running in a namespace that does not have a ResourceQuota applied", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is running in a namespace that has a ResourceQuota applied", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

const (
	sshServicePortProtocol = "TCP"
)

func testNoSSHDaemonsAllowed(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, put := range env.Pods {
		cut := put.Containers[0]

		// 1. Find SSH port
		port, err := netutil.GetSSHDaemonPort(cut)
		if err != nil {
			check.LogError("could not get ssh daemon port on %s, err: %v", cut, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Failed to get the ssh port for pod", false))
			continue
		}

		if port == "" {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is not running an SSH daemon", true))
			continue
		}

		sshServicePortNumber, err := strconv.Atoi(port)
		if err != nil {
			log.Error("error occurred while converting port %s from string to integer on %s", port, cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Failed to get the listening ports for pod", false))
			continue
		}

		// 2. Check if SSH port is listening
		sshPortInfo := netutil.PortInfo{PortNumber: sshServicePortNumber, Protocol: sshServicePortProtocol}
		listeningPorts, err := netutil.GetListeningPorts(cut)
		if err != nil {
			check.LogDebug("Failed to get the listening ports on %s, err: %v", cut, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Failed to get the listening ports for pod", false))
			continue
		}

		if _, ok := listeningPorts[sshPortInfo]; ok {
			check.LogDebug("Pod %s is running an SSH daemon", put)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is running an SSH daemon", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is not running an SSH daemon", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testPodRequestsAndLimits(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	check.LogInfo("Testing container resource requests and limits")

	// Loop through the containers, looking for containers that are missing requests or limits.
	// These need to be defined in order to pass.
	for _, cut := range env.Containers {
		if !resources.HasRequestsAndLimitsSet(cut) {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container is missing resource requests or limits", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has resource requests and limits", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func test1337UIDs(check *checksdb.Check, env *provider.TestEnvironment) {
	// Note this test is only ran as part of the 'extended' test suite.
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	const leetNum = 1337
	check.LogInfo("Testing pods to ensure none are using UID 1337")
	for _, put := range env.Pods {
		check.LogInfo("checking if pod %s has a securityContext RunAsUser 1337 (ns= %s)", put.Name, put.Namespace)
		if put.IsRunAsUserID(leetNum) {
			check.LogDebug("Pod: %s/%s is found to use securityContext RunAsUser 1337", put.Namespace, put.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is using securityContext RunAsUser 1337", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is not using securityContext RunAsUser 1337", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// a test for security context that are allowed from the documentation of the cnf
// an allowed one will pass the test

func testContainerSCC(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	highLevelCat := securitycontextcontainer.CategoryID1
	for _, pod := range env.Pods {
		listCategory := securitycontextcontainer.CheckPod(pod)
		for _, cat := range listCategory {
			if cat.Category > securitycontextcontainer.CategoryID1NoUID0 {
				aContainerOut := testhelper.NewContainerReportObject(cat.NameSpace, cat.Podname, cat.Containername, "container category is NOT category 1 or category NoUID0", false).
					SetType(testhelper.ContainerCategory).
					AddField(testhelper.Category, cat.Category.String())
				nonCompliantObjects = append(nonCompliantObjects, aContainerOut)
			} else {
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

func testNodePort(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, s := range env.Services {
		check.LogInfo("Testing %s", services.ToString(s))

		if s.Spec.Type == nodePort {
			check.LogDebug("FAILURE: Service %s (ns %s) type is nodePort", s.Name, s.Namespace)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject("Service is type NodePort", testhelper.ServiceType, false).
				AddField(testhelper.Namespace, s.Namespace).
				AddField(testhelper.ServiceName, s.Name).
				AddField(testhelper.ServiceMode, string(s.Spec.Type)))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewReportObject("Service is not type NodePort", testhelper.ServiceType, true).
				AddField(testhelper.Namespace, s.Namespace).
				AddField(testhelper.ServiceName, s.Name).
				AddField(testhelper.ServiceMode, string(s.Spec.Type)))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

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
			compliantObjects = append(compliantObjects, testhelper.NewNamespacedReportObject("This applies to CRDs under test", testhelper.RoleRuleType, true, env.Roles[roleIndex].Namespace).
				AddField(testhelper.RoleName, env.Roles[roleIndex].Name).
				AddField(testhelper.Group, aRule.Resource.Group).
				AddField(testhelper.ResourceName, aRule.Resource.Name).
				AddField(testhelper.Verb, aRule.Verb))
		}
		for _, aRule := range nonMatchingRules {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNamespacedReportObject("This rule does not apply to CRDs under test", testhelper.RoleRuleType, false, env.Roles[roleIndex].Namespace).
				AddField(testhelper.RoleName, env.Roles[roleIndex].Name).
				AddField(testhelper.Group, aRule.Resource.Group).
				AddField(testhelper.ResourceName, aRule.Resource.Name).
				AddField(testhelper.Verb, aRule.Verb))
		}

		if len(nonMatchingRules) == 0 {
			compliantObjects = append(compliantObjects, testhelper.NewNamespacedNamedReportObject("This role's rules only apply to CRDs under test",
				testhelper.RoleType, true, env.Roles[roleIndex].Namespace, env.Roles[roleIndex].Name))
		} else {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNamespacedNamedReportObject("This role's rules apply to a mix of CRDs under test and others. See non compliant role rule objects.",
				testhelper.RoleType, false, env.Roles[roleIndex].Namespace, env.Roles[roleIndex].Name))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}
