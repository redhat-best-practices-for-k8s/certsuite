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

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/namespace"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/rbac"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/resources"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/securitycontextcontainer"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/netutil"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/services"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/internal/crclient"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/stringhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
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

var _ = ginkgo.Describe(common.AccessControlTestKey, func() {

	logrus.Debugf("Entering %s suite", common.AccessControlTestKey)
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})
	ginkgo.ReportAfterEach(results.RecordResult)
	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSecContextIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		testContainerSCC(&env)
	})

	// Security Context: non-compliant capabilities (SYS_ADMIN)
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSysAdminIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		testSysAdminCapability(&env)
	})

	// Security Context: non-compliant capabilities (NET_ADMIN)
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNetAdminIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		testNetAdminCapability(&env)
	})

	// Security Context: non-compliant capabilities (NET_RAW)
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNetRawIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		testNetRawCapability(&env)
	})

	// Security Context: non-compliant capabilities (IPC_LOCK)
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestIpcLockIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		testIpcLockCapability(&env)
	})

	// Security Context: non-compliant capabilities (BPF)
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestBpfIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		testBpfCapability(&env)
	})

	// container security context: non-root user
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSecConNonRootUserIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		testSecConRootUser(&env)
	})
	// container security context: privileged escalation
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSecConPrivilegeEscalation)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		testSecConPrivilegeEscalation(&env)
	})
	// container security context: host port
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestContainerHostPort)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		testContainerHostPort(&env)
	})
	// container security context: host network
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodHostNetwork)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"))
		testPodHostNetwork(&env)
	})
	// pod host path
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodHostPath)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"))
		testPodHostPath(&env)
	})
	// pod host ipc
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodHostIPC)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"))
		testPodHostIPC(&env)
	})
	// pod host pid
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodHostPID)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"))
		testPodHostPID(&env)
	})
	// Namespace
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNamespaceBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Namespaces, "env.Namespaces"))
		testNamespace(&env)
	})
	// pod service account
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodServiceAccountBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"))
		testPodServiceAccount(&env)
	})
	// pod role bindings
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodRoleBindingsBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"))
		testPodRoleBindings(&env)
	})
	// pod cluster role bindings
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodClusterRoleBindingsBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"))
		testPodClusterRoleBindings(&env)
	})
	// automount service token
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodAutomountServiceAccountIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"))
		testAutomountServiceToken(&env)
	})
	// one process per container
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestOneProcessPerContainerIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		if env.DaemonsetFailedToSpawn {
			ginkgo.Skip("Debug Daemonset failed to spawn skipping TestOneProcessPerContainer")
		}
		testOneProcessPerContainer(&env)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSYSNiceRealtimeCapabilityIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"))
		testSYSNiceRealtimeCapability(&env)
	})
	// SYS_PTRACE capability
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSysPtraceCapabilityIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		shareProcessPods := env.GetShareProcessNamespacePods()
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(shareProcessPods, "shareProcessPods"))
		testSysPtraceCapability(shareProcessPods)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNamespaceResourceQuotaIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"))
		testNamespaceResourceQuota(&env)
	})
	// ssh daemons
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNoSSHDaemonsAllowedIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		if env.DaemonsetFailedToSpawn {
			ginkgo.Skip("Debug Daemonset failed to spawn skipping TestNoSSHDaemonsAllowed")
		}
		testNoSSHDaemonsAllowed(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodRequestsAndLimitsIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"))
		testPodRequestsAndLimits(&env)
	})

	// no 1337 UID's being used by pods
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.Test1337UIDIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"))
		test1337UIDs(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestServicesDoNotUseNodeportsIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"), testhelper.NewSkipObject(env.Pods, "env.Pods"))
		testNodePort(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestCrdRoleIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Crds, "env.Crds"), testhelper.NewSkipObject(env.Roles, "env.Roles"), testhelper.NewSkipObject(env.Namespaces, "env.Namespaces"))
		testCrdRoles(&env)
	})
})

func checkForbiddenCapability(containers []*provider.Container, capability string) (compliantObjects, nonCompliantObjects []*testhelper.ReportObject) {
	for _, cut := range containers {
		if cut.SecurityContext != nil && cut.SecurityContext.Capabilities != nil {
			if strings.Contains(cut.SecurityContext.Capabilities.String(), capability) {
				tnf.ClaimFilePrintf("Non compliant %s capability detected in container %s. All container caps: %s", capability, cut.String(), cut.SecurityContext.Capabilities.String())
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Non compliant capability detected in container", false).AddField(testhelper.SCCCapability, capability))
			} else {
				compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "No forbidden capabilities detected in container", true))
			}
		}
	}
	return compliantObjects, nonCompliantObjects
}

func testSysAdminCapability(env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(env.Containers, "SYS_ADMIN")
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testNetAdminCapability(env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(env.Containers, "NET_ADMIN")
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testNetRawCapability(env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(env.Containers, "NET_RAW")
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testIpcLockCapability(env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(env.Containers, "IPC_LOCK")
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testBpfCapability(env *provider.TestEnvironment) {
	compliantObjects, nonCompliantObjects := checkForbiddenCapability(env.Containers, "BPF")
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testSecConRootUser verifies that the container is not running as root
func testSecConRootUser(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		if put.IsRunAsUserID(0) {
			tnf.ClaimFilePrintf("Non compliant run as Root User detected (RunAsUser uid=0) in pod %s", put.Namespace+"."+put.Name)
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

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testSecConPrivilegeEscalation verifies that the container is not allowed privilege escalation
func testSecConPrivilegeEscalation(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		if cut.SecurityContext != nil && cut.SecurityContext.AllowPrivilegeEscalation != nil {
			if *(cut.SecurityContext.AllowPrivilegeEscalation) {
				tnf.ClaimFilePrintf("AllowPrivilegeEscalation is set to true in container %s.", cut.Podname+"."+cut.Name)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "AllowPrivilegeEscalation is set to true", false))
			} else {
				compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "AllowPrivilegeEscalation is set to false", true))
			}
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testContainerHostPort tests that containers are not configured with host port privileges
func testContainerHostPort(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		for _, aPort := range cut.Ports {
			if aPort.HostPort != 0 {
				tnf.ClaimFilePrintf("Host port %d is configured in container %s.", aPort.HostPort, cut.String())
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Host port is configured", false).
					SetType(testhelper.HostPortType).
					AddField(testhelper.PortNumber, strconv.Itoa(int(aPort.HostPort))))
			} else {
				compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Host port is not configured", true))
			}
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testPodHostNetwork verifies that the pod hostNetwork parameter is not set to true
func testPodHostNetwork(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		if put.Spec.HostNetwork {
			tnf.ClaimFilePrintf("Host network is set to true in pod %s.", put.Namespace+"."+put.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Host network is set to true", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Host network is not set to true", true))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testPodHostPath verifies that the pod hostpath parameter is not set to true
func testPodHostPath(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		podIsCompliant := true
		for idx := range put.Spec.Volumes {
			vol := &put.Spec.Volumes[idx]
			if vol.HostPath != nil && vol.HostPath.Path != "" {
				tnf.ClaimFilePrintf("Hostpath path: %s is set in pod %s.", vol.HostPath.Path, put.Namespace+"."+put.Name)
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

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testPodHostIPC verifies that the pod hostIpc parameter is not set to true
func testPodHostIPC(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		if put.Spec.HostIPC {
			tnf.ClaimFilePrintf("HostIpc is set in pod %s.", put.Namespace+"."+put.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "HostIpc is set to true", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "HostIpc is not set to true", true))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testPodHostPID verifies that the pod hostPid parameter is not set to true
func testPodHostPID(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		if put.Spec.HostPID {
			tnf.ClaimFilePrintf("HostPid is set in pod %s.", put.Namespace+"."+put.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "HostPid is set to true", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "HostPid is not set to true", true))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// Tests namespaces for invalid prefixed and CRs are not defined in namespaces not under test with CRDs under test
func testNamespace(env *provider.TestEnvironment) {
	ginkgo.By(fmt.Sprintf("CNF resources' Namespaces should not have any of the following prefixes: %v", invalidNamespacePrefixes))
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, namespace := range env.Namespaces {
		namespaceCompliant := true
		ginkgo.By(fmt.Sprintf("Checking namespace %s", namespace))
		for _, invalidPrefix := range invalidNamespacePrefixes {
			if strings.HasPrefix(namespace, invalidPrefix) {
				tnf.ClaimFilePrintf("Namespace %s has invalid prefix %s", namespace, invalidPrefix)
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
		testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
	}
	ginkgo.By(fmt.Sprintf("CNF pods should belong to any of the configured Namespaces: %v", env.Namespaces))
	ginkgo.By(fmt.Sprintf("CRs from autodiscovered CRDs should belong only to the configured Namespaces: %v", env.Namespaces))
	invalidCrs, err := namespace.TestCrsNamespaces(env.Crds, env.Namespaces)
	if err != nil {
		// Note: Leaving this as a direct call to Ginkgo fail instead of using the testhelper
		ginkgo.Fail("error retrieving CRs")
	}

	invalidCrsNum, claimsLog := namespace.GetInvalidCRsNum(invalidCrs)
	if invalidCrsNum > 0 && len(claimsLog.GetLogLines()) > 0 {
		tnf.ClaimFilePrintf("%s", claimsLog.GetLogLines())
		nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject("CRs are not in the configured namespaces", testhelper.Namespace, false))
	} else {
		compliantObjects = append(compliantObjects, testhelper.NewReportObject("CRs are in the configured namespaces", testhelper.Namespace, true))
	}
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testPodServiceAccount verifies that the pod utilizes a valid service account
func testPodServiceAccount(env *provider.TestEnvironment) {
	ginkgo.By("Tests that each pod utilizes a valid service account")
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		ginkgo.By(fmt.Sprintf("Testing service account for pod %s (ns: %s)", put.Name, put.Namespace))
		if put.Spec.ServiceAccountName == defaultServiceAccount {
			tnf.ClaimFilePrintf("Pod %s (ns: %s) does not have a valid service account name.", put.Name, put.Namespace)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod does not have a valid service account name", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has a service account name", true))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testPodRoleBindings verifies that the pod utilizes a valid role binding that does not cross non-CNF namespaces
//
//nolint:funlen
func testPodRoleBindings(env *provider.TestEnvironment) {
	ginkgo.By("Should not have RoleBinding in other namespaces")
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, put := range env.Pods {
		podIsCompliant := true
		ginkgo.By(fmt.Sprintf("Testing role binding for pod: %s namespace: %s", put.Name, put.Namespace))
		if put.Pod.Spec.ServiceAccountName == defaultServiceAccount {
			logrus.Infof("%s has an empty or default serviceAccountName, skipping.", put.String())
			// Add the pod to the non-compliant list
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewPodReportObject(put.Namespace, put.Name,
					"The serviceAccountName is either empty or default", false))
			podIsCompliant = false
		} else {
			logrus.Infof("%s has a serviceAccountName: %s, checking role bindings.", put.String(), put.Spec.ServiceAccountName)
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
						stringhelper.StringInSlice(env.Namespaces, env.RoleBindings[rbIndex].Namespace, false) {
						continue
					}

					// Finally, if the subject is a service account and the service account is in the same namespace as the pod, then we have a failure
					if subject.Kind == rbacv1.ServiceAccountKind &&
						subject.Namespace == put.Namespace &&
						subject.Name == put.Spec.ServiceAccountName {
						tnf.Logf(logrus.WarnLevel, "Pod: %s has the following role bindings that do not live in one of the CNF namespaces: %s", put, env.RoleBindings[rbIndex].Name)

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
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testPodClusterRoleBindings verifies that the pod does not use a cluster role binding
//
//nolint:dupl
func testPodClusterRoleBindings(env *provider.TestEnvironment) {
	ginkgo.By("Pods should not have ClusterRoleBindings")
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	logrus.Infof("There were %d cluster role bindings found in the cluster.", len(env.ClusterRoleBindings))

	for _, put := range env.Pods {
		podIsCompliant := true
		ginkgo.By(fmt.Sprintf("Testing cluster role binding for pod: %s namespace: %s", put.Name, put.Namespace))
		result, roleRefName, err := put.IsUsingClusterRoleBinding(env.ClusterRoleBindings)
		if err != nil {
			logrus.Errorf("failed to determine if pod %s/%s is using a cluster role binding: %v", put.Namespace, put.Name, err)
			podIsCompliant = false
		}

		// Pod was found to be using a cluster role binding.  This is not allowed.
		// Flagging this pod as a failed pod.
		if result {
			tnf.Logf(logrus.WarnLevel, "%s is using a cluster role binding", put.String())
			podIsCompliant = false
		}

		if podIsCompliant {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is not using a cluster role binding", true))
		} else {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is using a cluster role binding", false).
				AddField(testhelper.ClusterRoleName, roleRefName))
		}
	}
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testAutomountServiceToken(env *provider.TestEnvironment) {
	ginkgo.By("Should have automountServiceAccountToken set to false")

	msg := []string{}
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		ginkgo.By(fmt.Sprintf("check the existence of pod service account %s (ns= %s )", put.Namespace, put.Name))
		if put.Spec.ServiceAccountName == defaultServiceAccount {
			tnf.ClaimFilePrintf("Pod %s has been found with default service account name.", put.Name)
			ginkgo.Fail("Pod has been found with default service account name.")
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
		tnf.ClaimFilePrintf(strings.Join(msg, ""))
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testOneProcessPerContainer(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, cut := range env.Containers {
		// the Istio sidecar container "istio-proxy" launches two processes: "pilot-agent" and "envoy"
		if cut.IsIstioProxy() {
			continue
		}
		debugPod := env.DebugPods[cut.NodeName]
		if debugPod == nil {
			ginkgo.Fail(fmt.Sprintf("Debug pod not found on Node: %s", cut.NodeName))
		}
		ocpContext := clientsholder.NewContext(debugPod.Namespace, debugPod.Name, debugPod.Spec.Containers[0].Name)
		pid, err := crclient.GetPidFromContainer(cut, ocpContext)
		if err != nil {
			tnf.ClaimFilePrintf("Could not get PID for: %s, error: %v", cut, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, err.Error(), false))
			continue
		}

		nbProcesses, err := getNbOfProcessesInPidNamespace(ocpContext, pid, clientsholder.GetClientsHolder())
		if err != nil {
			tnf.ClaimFilePrintf("Could not get number of processes for: %s, error: %v", cut, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, err.Error(), false))
			continue
		}
		if nbProcesses > 1 {
			tnf.ClaimFilePrintf("%s has more than one process running", cut.String())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has more than one process running", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has only one process running", true))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testSYSNiceRealtimeCapability(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	// Loop through all of the labeled containers and compare their security context capabilities and whether
	// or not the node's kernel is realtime enabled.
	for _, cut := range env.Containers {
		n := env.Nodes[cut.NodeName]
		if n.IsRTKernel() && !strings.Contains(cut.SecurityContext.Capabilities.String(), "SYS_NICE") {
			tnf.ClaimFilePrintf("%s has been found running on a realtime kernel enabled node without SYS_NICE capability.", cut.String())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container is running on a realtime kernel enabled node without SYS_NICE capability", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container is not running on a realtime kernel enabled node", true))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testSysPtraceCapability(shareProcessPods []*provider.Pod) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
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
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has process namespace sharing enabled but no container allowing the SYS_PTRACE capability", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has process namespace sharing enabled and at least one container allowing the SYS_PTRACE capability", true))
		}
	}
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testNamespaceResourceQuota(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
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
			tnf.ClaimFilePrintf("Pod %s is running in a namespace that does not have a ResourceQuota applied.", put.String())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is running in a namespace that does not have a ResourceQuota applied", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is running in a namespace that has a ResourceQuota applied", true))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

const (
	sshServicePortProtocol = "TCP"
)

func findSSHPort(cut *provider.Container) (port string, isError bool) {
	port, err := netutil.GetSSHDaemonPort(cut)
	if err != nil {
		isError = true
		logrus.Errorf("error occurred while finding ssh port on %s, err %v", cut, err)
	}
	return port, isError
}

func testNoSSHDaemonsAllowed(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, put := range env.Pods {
		cut := put.Containers[0]

		// 1. Find SSH port
		port, foundError := findSSHPort(cut)
		if foundError {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Failed to get the ssh port for pod", false))
			continue
		}

		if port == "" {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is not running an SSH daemon", true))
			continue
		}

		sshServicePortNumber, err := strconv.Atoi(port)
		if err != nil {
			logrus.Errorf("error occurred while converting port %s from string to integer on %s", port, cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Failed to get the listening ports for pod", false))
			continue
		}

		// 2. Check if SSH port is listening
		sshPortInfo := netutil.PortInfo{PortNumber: sshServicePortNumber, Protocol: sshServicePortProtocol}
		listeningPorts, err := netutil.GetListeningPorts(cut)
		if err != nil {
			tnf.ClaimFilePrintf("Failed to get the listening ports on %s, err: %v", cut, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Failed to get the listening ports for pod", false))
			continue
		}

		if _, ok := listeningPorts[sshPortInfo]; ok {
			tnf.ClaimFilePrintf("Pod %s is running an SSH daemon", put)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is running an SSH daemon", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is not running an SSH daemon", true))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testPodRequestsAndLimits(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	ginkgo.By("Testing container resource requests and limits")

	// Loop through the containers, looking for containers that are missing requests or limits.
	// These need to be defined in order to pass.
	for _, cut := range env.Containers {
		if !resources.HasRequestsAndLimitsSet(cut) {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container is missing resource requests or limits", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has resource requests and limits", true))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func test1337UIDs(env *provider.TestEnvironment) {
	// Note this test is only ran as part of the 'extended' test suite.
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	const leetNum = 1337
	ginkgo.By("Testing pods to ensure none are using UID 1337")
	for _, put := range env.Pods {
		ginkgo.By(fmt.Sprintf("checking if pod %s has a securityContext RunAsUser 1337 (ns= %s)", put.Name, put.Namespace))
		if put.IsRunAsUserID(leetNum) {
			tnf.ClaimFilePrintf("Pod: %s/%s is found to use securityContext RunAsUser 1337", put.Namespace, put.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is using securityContext RunAsUser 1337", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is not using securityContext RunAsUser 1337", true))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// a test for security context that are allowed from the documentation of the cnf
// an allowed one will pass the test

func testContainerSCC(env *provider.TestEnvironment) {
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
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testNodePort(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, s := range env.Services {
		ginkgo.By(fmt.Sprintf("Testing %s", services.ToString(s)))

		if s.Spec.Type == nodePort {
			tnf.ClaimFilePrintf("FAILURE: Service %s (ns %s) type is nodePort", s.Name, s.Namespace)
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
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testCrdRoles(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	crdResources := rbac.GetCrdResources(env.Crds)
	for roleIndex := range env.Roles {
		if !stringhelper.StringInSlice(env.Namespaces, env.Roles[roleIndex].Namespace, false) {
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
	if len(nonCompliantObjects) == 0 && len(compliantObjects) == 0 {
		ginkgo.Skip("No role contains rules that apply to at least one CRD under test")
	}
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}
