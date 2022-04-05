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

package declaredandlistening

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/declaredandlistening"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/icmp"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/netcommons"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/internal/crclient"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

const (
	defaultNumPings = 5
	cmd             = `ss -tulwnH`
)

type Port []struct {
	ContainerPort int
	Name          string
	Protocol      string
}

//
// All actual test code belongs below here.  Utilities belong above.
//
var _ = ginkgo.Describe(common.NetworkingTestKey, func() {
	logrus.Debugf("Entering %s suite", common.NetworkingTestKey)

	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})
	ginkgo.ReportAfterEach(results.RecordResult)
	// Default interface ICMP IPv4 test case
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestICMPv4ConnectivityIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testDefaultNetworkConnectivity(&env, defaultNumPings, netcommons.IPv4)
	})
	// Multus interfaces ICMP IPv4 test case
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestICMPv4ConnectivityMultusIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testMultusNetworkConnectivity(&env, defaultNumPings, netcommons.IPv4)
	})
	// Default interface ICMP IPv6 test case
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestICMPv6ConnectivityIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testDefaultNetworkConnectivity(&env, defaultNumPings, netcommons.IPv6)
	})
	// Multus interfaces ICMP IPv6 test case
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestICMPv6ConnectivityMultusIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testMultusNetworkConnectivity(&env, defaultNumPings, netcommons.IPv6)
	})
	// Default interface ICMP IPv6 test case
	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestUndeclaredContainerPortsUsage)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testListenAndDeclared(&env)
	})
})

//nolint:funlen
func testListenAndDeclared(env *provider.TestEnvironment) {
	var k declaredandlistening.Key
	var failedPods []*provider.Pod
	for _, podUnderTest := range env.Pods {
		declaredPorts := make(map[declaredandlistening.Key]bool)
		listeningPorts := make(map[declaredandlistening.Key]bool)
		for _, cut := range podUnderTest.Containers {
			ports := cut.Data.Ports
			logrus.Debugf("%s declaredPorts: %v", podUnderTest, ports)
			for j := 0; j < len(ports); j++ {
				k.Port = int(ports[j].ContainerPort)
				k.Protocol = string(ports[j].Protocol)
				declaredPorts[k] = true
			}
		}
		firstPodContainer := podUnderTest.Containers[0]
		outStr, errStr, err := crclient.ExecCommandContainerNSEnter(cmd, firstPodContainer, env)
		if err != nil || errStr != "" {
			tnf.ClaimFilePrintf("Failed to execute command %s on %s, err: %s, errStr: %s", cmd, firstPodContainer, err, errStr)
			failedPods = append(failedPods, podUnderTest)
			continue
		}
		declaredandlistening.ParseListening(outStr, listeningPorts)
		if len(listeningPorts) == 0 {
			tnf.ClaimFilePrintf("None of the containers of %s have any listening port.", podUnderTest)
			continue
		}
		// compare between declaredPort,listeningPort
		undeclaredPorts := declaredandlistening.CheckIfListenIsDeclared(listeningPorts, declaredPorts)
		for k := range undeclaredPorts {
			tnf.ClaimFilePrintf("%s is listening on port %d protocol %d, but that port was not declared in any container spec.", podUnderTest, k.Port, k.Protocol)
		}
		if len(undeclaredPorts) != 0 {
			failedPods = append(failedPods, podUnderTest)
		}
	}
	if nf := len(failedPods); nf > 0 {
		ginkgo.Fail(fmt.Sprintf("Found %d pods with listening ports not declared", nf))
	}
}

// testDefaultNetworkConnectivity test the connectivity between the default interfaces of containers under test
func testDefaultNetworkConnectivity(env *provider.TestEnvironment, count int, aIPVersion netcommons.IPVersion) {
	netsUnderTest := make(map[string]netcommons.NetTestContext)
	for _, put := range env.Pods {
		if put.SkipNetTests {
			tnf.ClaimFilePrintf("Skipping pod %s because it is excluded from all connectivity tests", put.Data.Name)
			continue
		}
		netKey := "default" //nolint:goconst // only used once
		defaultIPAddress := put.Data.Status.PodIPs
		// The first container is used to get the network namespace
		icmp.ProcessContainerIpsPerNet(put.Containers[0], netKey, netcommons.PodIPsToStringList(defaultIPAddress), netsUnderTest, aIPVersion)
	}
	badNets, claimsLog := icmp.RunNetworkingTests(env, netsUnderTest, count, aIPVersion)

	// Saving all curated logs to claims file
	tnf.ClaimFilePrintf("%s", claimsLog)

	if n := len(badNets); n > 0 {
		logrus.Debugf("Failed nets: %+v", badNets)
		ginkgo.Fail(fmt.Sprintf("%d nets failed the default network %s ping test.", n, aIPVersion))
	}
}

// testMultusNetworkConnectivity tests the connectivity between the multus interfaces of the containers under test
func testMultusNetworkConnectivity(env *provider.TestEnvironment, count int, aIPVersion netcommons.IPVersion) {
	netsUnderTest := make(map[string]netcommons.NetTestContext)
	for _, put := range env.Pods {
		if put.SkipNetTests {
			tnf.ClaimFilePrintf("Skipping pod %s because it is excluded from all connectivity tests", put.Data.Name)
			continue
		}
		if put.SkipMultusNetTests {
			tnf.ClaimFilePrintf("Skipping pod %s because it is excluded from multus connectivity tests only", put.Data.Name)
			continue
		}
		for netKey, multusIPAddress := range put.MultusIPs {
			// The first container is used to get the network namespace
			icmp.ProcessContainerIpsPerNet(put.Containers[0], netKey, multusIPAddress, netsUnderTest, aIPVersion)
		}
	}
	badNets, claimsLog := icmp.RunNetworkingTests(env, netsUnderTest, count, aIPVersion)

	// Saving all curated logs to claims file
	tnf.ClaimFilePrintf("%s", claimsLog)

	if n := len(badNets); n > 0 {
		logrus.Debugf("Failed nets: %+v", badNets)
		ginkgo.Fail(fmt.Sprintf("%d nets failed the multus %s ping test.", n, aIPVersion))
	}
}
