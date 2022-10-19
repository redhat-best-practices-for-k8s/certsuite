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

package networking

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/icmp"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/netcommons"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/netutil"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/policies"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/services"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
	corev1 "k8s.io/api/core/v1"
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

// All actual test code belongs below here.  Utilities belong above.
var _ = ginkgo.Describe(common.NetworkingTestKey, func() {
	logrus.Debugf("Entering %s suite", common.NetworkingTestKey)

	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})
	ginkgo.ReportAfterEach(results.RecordResult)
	// Default interface ICMP IPv4 test case
	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestICMPv4ConnectivityIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers, env.Pods)
		testNetworkConnectivity(&env, netcommons.IPv4, netcommons.DEFAULT)
	})
	// Multus interfaces ICMP IPv4 test case
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestICMPv4ConnectivityMultusIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers, env.Pods)
		testNetworkConnectivity(&env, netcommons.IPv4, netcommons.MULTUS)
	})
	// Default interface ICMP IPv6 test case
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestICMPv6ConnectivityIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers, env.Pods)
		testNetworkConnectivity(&env, netcommons.IPv6, netcommons.DEFAULT)
	})
	// Multus interfaces ICMP IPv6 test case
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestICMPv6ConnectivityMultusIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers, env.Pods)
		testNetworkConnectivity(&env, netcommons.IPv6, netcommons.MULTUS)
	})
	// Default interface ICMP IPv6 test case
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestUndeclaredContainerPortsUsage)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers, env.Pods)
		testUndeclaredContainerPortsUsage(&env)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestServicesDoNotUseNodeportsIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers, env.Pods)
		testNodePort(&env)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestOCPReservedPortsUsage)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers, env.Pods)
		testOCPReservedPortsUsage(&env)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestServiceDualStackIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Services)
		testDualStackServices(&env)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNFTablesIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		testIsNFTablesConfigPresent(&env)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestIPTablesIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Containers)
		testIsIPTablesConfigPresent(&env)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNetworkPolicyDenyAllIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		testNetworkPolicyDenyAll(&env)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestReservedExtendedPartnerPorts)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		testPartnerSpecificTCPPorts(&env)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestDpdkCPUPinningExecProbe)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		dpdkPods := env.GetCPUPinningPodsWithDpdk()
		testhelper.SkipIfEmptyAny(ginkgo.Skip, dpdkPods)
		testExecProbDenyAtCPUPinning(dpdkPods)
	})
})

func testExecProbDenyAtCPUPinning(dpdkPods []*provider.Pod) {
	ginkgo.By("Check if exec probe is happening")

	for _, cpuPinnedPod := range dpdkPods {
		for _, x := range cpuPinnedPod.Containers {
			if x.LivenessProbe != nil || x.StartupProbe != nil || x.ReadinessProbe != nil {
				ginkgo.Fail("Exec prob is not allowed")
			}
		}
	}
}

//nolint:funlen
func testUndeclaredContainerPortsUsage(env *provider.TestEnvironment) {
	var portInfo netutil.PortInfo
	var failedPods []*provider.Pod
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
			tnf.ClaimFilePrintf("Failed to get the container's listening ports, err: %s", err)
			failedPods = append(failedPods, put)
			continue
		}
		if len(listeningPorts) == 0 {
			tnf.ClaimFilePrintf("None of the containers of %s have any listening port.", put)
			continue
		}

		// Verify that all the listening ports have been declared in the container spec
		failedPod := false
		for listeningPort := range listeningPorts {
			if !declaredPorts[listeningPort] {
				tnf.ClaimFilePrintf("%s is listening on port %d protocol %s, but that port was not declared in any container spec.", put, listeningPort.PortNumber, listeningPort.Protocol)
				failedPod = true
			}
		}
		if failedPod {
			failedPods = append(failedPods, put)
		}
	}
	testhelper.AddTestResultLog("Non-compliant", failedPods, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testNodePort(env *provider.TestEnvironment) {
	badServices := []string{}
	for _, s := range env.Services {
		ginkgo.By(fmt.Sprintf("Testing %s", services.ToString(s)))

		if s.Spec.Type == nodePort {
			tnf.ClaimFilePrintf("FAILURE: Service %s (ns %s) type is nodePort", s.Name, s.Namespace)
			badServices = append(badServices, fmt.Sprintf("ns: %s, name: %s", s.Namespace, s.Name))
		}
	}
	testhelper.AddTestResultLog("Non-compliant", badServices, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testDefaultNetworkConnectivity test the connectivity between the default interfaces of containers under test
func testNetworkConnectivity(env *provider.TestEnvironment, aIPVersion netcommons.IPVersion, aType netcommons.IFType) {
	netsUnderTest, claimsLog := icmp.BuildNetTestContext(env.Pods, aIPVersion, aType)
	// Saving  curated logs to claims file
	tnf.ClaimFilePrintf("%s", claimsLog.GetLogLines())
	badNets, claimsLog, skip := icmp.RunNetworkingTests(netsUnderTest, defaultNumPings, aIPVersion)
	// Saving curated logs to claims file
	tnf.ClaimFilePrintf("%s", claimsLog.GetLogLines())
	if skip {
		ginkgo.Skip(fmt.Sprintf("There are no %s networks to test, skipping test", aIPVersion))
	}
	if n := len(badNets); n > 0 {
		logrus.Debugf("Failed nets: %+v", badNets)
		ginkgo.Fail(fmt.Sprintf("%d nets failed the %s network %s ping test.", n, aType, aIPVersion))
	}
}

func testOCPReservedPortsUsage(env *provider.TestEnvironment) {
	// List of all ports reserved by OpenShift
	OCPReservedPorts := map[int32]bool{22623: true, 22624: true}

	rogueContainers := netcommons.FindRogueContainersDeclaringPorts(env.Containers, OCPReservedPorts)
	roguePods, failedContainers := netcommons.FindRoguePodsListeningToPorts(env.Pods, OCPReservedPorts)

	if n := len(rogueContainers); n > 0 {
		errMsg := fmt.Sprintf("Number of containers declaring ports reserved by OpenShift: %d", n)
		tnf.ClaimFilePrintf(errMsg)
		ginkgo.Fail(errMsg)
	}

	if n := len(roguePods); n > 0 {
		errMsg := fmt.Sprintf("Number of pods having one or more containers listening on ports reserved by OpenShift: %d", n)
		tnf.ClaimFilePrintf(errMsg)
		ginkgo.Fail(errMsg)
	}

	if failedContainers > 0 {
		errMsg := fmt.Sprintf("Number of containers in which the test could not be performed due to an error: %d", failedContainers)
		tnf.ClaimFilePrintf(errMsg)
		ginkgo.Fail(errMsg)
	}
}

func testPartnerSpecificTCPPorts(env *provider.TestEnvironment) {
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

	rogueContainers := netcommons.FindRogueContainersDeclaringPorts(env.Containers, ReservedPorts)
	roguePods, failedContainers := netcommons.FindRoguePodsListeningToPorts(env.Pods, ReservedPorts)

	if n := len(rogueContainers); n > 0 {
		errMsg := fmt.Sprintf("Number of containers declaring ports reserved by Partner: %d", n)
		tnf.ClaimFilePrintf(errMsg)
		ginkgo.Fail(errMsg)
	}

	if n := len(roguePods); n > 0 {
		errMsg := fmt.Sprintf("Number of pods having one or more containers listening on ports reserved by Partner: %d", n)
		tnf.ClaimFilePrintf(errMsg)
		ginkgo.Fail(errMsg)
	}

	if failedContainers > 0 {
		errMsg := fmt.Sprintf("Number of containers in which the test could not be performed due to an error: %d", failedContainers)
		tnf.ClaimFilePrintf(errMsg)
		ginkgo.Fail(errMsg)
	}
}

func testDualStackServices(env *provider.TestEnvironment) {
	var nonCompliantServices []*corev1.Service
	var errorServices []*corev1.Service
	ginkgo.By("Testing services (should be either single stack ipv6 or dual-stack)")
	for _, s := range env.Services {
		result, err := services.GetServiceIPVersion(s)
		if err != nil {
			tnf.ClaimFilePrintf("%s", err)
			errorServices = append(errorServices, s)
		}
		if result == netcommons.Undefined || result == netcommons.IPv4 {
			nonCompliantServices = append(nonCompliantServices, s)
		}
	}
	if len(nonCompliantServices) > 0 {
		tnf.ClaimFilePrintf("Non compliant services:\n %s", services.ToStringSlice(nonCompliantServices))
		ginkgo.Fail(fmt.Sprintf("Found %d non compliant services (either non single stack ipv6 or non dual-stack)", len(nonCompliantServices)))
	}
	if len(errorServices) > 0 {
		tnf.ClaimFilePrintf("Services in error:\n %s", services.ToStringSlice(errorServices))
		ginkgo.Fail(fmt.Sprintf("Found %d services in error, check error log for details", len(errorServices)))
	}
}

const (
	nfTables  = "nftables"
	ipTables  = "iptables"
	ip6Tables = "ip6tables"
)

func testIsConfigPresent(env *provider.TestEnvironment, name string) {
	var badContainers, errorContainers []*provider.Container
	var function func(cut *provider.Container) (bool, string, error)
	switch name {
	case nfTables:
		function = netutil.IsNFTablesPresent
	case ipTables:
		function = netutil.IsIPTablesPresent
	case ip6Tables:
		function = netutil.IsIP6TablesPresent
	default:
		ginkgo.Fail(fmt.Sprintf("Internal error: configuration %s is not supported", name))
	}
	for _, cut := range env.Containers {
		result, log, err := function(cut)
		if err != nil {
			tnf.ClaimFilePrintf("Could not check %s config on: %s, error: %s log: %s", name, cut, err, log)
			errorContainers = append(errorContainers, cut)
			continue
		}
		if result {
			tnf.ClaimFilePrintf("Non-compliant %s config on: %s log: %s", name, cut, log)
			badContainers = append(badContainers, cut)
		}
	}
	testhelper.AddTestResultLog("Non-compliant", badContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
	testhelper.AddTestResultLog("Error", errorContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testIsNFTablesConfigPresent(env *provider.TestEnvironment) {
	testIsConfigPresent(env, nfTables)
}

func testIsIPTablesConfigPresent(env *provider.TestEnvironment) {
	testIsConfigPresent(env, ipTables)
	testIsConfigPresent(env, ip6Tables)
}

func testNetworkPolicyDenyAll(env *provider.TestEnvironment) {
	ginkgo.By("Test for Deny All in network policies")
	var podsMissingDenyAllDefaultPolicies []string

	// Loop through the pods, looking for corresponding entries within a deny-all network policy (both ingress and egress).
	// This ensures that each pod is accounted for that we are tasked with testing and excludes any pods that are not marked
	// for testing (via the labels).
	for _, put := range env.Pods {
		denyAllEgressFound := false
		denyAllIngressFound := false

		// Look through all of the network policies for a matching namespace.
		for index := range env.NetworkPolicies {
			logrus.Debugf("Testing network policy %s against pod %s", env.NetworkPolicies[index].Name, put.String())

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
		if !denyAllIngressFound {
			podsMissingDenyAllDefaultPolicies = append(podsMissingDenyAllDefaultPolicies, put.Name)
			tnf.ClaimFilePrintf("%s was found to not have a default ingress deny-all network policy.", put.Name)
		}

		if !denyAllEgressFound {
			podsMissingDenyAllDefaultPolicies = append(podsMissingDenyAllDefaultPolicies, put.Name)
			tnf.ClaimFilePrintf("%s was found to not have a default egress deny-all network policy.", put.Name)
		}
	}

	testhelper.AddTestResultLog("Non-compliant", podsMissingDenyAllDefaultPolicies, tnf.ClaimFilePrintf, ginkgo.Fail)
}
