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

package networking

import (
	"fmt"
	"strconv"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/icmp"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/netcommons"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/netutil"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/policies"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/services"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/checksdb"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
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
		check.LogInfo("Check %s: getting test environment.", check.ID)
		env = provider.GetTestEnvironment()
		return nil
	}
)

//nolint:funlen
func LoadChecks() {
	log.Debug("Entering %s suite", common.NetworkingTestKey)

	checksGroup := checksdb.NewChecksGroup(common.NetworkingTestKey).
		WithBeforeEachFn(beforeEachFn)

	// Default interface ICMP IPv4 test case
	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestICMPv4ConnectivityIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNetworkConnectivity(&env, netcommons.IPv4, netcommons.DEFAULT, c)
			return nil
		}))

	// Multus interfaces ICMP IPv4 test case
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestICMPv4ConnectivityMultusIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNetworkConnectivity(&env, netcommons.IPv4, netcommons.MULTUS, c)
			return nil
		}))

	// Default interface ICMP IPv6 test case
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestICMPv6ConnectivityIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNetworkConnectivity(&env, netcommons.IPv6, netcommons.DEFAULT, c)
			return nil
		}))

	// Multus interfaces ICMP IPv6 test case
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestICMPv6ConnectivityMultusIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNetworkConnectivity(&env, netcommons.IPv6, netcommons.MULTUS, c)
			return nil
		}))

	// Undeclared container ports usage test case
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestUndeclaredContainerPortsUsage)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testUndeclaredContainerPortsUsage(c, &env)
			return nil
		}))

	// OCP reserved ports usage test case
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestOCPReservedPortsUsage)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env), testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testOCPReservedPortsUsage(c, &env)
			return nil
		}))

	// Dual stack services test case
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestServiceDualStackIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoServicesUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testDualStackServices(c, &env)
			return nil
		}))

	// Network policy deny all test case
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNetworkPolicyDenyAllIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testNetworkPolicyDenyAll(c, &env)
			return nil
		}))

	// Extended partner ports test case
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestReservedExtendedPartnerPorts)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env), testhelper.GetDaemonSetFailedToSpawnSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPartnerSpecificTCPPorts(c, &env)
			return nil
		}))

	// DPDK CPU pinning exec probe test case
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestDpdkCPUPinningExecProbe)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoCPUPinningPodsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			dpdkPods := env.GetCPUPinningPodsWithDpdk()
			testExecProbDenyAtCPUPinning(c, dpdkPods)
			return nil
		}))

	// Restart on reboot label test case
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestRestartOnRebootLabelOnPodsUsingSRIOV)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoSRIOVPodsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			sriovPods, err := env.GetPodsUsingSRIOV()
			if err != nil {
				return fmt.Errorf("failure getting pods using SRIOV: %v", err)
			}
			testRestartOnRebootLabelOnPodsUsingSriov(c, sriovPods)
			return nil
		}))
}

func testExecProbDenyAtCPUPinning(check *checksdb.Check, dpdkPods []*provider.Pod) {
	check.LogInfo("Check if exec probe is happening")
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, cpuPinnedPod := range dpdkPods {
		execProbeFound := false
		for _, cut := range cpuPinnedPod.Containers {
			if cut.HasExecProbes() {
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(cpuPinnedPod.Namespace, cpuPinnedPod.Name, "Exec prob is not allowed", false))
				execProbeFound = true
			}
		}

		if !execProbeFound {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(cpuPinnedPod.Namespace, cpuPinnedPod.Name, "Exec prob is allowed", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

//nolint:funlen
func testUndeclaredContainerPortsUsage(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	var portInfo netutil.PortInfo
	for _, put := range env.Pods {
		// First get the ports declared in the Pod's containers spec
		declaredPorts := make(map[netutil.PortInfo]bool)
		for _, cut := range put.Containers {
			for _, port := range cut.Ports {
				portInfo.PortNumber = int(port.ContainerPort)
				portInfo.Protocol = string(port.Protocol)
				declaredPorts[portInfo] = true
			}
		}

		// Then check the actual ports that the containers are listening on
		firstPodContainer := put.Containers[0]
		listeningPorts, err := netutil.GetListeningPorts(firstPodContainer)
		if err != nil {
			check.LogDebug("Failed to get the container's listening ports, err: %v", err)
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewPodReportObject(put.Namespace, put.Name, fmt.Sprintf("Failed to get the container's listening ports, err: %v", err), false))
			continue
		}
		if len(listeningPorts) == 0 {
			check.LogDebug("None of the containers of %s have any listening port.", put)
			continue
		}

		// Verify that all the listening ports have been declared in the container spec
		failedPod := false
		for listeningPort := range listeningPorts {
			if put.ContainsIstioProxy() && netcommons.ReservedIstioPorts[int32(listeningPort.PortNumber)] {
				check.LogDebug("%s is listening on port %d protocol %s, but the pod also contains istio-proxy. Ignoring.",
					put, listeningPort.PortNumber, listeningPort.Protocol)
				continue
			}

			if ok := declaredPorts[listeningPort]; !ok {
				check.LogDebug("%s is listening on port %d protocol %s, but that port was not declared in any container spec.",
					put, listeningPort.PortNumber, listeningPort.Protocol)
				failedPod = true
				nonCompliantObjects = append(nonCompliantObjects,
					testhelper.NewPodReportObject(put.Namespace, put.Name,
						"Listening port was declared in no container spec", false).
						SetType(testhelper.ListeningPortType).
						AddField(testhelper.PortNumber, strconv.Itoa(listeningPort.PortNumber)).
						AddField(testhelper.PortProtocol, listeningPort.Protocol))
			} else {
				compliantObjects = append(compliantObjects,
					testhelper.NewPodReportObject(put.Namespace, put.Name,
						"Listening port was declared in container spec", true).
						SetType(testhelper.ListeningPortType).
						AddField(testhelper.PortNumber, strconv.Itoa(listeningPort.PortNumber)).
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

// testDefaultNetworkConnectivity test the connectivity between the default interfaces of containers under test
func testNetworkConnectivity(env *provider.TestEnvironment, aIPVersion netcommons.IPVersion, aType netcommons.IFType, check *checksdb.Check) {
	netsUnderTest, claimsLog := icmp.BuildNetTestContext(env.Pods, aIPVersion, aType)
	// Saving  curated logs to claims file
	check.LogDebug("%s", claimsLog.GetLogLines())
	report, claimsLog, skip := icmp.RunNetworkingTests(netsUnderTest, defaultNumPings, aIPVersion)
	// Saving curated logs to claims file
	check.LogDebug("%s", claimsLog.GetLogLines())
	if skip {
		check.LogInfo("There are no %s networks to test with at least 2 pods, skipping test", aIPVersion)
	}
	check.SetResult(report.CompliantObjectsOut, report.NonCompliantObjectsOut)
}

func testOCPReservedPortsUsage(check *checksdb.Check, env *provider.TestEnvironment) {
	// List of all ports reserved by OpenShift
	OCPReservedPorts := map[int32]bool{
		22623: true,
		22624: true}
	compliantObjects, nonCompliantObjects := netcommons.TestReservedPortsUsage(env, OCPReservedPorts, "OCP")
	check.SetResult(compliantObjects, nonCompliantObjects)
}

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
	compliantObjects, nonCompliantObjects := netcommons.TestReservedPortsUsage(env, ReservedPorts, "Partner")
	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testDualStackServices(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	check.LogInfo("Testing services (should be either single stack ipv6 or dual-stack)")
	for _, s := range env.Services {
		serviceIPVersion, err := services.GetServiceIPVersion(s)
		if err != nil {
			check.LogDebug("%s", err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject("Could not get IP Version from service", testhelper.ServiceType, false).
				AddField(testhelper.Namespace, s.Namespace).
				AddField(testhelper.ServiceName, s.Name))
		}
		if serviceIPVersion == netcommons.Undefined || serviceIPVersion == netcommons.IPv4 {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject("Service supports only IPv4", testhelper.ServiceType, false).
				AddField(testhelper.Namespace, s.Namespace).
				AddField(testhelper.ServiceName, s.Name).
				AddField(testhelper.ServiceIPVersion, serviceIPVersion.String()))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewReportObject("Service support IPv6 or is dual stack", testhelper.ServiceType, false).
				AddField(testhelper.Namespace, s.Namespace).
				AddField(testhelper.ServiceName, s.Name).
				AddField(testhelper.ServiceIPVersion, serviceIPVersion.String()))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testNetworkPolicyDenyAll(check *checksdb.Check, env *provider.TestEnvironment) {
	check.LogInfo("Test for Deny All in network policies")
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	// Loop through the pods, looking for corresponding entries within a deny-all network policy (both ingress and egress).
	// This ensures that each pod is accounted for that we are tasked with testing and excludes any pods that are not marked
	// for testing (via the labels).
	for _, put := range env.Pods {
		denyAllEgressFound := false
		denyAllIngressFound := false

		// Look through all of the network policies for a matching namespace.
		for index := range env.NetworkPolicies {
			check.LogDebug("Testing network policy %s against pod %s", env.NetworkPolicies[index].Name, put.String())

			// Skip any network policies that don't match the namespace of the pod we are testing.
			if env.NetworkPolicies[index].Namespace != put.Namespace {
				continue
			}

			// Match the pod namespace with the network policy namespace.
			if policies.LabelsMatch(env.NetworkPolicies[index].Spec.PodSelector, put.Labels) {
				if !denyAllEgressFound {
					denyAllEgressFound = policies.IsNetworkPolicyCompliant(&env.NetworkPolicies[index], networkingv1.PolicyTypeEgress)
				}
				if !denyAllIngressFound {
					denyAllIngressFound = policies.IsNetworkPolicyCompliant(&env.NetworkPolicies[index], networkingv1.PolicyTypeIngress)
				}
			}
		}

		// Network policy has not been found that contains a deny-all rule for both ingress and egress.
		podIsCompliant := true
		if !denyAllIngressFound {
			check.LogDebug("%s was found to not have a default ingress deny-all network policy.", put.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod was found to not have a default ingress deny-all network policy", false))
			podIsCompliant = false
		}

		if !denyAllEgressFound {
			check.LogDebug("%s was found to not have a default egress deny-all network policy.", put.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod was found to not have a default egress deny-all network policy", false))
			podIsCompliant = false
		}

		if podIsCompliant {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has a default ingress/egress deny-all network policy", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testRestartOnRebootLabelOnPodsUsingSriov(check *checksdb.Check, sriovPods []*provider.Pod) {
	const (
		restartOnRebootLabel = "restart-on-reboot"
	)

	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, pod := range sriovPods {
		check.LogDebug("Pod %s uses SRIOV network/s. Checking label %s existence & value.", pod, restartOnRebootLabel)

		labelValue, exist := pod.GetLabels()[restartOnRebootLabel]
		if !exist {
			check.LogDebug("Pod %s is using SRIOV but the label %s was not found.", pod, restartOnRebootLabel)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(pod.Namespace, pod.Name, fmt.Sprintf("Pod uses SRIOV but the label %s was not found", restartOnRebootLabel), false))
			continue
		}

		if labelValue != "true" {
			check.LogDebug("Pod %s is using SRIOV but the %s label value is not true.", pod, restartOnRebootLabel)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(pod.Namespace, pod.Name, fmt.Sprintf("Pod uses SRIOV but the label %s is not set to true", restartOnRebootLabel), false))
			continue
		}

		compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(pod.Namespace, pod.Name, fmt.Sprintf("Pod uses SRIOV and the label %s is set to true", restartOnRebootLabel), true))
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}
