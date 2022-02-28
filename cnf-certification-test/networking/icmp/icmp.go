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

package icmp

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/netcommons"
	"github.com/test-network-function/cnf-certification-test/internal/crclient"
	"github.com/test-network-function/cnf-certification-test/pkg/loghelper"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
)

const (
	// ConnectInvalidArgumentRegex is a regex which matches when an invalid IP address or hostname is provided as input.
	ConnectInvalidArgumentRegex = `(?m)connect: Invalid argument$`
	// SuccessfulOutputRegex matches a successfully run "ping" command.  That does not mean that no errors or drops
	// occurred during the test.
	SuccessfulOutputRegex = `(?m)(\d+) packets transmitted, (\d+)( packets){0,1} received, (?:\+(\d+) errors)?.*$`
)

type PingResults struct {
	outcome     int
	transmitted int
	received    int
	errors      int
}

func (results PingResults) String() string {
	return fmt.Sprintf("outcome: %s transmitted: %d received: %d errors: %d", testhelper.ResultToString(results.outcome), results.transmitted, results.received, results.errors)
}

// processContainerIpsPerNet takes a container ip addresses for a given network attachment's and uses it as a test target.
// The first container in the loop is selected as the test initiator. the Oc context of the container is used to initiate the pings
func ProcessContainerIpsPerNet(containerID *provider.Container,
	netKey string,
	ipAddresses []string,
	netsUnderTest map[string]netcommons.NetTestContext,
	aIPVersion netcommons.IPVersion) {
	ipAddressesFiltered := netcommons.FilterIPListByIPVersion(ipAddresses, aIPVersion)
	if len(ipAddressesFiltered) == 0 {
		// if no multus addresses found, skip this container
		logrus.Debugf("Skipping %s, Network %s because no multus IPs are present", containerID.StringShort(), netKey)
		return
	}
	// Create an entry at "key" if it is not present
	if _, ok := netsUnderTest[netKey]; !ok {
		netsUnderTest[netKey] = netcommons.NetTestContext{}
	}
	// get a copy of the content
	entry := netsUnderTest[netKey]
	// Then modify the copy
	firstIPIndex := 0
	if entry.TesterSource.ContainerIdentifier == nil {
		logrus.Debugf("%s selected to initiate ping tests", containerID.StringShort())
		entry.TesterSource.ContainerIdentifier = containerID
		// if multiple interfaces are present for this network on this container/pod, pick the first one as the tester source ip
		entry.TesterSource.IP = ipAddressesFiltered[firstIPIndex]
		// do no include tester's IP in the list of destination IPs to ping
		firstIPIndex++
	}

	for _, aIP := range ipAddressesFiltered[firstIPIndex:] {
		ipDestEntry := netcommons.ContainerIP{}
		ipDestEntry.ContainerIdentifier = containerID
		ipDestEntry.IP = aIP
		entry.DestTargets = append(entry.DestTargets, ipDestEntry)
	}

	// Then reassign map entry
	netsUnderTest[netKey] = entry
}

// runNetworkingTests takes a map netcommons.NetTestContext, e.g. one context per network attachment
// and runs pings test with it. Returns a network name to a slice of bad target IPs map.
func RunNetworkingTests(env *provider.TestEnvironment,
	netsUnderTest map[string]netcommons.NetTestContext,
	count int,
	aIPVersion netcommons.IPVersion) (badNets map[string][]string, claimsLog loghelper.CuratedLogLines) {
	logrus.Debugf("%s", netcommons.PrintNetTestContextMap(netsUnderTest))
	if len(netsUnderTest) == 0 {
		logrus.Debugf("There are no %s networks to test, skipping test", aIPVersion)
		return badNets, claimsLog
	}
	// maps a net name to a list of failed destination IPs
	badNets = map[string][]string{}
	// if no network can be tested, then we need to skip the test entirely.
	// If at least one network can be tested (e.g. > 2 IPs/ interfaces present), then we do not skip the test
	atLeastOneNetworkTested := false
	for netName, netUnderTest := range netsUnderTest {
		if len(netUnderTest.DestTargets) == 0 {
			logrus.Debugf("There are no containers to ping for %s network %s. A minimum of 2 containers is needed to run a ping test (a source and a destination) Skipping test", aIPVersion, netName)
			continue
		}
		atLeastOneNetworkTested = true
		logrus.Debugf("%s Ping tests on network %s. Number of target IPs: %d", aIPVersion, netName, len(netUnderTest.DestTargets))
		for _, aDestIP := range netUnderTest.DestTargets {
			logrus.Debugf("%s ping test on network %s from ( %s  srcip: %s ) to ( %s dstip: %s )",
				aIPVersion, netName,
				netUnderTest.TesterSource.ContainerIdentifier.StringShort(), netUnderTest.TesterSource.IP,
				aDestIP.ContainerIdentifier.StringShort(), aDestIP.IP)
			result, err := testPing(env, netUnderTest.TesterSource.ContainerIdentifier, aDestIP, count)
			logrus.Debugf("Ping results: %s", result.String())
			claimsLog = claimsLog.AddLogLine("%s ping test on network %s from ( %s  srcip: %s ) to ( %s dstip: %s ) result: %s",
				aIPVersion, netName,
				netUnderTest.TesterSource.ContainerIdentifier.StringShort(), netUnderTest.TesterSource.IP,
				aDestIP.ContainerIdentifier.StringShort(), aDestIP.IP, result.String())
			if err != nil {
				logrus.Debugf("Ping failed with err:%s", err)
			}
			if result.outcome != testhelper.SUCCESS {
				if failedDestIps, netFound := badNets[netName]; netFound {
					badNets[netName] = append(failedDestIps, aDestIP.IP)
				} else {
					badNets[netName] = []string{aDestIP.IP}
				}
			}
		}
	}
	if !atLeastOneNetworkTested {
		logrus.Debugf("There are no network to test for any %s networks, skipping test", aIPVersion)
	}
	return badNets, claimsLog
}

// testPing Initiates a ping test between a source container and network (1 ip) and a destination container and network (1 ip)
func testPing(env *provider.TestEnvironment, sourceContainerID *provider.Container, targetContainerIP netcommons.ContainerIP, count int) (results PingResults, err error) {
	command := fmt.Sprintf("ping -c %d %s", count, targetContainerIP.IP)
	stdout, stderr, err := crclient.ExecCommandContainerNSEnter(command, sourceContainerID, env)
	if err != nil || stderr != "" {
		results.outcome = testhelper.ERROR
		return results, errors.Errorf("Ping failed with stderr:%s err:%s", stderr, err)
	}
	results, err = parsePingResult(stdout, stderr)
	return results, err
}

func parsePingResult(stdout, stderr string) (results PingResults, err error) {
	re := regexp.MustCompile(ConnectInvalidArgumentRegex)
	matched := re.FindStringSubmatch(stdout)
	// If we find a error log we fail
	if matched != nil {
		results.outcome = testhelper.ERROR
		return results, errors.Errorf("Ping failed with invalid arguments, stdout: %s, stderr: %s", stdout, stderr)
	}
	re = regexp.MustCompile(SuccessfulOutputRegex)
	matched = re.FindStringSubmatch(stdout)
	// If we do not find a successful log, we fail
	if matched == nil {
		results.outcome = testhelper.FAILURE
		return results, errors.Errorf("Ping output did not match successful regex, stdout: %s, stderr: %s", stdout, stderr)
	}
	// Ignore errors in converting matches to decimal integers.
	// Regular expression `stat` is required to underwrite this assumption.
	results.transmitted, _ = strconv.Atoi(matched[1])
	results.received, _ = strconv.Atoi(matched[2])
	results.errors, _ = strconv.Atoi(matched[4])
	switch {
	case results.transmitted == 0 || results.errors > 0:
		results.outcome = testhelper.ERROR
	case results.received > 0 && (results.transmitted-results.received) <= 1:
		results.outcome = testhelper.SUCCESS
	default:
		results.outcome = testhelper.FAILURE
	}
	return results, nil
}
