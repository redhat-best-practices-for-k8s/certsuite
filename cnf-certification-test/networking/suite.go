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

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/icmp"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/netcommons"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/netutil"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/policies"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/services"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
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

	// Default interface ICMP IPv4 test case
	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestICMPv4ConnectivityIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"), testhelper.NewSkipObject(env.Pods, "env.Pods"))
		if env.DaemonsetFailedToSpawn {
			ginkgo.Skip("Debug Daemonset failed to spawn skipping testNetworkConnectivity ICMP IPv4")
		}
		testNetworkConnectivity(&env, netcommons.IPv4, netcommons.DEFAULT)
	})
	// Multus interfaces ICMP IPv4 test case
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestICMPv4ConnectivityMultusIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"), testhelper.NewSkipObject(env.Pods, "env.Pods"))
		if env.DaemonsetFailedToSpawn {
			ginkgo.Skip("Debug Daemonset failed to spawn skipping testNetworkConnectivity Multus IPv4")
		}
		testNetworkConnectivity(&env, netcommons.IPv4, netcommons.MULTUS)
	})
	// Default interface ICMP IPv6 test case
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestICMPv6ConnectivityIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"), testhelper.NewSkipObject(env.Pods, "env.Pods"))
		if env.DaemonsetFailedToSpawn {
			ginkgo.Skip("Debug Daemonset failed to spawn skipping testNetworkConnectivity ICMP IPv6")
		}
		testNetworkConnectivity(&env, netcommons.IPv6, netcommons.DEFAULT)
	})
	// Multus interfaces ICMP IPv6 test case
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestICMPv6ConnectivityMultusIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"), testhelper.NewSkipObject(env.Pods, "env.Pods"))
		if env.DaemonsetFailedToSpawn {
			ginkgo.Skip("Debug Daemonset failed to spawn skipping testNetworkConnectivity Multus IPv6")
		}
		testNetworkConnectivity(&env, netcommons.IPv6, netcommons.MULTUS)
	})
	// Default interface ICMP IPv6 test case
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestUndeclaredContainerPortsUsage)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"), testhelper.NewSkipObject(env.Pods, "env.Pods"))
		if env.DaemonsetFailedToSpawn {
			ginkgo.Skip("Debug Daemonset failed to spawn skipping testUndeclaredContainerPortsUsage")
		}
		testUndeclaredContainerPortsUsage(&env)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestOCPReservedPortsUsage)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"), testhelper.NewSkipObject(env.Pods, "env.Pods"))
		if env.DaemonsetFailedToSpawn {
			ginkgo.Skip("Debug Daemonset failed to spawn skipping testOCPReservedPortsUsage")
		}
		testOCPReservedPortsUsage(&env)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestServiceDualStackIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Services, "env.Services"))
		testDualStackServices(&env)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestNetworkPolicyDenyAllIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"))
		testNetworkPolicyDenyAll(&env)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestReservedExtendedPartnerPorts)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"))
		if env.DaemonsetFailedToSpawn {
			ginkgo.Skip("Debug Daemonset failed to spawn skipping testPartnerSpecificTCPPorts")
		}
		testPartnerSpecificTCPPorts(&env)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestDpdkCPUPinningExecProbe)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		dpdkPods := env.GetCPUPinningPodsWithDpdk()
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(dpdkPods, "dpdkPods"))
		testExecProbDenyAtCPUPinning(dpdkPods)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestRestartOnRebootLabelOnPodsUsingSRIOV)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		sriovPods, err := env.GetPodsUsingSRIOV()
		if err != nil {
			ginkgo.Fail(fmt.Sprintf("Failure getting pods using SRIOV: %v", err))
		}
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(sriovPods, "sriovPods"))
		testRestartOnRebootLabelOnPodsUsingSriov(sriovPods)
	})
})

func testExecProbDenyAtCPUPinning(dpdkPods []*provider.Pod) {
	ginkgo.By("Check if exec probe is happening")

	for _, cpuPinnedPod := range dpdkPods {
		for _, cut := range cpuPinnedPod.Containers {
			if cut.HasExecProbes() {
				ginkgo.Fail("Exec prob is not allowed")
			}
		}
	}
}

//nolint:funlen
func testUndeclaredContainerPortsUsage(env *provider.TestEnvironment) {
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
			tnf.ClaimFilePrintf("Failed to get the container's listening ports, err: %v", err)
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewPodReportObject(put.Namespace, put.Name, fmt.Sprintf("Failed to get the container's listening ports, err: %v", err), false))
			continue
		}
		if len(listeningPorts) == 0 {
			tnf.ClaimFilePrintf("None of the containers of %s have any listening port.", put)
			continue
		}

		// Verify that all the listening ports have been declared in the container spec
		failedPod := false
		for listeningPort := range listeningPorts {
			if put.ContainsIstioProxy() && netcommons.ReservedIstioPorts[int32(listeningPort.PortNumber)] {
				tnf.ClaimFilePrintf("%s is listening on port %d protocol %s, but the pod also contains istio-proxy. Ignoring.",
					put, listeningPort.PortNumber, listeningPort.Protocol)
				continue
			}

			if ok := declaredPorts[listeningPort]; !ok {
				tnf.ClaimFilePrintf("%s is listening on port %d protocol %s, but that port was not declared in any container spec.",
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
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testDefaultNetworkConnectivity test the connectivity between the default interfaces of containers under test
func testNetworkConnectivity(env *provider.TestEnvironment, aIPVersion netcommons.IPVersion, aType netcommons.IFType) {
	netsUnderTest, claimsLog := icmp.BuildNetTestContext(env.Pods, aIPVersion, aType)
	// Saving  curated logs to claims file
	tnf.ClaimFilePrintf("%s", claimsLog.GetLogLines())
	report, claimsLog, skip := icmp.RunNetworkingTests(netsUnderTest, defaultNumPings, aIPVersion)
	// Saving curated logs to claims file
	tnf.ClaimFilePrintf("%s", claimsLog.GetLogLines())
	if skip {
		ginkgo.Skip(fmt.Sprintf("There are no %s networks to test with at least 2 pods, skipping test", aIPVersion))
	}
	testhelper.AddTestResultReason(report.CompliantObjectsOut, report.NonCompliantObjectsOut, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testOCPReservedPortsUsage(env *provider.TestEnvironment) {
	// List of all ports reserved by OpenShift
	OCPReservedPorts := map[int32]bool{
		22623: true,
		22624: true}
	compliantObjects, nonCompliantObjects := netcommons.TestReservedPortsUsage(env, OCPReservedPorts, "OCP")
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
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
	compliantObjects, nonCompliantObjects := netcommons.TestReservedPortsUsage(env, ReservedPorts, "Partner")
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testDualStackServices(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	ginkgo.By("Testing services (should be either single stack ipv6 or dual-stack)")
	for _, s := range env.Services {
		serviceIPVersion, err := services.GetServiceIPVersion(s)
		if err != nil {
			tnf.ClaimFilePrintf("%s", err)
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

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testNetworkPolicyDenyAll(env *provider.TestEnvironment) {
	ginkgo.By("Test for Deny All in network policies")
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
		podIsCompliant := true
		if !denyAllIngressFound {
			tnf.ClaimFilePrintf("%s was found to not have a default ingress deny-all network policy.", put.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod was found to not have a default ingress deny-all network policy", false))
			podIsCompliant = false
		}

		if !denyAllEgressFound {
			tnf.ClaimFilePrintf("%s was found to not have a default egress deny-all network policy.", put.Name)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod was found to not have a default egress deny-all network policy", false))
			podIsCompliant = false
		}

		if podIsCompliant {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has a default ingress/egress deny-all network policy", true))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testRestartOnRebootLabelOnPodsUsingSriov(sriovPods []*provider.Pod) {
	const (
		restartOnRebootLabel = "restart-on-reboot"
	)

	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, pod := range sriovPods {
		logrus.Debugf("Pod %s uses SRIOV network/s. Checking label %s existence & value.", pod, restartOnRebootLabel)

		labelValue, exist := pod.GetLabels()[restartOnRebootLabel]
		if !exist {
			tnf.ClaimFilePrintf("Pod %s is using SRIOV but the label %s was not found.", pod, restartOnRebootLabel)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(pod.Namespace, pod.Name, fmt.Sprintf("Pod uses SRIOV but the label %s was not found", restartOnRebootLabel), false))
			continue
		}

		if labelValue != "true" {
			tnf.ClaimFilePrintf("Pod %s is using SRIOV but the %s label value is not true.", pod, restartOnRebootLabel)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(pod.Namespace, pod.Name, fmt.Sprintf("Pod uses SRIOV but the label %s is not set to true", restartOnRebootLabel), false))
			continue
		}

		compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(pod.Namespace, pod.Name, fmt.Sprintf("Pod uses SRIOV and the label %s is set to true", restartOnRebootLabel), true))
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}
