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

package networking

import (
	"fmt"
	"strconv"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/networking/icmp"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/networking/netcommons"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/networking/netutil"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/networking/policies"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/networking/services"
	networkingv1 "k8s.io/api/networking/v1"
)

const (
	defaultNumPings = 5
	nodePort        = "NodePort"
)

type Port []struct {
	ContainerPort int
	Name          string
	Protocol      string
}

var (
	env provider.TestEnvironment

	beforeEachFn = func(check *checksdb.Check) error {
		env = provider.GetTestEnvironment()
		return nil
	}
)

// LoadChecks registers all networking checks and returns a setup function.
//
// It creates several check groups, adds individual checks for network
// connectivity and pod/daemonset conditions, and configures skip logic
// based on the test environment. The returned closure performs any
// required initialization before the checks are executed.
func LoadChecks() {
	log.Debug("Loading %s suite checks", common.NetworkingTestKey)

	checksGroup := checksdb.NewChecksGroup(common.NetworkingTestKey).
		WithBeforeEachFn(beforeEachFn)

	// Default interface ICMP IPv4 test case
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestICMPv4ConnectivityIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNetworkConnectivity(&env, netcommons.IPv4, netcommons.DEFAULT, c)
			return nil
		}))

	// Multus interfaces ICMP IPv4 test case
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestICMPv4ConnectivityMultusIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNetworkConnectivity(&env, netcommons.IPv4, netcommons.MULTUS, c)
			return nil
		}))

	// Default interface ICMP IPv6 test case
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestICMPv6ConnectivityIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNetworkConnectivity(&env, netcommons.IPv6, netcommons.DEFAULT, c)
			return nil
		}))

	// Multus interfaces ICMP IPv6 test case
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestICMPv6ConnectivityMultusIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNetworkConnectivity(&env, netcommons.IPv6, netcommons.MULTUS, c)
			return nil
		}))

	// Undeclared container ports usage test case
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestUndeclaredContainerPortsUsage)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testUndeclaredContainerPortsUsage(c, &env)
			return nil
		}))

	// OCP reserved ports usage test case
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestOCPReservedPortsUsage)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOCPReservedPortsUsage(c, &env)
			return nil
		}))

	// Dual stack services test case
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestServiceDualStackIdentifier)).
		WithSkipCheckFn(testhelper.GetNoServicesUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testDualStackServices(c, &env)
			return nil
		}))

	// Network policy deny all test case
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestNetworkPolicyDenyAllIdentifier)).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNetworkPolicyDenyAll(c, &env)
			return nil
		}))

	// Extended partner ports test case
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestReservedExtendedPartnerPorts)).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPartnerSpecificTCPPorts(c, &env)
			return nil
		}))

	// DPDK CPU pinning exec probe test case
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestDpdkCPUPinningExecProbe)).
		WithSkipCheckFn(testhelper.GetNoCPUPinningPodsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			dpdkPods := env.GetCPUPinningPodsWithDpdk()
			testExecProbDenyAtCPUPinning(c, dpdkPods)
			return nil
		}))

	// Restart on reboot label test case
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestRestartOnRebootLabelOnPodsUsingSRIOV)).
		WithSkipCheckFn(testhelper.GetNoSRIOVPodsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			sriovPods, err := env.GetPodsUsingSRIOV()
			if err != nil {
				return fmt.Errorf("failure getting pods using SRIOV: %v", err)
			}
			testRestartOnRebootLabelOnPodsUsingSriov(c, sriovPods)
			return nil
		}))

	// SRIOV MTU test case
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(
		identifiers.TestNetworkAttachmentDefinitionSRIOVUsingMTU)).
		WithSkipCheckFn(testhelper.GetNoSRIOVPodsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			sriovPods, err := env.GetPodsUsingSRIOV()
			if err != nil {
				return fmt.Errorf("failure getting pods using SRIOV: %v", err)
			}
			testNetworkAttachmentDefinitionSRIOVUsingMTU(c, sriovPods)
			return nil
		}))
}

// testExecProbDenyAtCPUPinning verifies that containers with exec probes set to deny at CPU pinning are correctly reported.
//
// It receives a Check object and a slice of Pod objects, examines each pod for the presence of exec probes,
// logs relevant information, creates report entries for pods lacking such probes, and updates the check result
// accordingly. The function does not return a value but modifies the supplied Check via SetResult.
func testExecProbDenyAtCPUPinning(check *checksdb.Check, dpdkPods []*provider.Pod) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, cpuPinnedPod := range dpdkPods {
		execProbeFound := false
		for _, cut := range cpuPinnedPod.Containers {
			check.LogInfo("Testing Container %q", cut)
			if cut.HasExecProbes() {
				check.LogError("Container %q defines an exec probe", cut)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(cpuPinnedPod.Namespace, cpuPinnedPod.Name, "Exec prob is not allowed", false))
				execProbeFound = true
			}
		}

		if !execProbeFound {
			check.LogInfo("Pod %q does not define any exec probe", cpuPinnedPod)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(cpuPinnedPod.Namespace, cpuPinnedPod.Name, "Exec prob is allowed", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testUndeclaredContainerPortsUsage examines each pod in the environment to determine if any containers expose ports that are not declared in their manifests.  
// It retrieves listening ports from the cluster, compares them against the expected port list for each container, and records discrepancies in a report object.  
// The function logs progress and errors during its execution and updates the test result with details of any undeclared ports found.
func testUndeclaredContainerPortsUsage(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	var portInfo netutil.PortInfo
	for _, put := range env.Pods {
		// First get the ports declared in the Pod's containers spec
		declaredPorts := make(map[netutil.PortInfo]bool)
		for _, cut := range put.Containers {
			check.LogInfo("Testing Container %q", cut)
			for _, port := range cut.Ports {
				portInfo.PortNumber = port.ContainerPort
				portInfo.Protocol = string(port.Protocol)
				declaredPorts[portInfo] = true
			}
		}

		// Then check the actual ports that the containers are listening on
		firstPodContainer := put.Containers[0]
		listeningPorts, err := netutil.GetListeningPorts(firstPodContainer)
		if err != nil {
			check.LogError("Failed to get container %q listening ports, err: %v", firstPodContainer, err)
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewPodReportObject(put.Namespace, put.Name, fmt.Sprintf("Failed to get the container's listening ports, err: %v", err), false))
			continue
		}
		if len(listeningPorts) == 0 {
			check.LogInfo("None of the containers of %q have any listening port.", put)
			compliantObjects = append(compliantObjects,
				testhelper.NewPodReportObject(put.Namespace, put.Name, "None of the containers have any listening ports", true))
			continue
		}

		// Verify that all the listening ports have been declared in the container spec
		failedPod := false
		for listeningPort := range listeningPorts {
			if put.ContainsIstioProxy() && netcommons.ReservedIstioPorts[listeningPort.PortNumber] {
				check.LogInfo("%q is listening on port %d protocol %q, but the pod also contains istio-proxy. Ignoring.",
					put, listeningPort.PortNumber, listeningPort.Protocol)
				continue
			}

			if ok := declaredPorts[listeningPort]; !ok {
				check.LogError("%q is listening on port %d protocol %q, but that port was not declared in any container spec.",
					put, listeningPort.PortNumber, listeningPort.Protocol)
				failedPod = true
				nonCompliantObjects = append(nonCompliantObjects,
					testhelper.NewPodReportObject(put.Namespace, put.Name,
						"Listening port was declared in no container spec", false).
						SetType(testhelper.ListeningPortType).
						AddField(testhelper.PortNumber, strconv.Itoa(int(listeningPort.PortNumber))).
						AddField(testhelper.PortProtocol, listeningPort.Protocol))
			} else {
				check.LogInfo("%q is listening on declared port %d protocol %q", put, listeningPort.PortNumber, listeningPort.Protocol)
				compliantObjects = append(compliantObjects,
					testhelper.NewPodReportObject(put.Namespace, put.Name,
						"Listening port was declared in container spec", true).
						SetType(testhelper.ListeningPortType).
						AddField(testhelper.PortNumber, strconv.Itoa(int(listeningPort.PortNumber))).
						AddField(testhelper.PortProtocol, listeningPort.Protocol))
			}
		}
		if failedPod {
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewPodReportObject(put.Namespace, put.Name, "At least one port was listening but not declared in any container specs", false))
		} else {
			compliantObjects = append(compliantObjects,
				testhelper.NewPodReportObject(put.Namespace, put.Name, "All listening were declared in containers specs", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testNetworkConnectivity verifies that the default network interfaces of containers can reach each other across different IP versions and interface types.
//
// It accepts a test environment, an IP version, an interface type, and a check definition.
// The function builds a networking test context, logs progress, runs connectivity tests,
// and records the result. The returned closure performs these steps when executed.
func testNetworkConnectivity(env *provider.TestEnvironment, aIPVersion netcommons.IPVersion, aType netcommons.IFType, check *checksdb.Check) {
	netsUnderTest := icmp.BuildNetTestContext(env.Pods, aIPVersion, aType, check.GetLogger())
	report, skip := icmp.RunNetworkingTests(netsUnderTest, defaultNumPings, aIPVersion, check.GetLogger())
	if skip {
		check.LogInfo("There are no %q networks to test with at least 2 pods, skipping test", aIPVersion)
	}
	check.SetResult(report.CompliantObjectsOut, report.NonCompliantObjectsOut)
}

// testOCPReservedPortsUsage tests that OpenShift reserved ports are correctly used by the system.
//
// It runs the TestReservedPortsUsage check against the provided test environment,
// logs any relevant information, and records the result of the check in the
// checks database. The function receives a pointer to a Check object for storing
// results and a pointer to the current TestEnvironment which contains all
// necessary configuration and context. No return value is produced; results are
// stored via SetResult.
func testOCPReservedPortsUsage(check *checksdb.Check, env *provider.TestEnvironment) {
	// List of all ports reserved by OpenShift
	OCPReservedPorts := map[int32]bool{
		22623: true,
		22624: true}
	compliantObjects, nonCompliantObjects := netcommons.TestReservedPortsUsage(env, OCPReservedPorts, "OCP", check.GetLogger())
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPartnerSpecificTCPPorts checks that the networking environment correctly handles reserved TCP ports for a given partner.
//
// It takes a Check object and a TestEnvironment pointer.
// The function logs its progress, verifies that the required
// partner-specific TCP ports are available using TestReservedPortsUsage,
// and records the result in the check. No value is returned.
func testPartnerSpecificTCPPorts(check *checksdb.Check, env *provider.TestEnvironment) {
	// List of all of the ports reserved by partner
	ReservedPorts := map[int32]bool{
		15443: true,
		15090: true,
		15021: true,
		15020: true,
		15014: true,
		15008: true,
		15006: true,
		15001: true,
		15000: true,
	}
	compliantObjects, nonCompliantObjects := netcommons.TestReservedPortsUsage(env, ReservedPorts, "Partner", check.GetLogger())
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testDualStackServices checks that a service is reachable over both IPv4 and IPv6 addresses.
//
// It retrieves the service IP version for each address family, attempts to ping
// the service from the test environment, and records success or failure in a
// report object. The function logs informational messages during execution,
// handles errors by logging them, and sets the overall result status based on
// whether all pings succeeded. No return value is produced; results are
// reported via the testing framework's reporting mechanisms.
func testDualStackServices(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, s := range env.Services {
		check.LogInfo("Testing Service %q", s.Name)
		serviceIPVersion, err := services.GetServiceIPVersion(s)
		if err != nil {
			check.LogError("Could not get IP version from Service %q, err=%v", s.Name, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject("Could not get IP Version from service", testhelper.ServiceType, false).
				AddField(testhelper.Namespace, s.Namespace).
				AddField(testhelper.ServiceName, s.Name))
		}
		if serviceIPVersion == netcommons.Undefined || serviceIPVersion == netcommons.IPv4 {
			check.LogError("Service %q (ns: %q) only supports IPv4", s.Name, s.Namespace)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject("Service supports only IPv4", testhelper.ServiceType, false).
				AddField(testhelper.Namespace, s.Namespace).
				AddField(testhelper.ServiceName, s.Name).
				AddField(testhelper.ServiceIPVersion, serviceIPVersion.String()))
		} else {
			check.LogInfo("Service %q (ns: %q) supports IPv6 or is dual stack", s.Name, s.Namespace)
			compliantObjects = append(compliantObjects, testhelper.NewReportObject("Service supports IPv6 or is dual stack", testhelper.ServiceType, true).
				AddField(testhelper.Namespace, s.Namespace).
				AddField(testhelper.ServiceName, s.Name).
				AddField(testhelper.ServiceIPVersion, serviceIPVersion.String()))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testNetworkPolicyDenyAll verifies that a Kubernetes cluster enforces a "deny all" NetworkPolicy correctly.
//
// It receives a check object and a test environment containing the namespace and pods to examine.
// The function logs progress, checks whether all pods match the expected labels,
// determines compliance of each pod with the deny‑all policy,
// records any errors encountered, and updates the check result accordingly.
func testNetworkPolicyDenyAll(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	// Loop through the pods, looking for corresponding entries within a deny-all network policy (both ingress and egress).
	// This ensures that each pod is accounted for that we are tasked with testing and excludes any pods that are not marked
	// for testing (via the labels).
	for _, put := range env.Pods {
		check.LogInfo("Testing Pod %q", put)
		denyAllEgressFound := false
		denyAllIngressFound := false

		// Look through all of the network policies for a matching namespace.
		for index := range env.NetworkPolicies {
			networkPolicy := env.NetworkPolicies[index]
			check.LogInfo("Testing Network policy %q against pod %q", networkPolicy.Name, put)

			// Skip any network policies that don't match the namespace of the pod we are testing.
			if networkPolicy.Namespace != put.Namespace {
				check.LogInfo("Skipping Network policy %q (namespace %q does not match Pod namespace %q)", networkPolicy.Name, networkPolicy.Namespace, put.Namespace)
				continue
			}

			// Match the pod namespace with the network policy namespace.
			if policies.LabelsMatch(networkPolicy.Spec.PodSelector, put.Labels) {
				var reason string
				if !denyAllEgressFound {
					denyAllEgressFound, reason = policies.IsNetworkPolicyCompliant(&networkPolicy, networkingv1.PolicyTypeEgress)
					if reason != "" {
						check.LogError("Network policy %q is not compliant, reason=%q", networkPolicy.Name, reason)
					}
				}
				if !denyAllIngressFound {
					denyAllIngressFound, reason = policies.IsNetworkPolicyCompliant(&networkPolicy, networkingv1.PolicyTypeIngress)
					if reason != "" {
						check.LogError("Network policy %q is not compliant, reason=%q", networkPolicy.Name, reason)
					}
				}
			}
		}

		// Network policy has not been found that contains a deny-all rule for both ingress and egress.
		podIsCompliant := true
		if !denyAllIngressFound {
			check.LogError("Pod %q was found to not have a default ingress deny-all network policy.", put)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod was found to not have a default ingress deny-all network policy", false))
			podIsCompliant = false
		}

		if !denyAllEgressFound {
			check.LogError("Pod %q was found to not have a default egress deny-all network policy.", put)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod was found to not have a default egress deny-all network policy", false))
			podIsCompliant = false
		}

		if podIsCompliant {
			check.LogInfo("Pod %q has a default ingress/egress deny-all network policy", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has a default ingress/egress deny-all network policy", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testRestartOnRebootLabelOnPodsUsingSriov checks that pods using SR‑IOV have the appropriate reboot label set and reports any discrepancies.
//
// It iterates over a slice of Pod objects, retrieves each pod’s labels,
// and verifies whether the reboot label is present when SR‑IOV is enabled.
// For each pod, it logs information about the check, records errors if
// labels are missing or incorrect, and appends a report object to the
// provided checks database. The function does not return a value but
// updates the check results via SetResult on the database.
func testRestartOnRebootLabelOnPodsUsingSriov(check *checksdb.Check, sriovPods []*provider.Pod) {
	const (
		restartOnRebootLabel = "restart-on-reboot"
	)

	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, pod := range sriovPods {
		check.LogInfo("Testing SRIOV Pod %q", pod)

		labelValue, exist := pod.GetLabels()[restartOnRebootLabel]
		if !exist {
			check.LogError("Pod %q uses SRIOV but the label %q was not found.", pod, restartOnRebootLabel)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(pod.Namespace, pod.Name, fmt.Sprintf("Pod uses SRIOV but the label %s was not found", restartOnRebootLabel), false))
			continue
		}

		if labelValue != "true" {
			check.LogError("Pod %q uses SRIOV but the %q label value is not true.", pod, restartOnRebootLabel)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(pod.Namespace, pod.Name, fmt.Sprintf("Pod uses SRIOV but the label %s is not set to true", restartOnRebootLabel), false))
			continue
		}

		check.LogInfo("Pod %q uses SRIOV and the %q label is set to true", pod, restartOnRebootLabel)
		compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(pod.Namespace, pod.Name, fmt.Sprintf("Pod uses SRIOV and the label %s is set to true", restartOnRebootLabel), true))
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testNetworkAttachmentDefinitionSRIOVUsingMTU checks whether pods using SR‑I/O‑V network attachments have the correct MTU setting.
//
// It receives a check database entry and a slice of pod objects, iterates over each pod to determine if it is configured with SR‑I/O‑V and an MTU value,
// logs any discrepancies, updates the pod report objects accordingly, and records the overall result in the check.
func testNetworkAttachmentDefinitionSRIOVUsingMTU(check *checksdb.Check, sriovPods []*provider.Pod) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, pod := range sriovPods {
		result, err := pod.IsUsingSRIOVWithMTU()
		if err != nil {
			check.LogError("Failed to check if pod %q uses SRIOV with MTU, err: %v", pod, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(pod.Namespace, pod.Name, "Failed to check if pod uses SRIOV with MTU", false))
			continue
		}

		if result {
			check.LogInfo("Pod %q uses SRIOV with MTU", pod)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(pod.Namespace, pod.Name, "Pod uses SRIOV with MTU", true))
		} else {
			check.LogError("Pod %q uses SRIOV but the MTU is not set explicitly", pod)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(pod.Namespace, pod.Name, "Pod uses SRIOV but the MTU is not set explicitly", false))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}
