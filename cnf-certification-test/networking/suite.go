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

package networking

import (
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/netcommons"

	"github.com/onsi/ginkgo/v2"

	"fmt"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/icmp"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

const (
	defaultNumPings = 5
)

//
// All actual test code belongs below here.  Utilities belong above.
//
var _ = ginkgo.Describe(common.NetworkingTestKey, func() {
	logrus.Debugf("Entering %s suite", common.NetworkingTestKey)

	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
		provider.WaitDebugPodReady()
	})

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
})

// testDefaultNetworkConnectivity test the connectivity between the default interfaces of containers under test
func testDefaultNetworkConnectivity(env *provider.TestEnvironment, count int, aIPVersion netcommons.IPVersion) {
	netsUnderTest := make(map[string]netcommons.NetTestContext)
	for _, put := range env.Pods {
		// The first container is used to get the network namespace
		aContainerInPod := &put.Spec.Containers[0]
		if _, ok := env.SkipNetTests[put]; ok {
			tnf.ClaimFilePrintf("Skipping pod %s because it is excluded from all connectivity tests", put.Name)
			continue
		}
		netKey := "default" //nolint:goconst // only used once
		defaultIPAddress := put.Status.PodIPs

		icmp.ProcessContainerIpsPerNet(env.ContainersMap[aContainerInPod], netKey, netcommons.PodIPsToStringList(defaultIPAddress), netsUnderTest, aIPVersion)
	}
	badNets, claimsLog := icmp.RunNetworkingTests(env, netsUnderTest, count, aIPVersion)

	// Saving all curated logs to claims file
	tnf.ClaimFilePrintf("%s", claimsLog)

	if n := len(badNets); n > 0 {
		logrus.Debugf("Failed nets: %+v", badNets)
		tnf.GinkgoFail(fmt.Sprintf("%d nets failed the default network %s ping test.", n, aIPVersion))
	}
}

// testMultusNetworkConnectivity tests the connectivity between the multus interfaces of the containers under test
func testMultusNetworkConnectivity(env *provider.TestEnvironment, count int, aIPVersion netcommons.IPVersion) {
	netsUnderTest := make(map[string]netcommons.NetTestContext)
	for _, put := range env.Pods {
		// The first container is used to get the network namespace
		aContainerInPod := &put.Spec.Containers[0]

		if _, ok := env.SkipNetTests[put]; ok {
			tnf.ClaimFilePrintf("Skipping pod %s because it is excluded from all connectivity tests", put.Name)
			continue
		}
		if _, ok := env.SkipMultusNetTests[put]; ok {
			tnf.ClaimFilePrintf("Skipping pod %s because it is excluded from multus connectivity tests only", put.Name)
			continue
		}
		for netKey, multusIPAddress := range env.MultusIPs[put] {
			icmp.ProcessContainerIpsPerNet(env.ContainersMap[aContainerInPod], netKey, multusIPAddress, netsUnderTest, aIPVersion)
		}
	}
	badNets, claimsLog := icmp.RunNetworkingTests(env, netsUnderTest, count, aIPVersion)

	// Saving all curated logs to claims file
	tnf.ClaimFilePrintf("%s", claimsLog)

	if n := len(badNets); n > 0 {
		logrus.Debugf("Failed nets: %+v", badNets)
		tnf.GinkgoFail(fmt.Sprintf("%d nets failed the multus %s ping test.", n, aIPVersion))
	}
}
