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

package icmp

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/crclient"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/networking/netcommons"
)

const (
	// ConnectInvalidArgumentRegex is a regex which matches when an invalid IP address or hostname is provided as input.
	ConnectInvalidArgumentRegex = `(?m)connect: Invalid argument$`
	// SuccessfulOutputRegex matches a successfully run "ping" command.  That does not mean that no errors or drops
	// occurred during the test.
	SuccessfulOutputRegex = `(?m)(\d+) packets transmitted, (\d+)( packets){0,1} received, (?:\+(\d+) errors)?.*$`
)

// PingResults holds the statistics of a ping command execution.
//
// It stores the number of transmitted packets, the number received,
// the outcome status code, and any errors encountered during the
// ping operation. The String method formats these values into a human‑readable string.
type PingResults struct {
	outcome     int
	transmitted int
	received    int
	errors      int
}

// String returns a human‑readable summary of the ping test results.
//
// It formats the data contained in a PingResults instance into a single
// string, including metrics such as packet loss and latency.
// The returned string is suitable for logging or displaying to users.
func (results PingResults) String() string {
	return fmt.Sprintf("outcome: %s transmitted: %d received: %d errors: %d", testhelper.ResultToString(results.outcome), results.transmitted, results.received, results.errors)
}

// BuildNetTestContext creates a map of network test contexts for the given pods.
//
// It accepts a slice of pod pointers, an IP version indicator, an interface type,
// and a logger. For each pod it generates a NetTestContext containing
// information about container IPs, interface configurations, and other
// networking parameters needed for ICMP tests. The function logs progress
// messages and returns the constructed map keyed by pod identifiers.
func BuildNetTestContext(pods []*provider.Pod, aIPVersion netcommons.IPVersion, aType netcommons.IFType, logger *log.Logger) (netsUnderTest map[string]netcommons.NetTestContext) {
	netsUnderTest = make(map[string]netcommons.NetTestContext)
	for _, put := range pods {
		logger.Info("Testing Pod %q", put)
		if put.SkipNetTests {
			logger.Info("Skipping %q because it is excluded from all connectivity tests", put)
			continue
		}

		if aType == netcommons.MULTUS {
			if put.SkipMultusNetTests {
				logger.Info("Skipping pod %q because it is excluded from %q connectivity tests only", put.Name, aType)
				continue
			}
			for netKey, multusNetworkInterface := range put.MultusNetworkInterfaces {
				// The first container is used to get the network namespace
				processContainerIpsPerNet(put.Containers[0], netKey, multusNetworkInterface.IPs, multusNetworkInterface.Interface, netsUnderTest, aIPVersion, logger)
			}
			continue
		}

		const defaultNetKey = "default"
		defaultIPAddress := put.Status.PodIPs
		// The first container is used to get the network namespace
		processContainerIpsPerNet(put.Containers[0], defaultNetKey, netcommons.PodIPsToStringList(defaultIPAddress), "", netsUnderTest, aIPVersion, logger)
	}
	return netsUnderTest
}

// processContainerIpsPerNet processes container IP addresses for a given network attachment and uses them as test targets.
//
// It selects the first container in the loop as the test initiator and uses that container's OC context to initiate ping tests.
// The function filters IPs by version, logs debug information, and returns a closure that performs the actual test when invoked.
func processContainerIpsPerNet(containerID *provider.Container,
	netKey string,
	ipAddresses []string,
	ifName string,
	netsUnderTest map[string]netcommons.NetTestContext,
	aIPVersion netcommons.IPVersion,
	logger *log.Logger) {
	ipAddressesFiltered := netcommons.FilterIPListByIPVersion(ipAddresses, aIPVersion)
	if len(ipAddressesFiltered) == 0 {
		// if no multus addresses found, skip this container
		logger.Debug("Skipping %q, Network %q because no multus IPs are present", containerID, netKey)
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
		logger.Debug("%q selected to initiate ping tests", containerID)
		entry.TesterSource.ContainerIdentifier = containerID
		// if multiple interfaces are present for this network on this container/pod, pick the first one as the tester source ip
		entry.TesterSource.IP = ipAddressesFiltered[firstIPIndex]
		if ifName != "" {
			entry.TesterSource.InterfaceName = ifName
		}
		// do no include tester's IP in the list of destination IPs to ping
		firstIPIndex++
	}

	for _, aIP := range ipAddressesFiltered[firstIPIndex:] {
		ipDestEntry := netcommons.ContainerIP{}
		ipDestEntry.ContainerIdentifier = containerID
		ipDestEntry.IP = aIP
		// if the interface name is not empty, then add it to the destination entry
		if ifName != "" {
			ipDestEntry.InterfaceName = ifName
		}
		entry.DestTargets = append(entry.DestTargets, ipDestEntry)
	}

	// Then reassign map entry
	netsUnderTest[netKey] = entry
}

// RunNetworkingTests executes ping tests across multiple network attachments and reports failures.
//
// It receives a map of NetTestContext values keyed by network name, the number of retries,
// an IP version selector, and a logger. For each network context it runs a series of
// ping tests against target IPs, collecting any that fail. The function returns a map
// from network names to slices of failed target IP addresses along with a boolean
// indicating whether any failures were encountered. This allows callers to quickly
// identify which networks or targets are problematic and to generate detailed reports.
func RunNetworkingTests( //nolint:funlen
	netsUnderTest map[string]netcommons.NetTestContext,
	count int,
	aIPVersion netcommons.IPVersion,
	logger *log.Logger) (report testhelper.FailureReasonOut, skip bool) {
	logger.Debug("%s", netcommons.PrintNetTestContextMap(netsUnderTest))
	skip = false
	if len(netsUnderTest) == 0 {
		logger.Debug("There are no %q networks to test, skipping test", aIPVersion)
		skip = true
		return report, skip
	}
	// if no network can be tested, then we need to skip the test entirely.
	// If at least one network can be tested (e.g. > 2 IPs/ interfaces present), then we do not skip the test
	atLeastOneNetworkTested := false
	compliantNets := map[string]int{}
	nonCompliantNets := map[string]int{}
	for netName, netUnderTest := range netsUnderTest {
		compliantNets[netName] = 0
		nonCompliantNets[netName] = 0
		if len(netUnderTest.DestTargets) == 0 {
			logger.Debug("There are no containers to ping for %q network %q. A minimum of 2 containers is needed to run a ping test (a source and a destination) Skipping test", aIPVersion, netName)
			continue
		}
		atLeastOneNetworkTested = true
		logger.Debug("%q Ping tests on network %q. Number of target IPs: %d", aIPVersion, netName, len(netUnderTest.DestTargets))

		for _, aDestIP := range netUnderTest.DestTargets {
			logger.Debug("%q ping test on network %q from ( %q  srcip: %q ) to ( %q dstip: %q )",
				aIPVersion, netName,
				netUnderTest.TesterSource.ContainerIdentifier, netUnderTest.TesterSource.IP,
				aDestIP.ContainerIdentifier, aDestIP.IP)
			result, err := TestPing(netUnderTest.TesterSource.ContainerIdentifier, aDestIP, count)
			logger.Debug("Ping results: %q", result)
			logger.Info("%q ping test on network %q from ( %q  srcip: %q ) to ( %q dstip: %q ) result: %q",
				aIPVersion, netName,
				netUnderTest.TesterSource.ContainerIdentifier, netUnderTest.TesterSource.IP,
				aDestIP.ContainerIdentifier, aDestIP.IP, result)
			if err != nil {
				logger.Debug("Ping failed, err=%v", err)
			}
			if result.outcome != testhelper.SUCCESS {
				logger.Error("Ping from %q (srcip: %q) to %q (dstip: %q) failed",
					netUnderTest.TesterSource.ContainerIdentifier,
					netUnderTest.TesterSource.IP,
					aDestIP.ContainerIdentifier,
					aDestIP.IP)
				nonCompliantNets[netName]++
				nonCompliantObject := testhelper.NewContainerReportObject(netUnderTest.TesterSource.ContainerIdentifier.Namespace,
					netUnderTest.TesterSource.ContainerIdentifier.Podname,
					netUnderTest.TesterSource.ContainerIdentifier.Name, "Pinging destination container/IP from source container (identified by Namespace/Pod Name/Container Name) Failed", false).
					SetType(testhelper.ICMPResultType).
					AddField(testhelper.NetworkName, netName).
					AddField(testhelper.SourceIP, netUnderTest.TesterSource.IP).
					AddField(testhelper.DestinationNamespace, aDestIP.ContainerIdentifier.Namespace).
					AddField(testhelper.DestinationPodName, aDestIP.ContainerIdentifier.Podname).
					AddField(testhelper.DestinationContainerName, aDestIP.ContainerIdentifier.Name).
					AddField(testhelper.DestinationIP, aDestIP.IP)
				report.NonCompliantObjectsOut = append(report.NonCompliantObjectsOut, nonCompliantObject)
			} else {
				logger.Info("Ping from %q (srcip: %q) to %q (dstip: %q) succeeded",
					netUnderTest.TesterSource.ContainerIdentifier,
					netUnderTest.TesterSource.IP,
					aDestIP.ContainerIdentifier,
					aDestIP.IP)
				compliantNets[netName]++
				CompliantObject := testhelper.NewContainerReportObject(netUnderTest.TesterSource.ContainerIdentifier.Namespace,
					netUnderTest.TesterSource.ContainerIdentifier.Podname,
					netUnderTest.TesterSource.ContainerIdentifier.Name, "Pinging destination container/IP from source container (identified by Namespace/Pod Name/Container Name) Succeeded", true).
					SetType(testhelper.ICMPResultType).
					AddField(testhelper.NetworkName, netName).
					AddField(testhelper.SourceIP, netUnderTest.TesterSource.IP).
					AddField(testhelper.DestinationNamespace, aDestIP.ContainerIdentifier.Namespace).
					AddField(testhelper.DestinationPodName, aDestIP.ContainerIdentifier.Podname).
					AddField(testhelper.DestinationContainerName, aDestIP.ContainerIdentifier.Name).
					AddField(testhelper.DestinationIP, aDestIP.IP)
				report.CompliantObjectsOut = append(report.CompliantObjectsOut, CompliantObject)
			}
		}
		if nonCompliantNets[netName] != 0 {
			logger.Error("ICMP tests failed for %d IP source/destination in this network", nonCompliantNets[netName])
			report.NonCompliantObjectsOut = append(report.NonCompliantObjectsOut, testhelper.NewReportObject(fmt.Sprintf("ICMP tests failed for %d IP source/destination in this network", nonCompliantNets[netName]), testhelper.NetworkType, false).
				AddField(testhelper.NetworkName, netName))
		}
		if compliantNets[netName] != 0 {
			logger.Info("ICMP tests were successful for all %d IP source/destination in this network", compliantNets[netName])
			report.CompliantObjectsOut = append(report.CompliantObjectsOut, testhelper.NewReportObject(fmt.Sprintf("ICMP tests were successful for all %d IP source/destination in this network", compliantNets[netName]), testhelper.NetworkType, true).
				AddField(testhelper.NetworkName, netName))
		}
	}
	if !atLeastOneNetworkTested {
		logger.Debug("There are no %q networks to test, skipping test", aIPVersion)
		skip = true
	}

	return report, skip
}

// TestPing Initiates a ping test between a source container and network (1 ip) and a destination container and network (1 ip)
var TestPing = func(sourceContainerID *provider.Container, targetContainerIP netcommons.ContainerIP, count int) (results PingResults, err error) {
	// Specify the interface to use for the ping test (if any)
	interfaceFlag := fmt.Sprintf("-I %s", targetContainerIP.InterfaceName)
	if targetContainerIP.InterfaceName == "" {
		interfaceFlag = ""
	}
	command := fmt.Sprintf("ping %s -c %d %s", interfaceFlag, count, targetContainerIP.IP)
	stdout, stderr, err := crclient.ExecCommandContainerNSEnter(command, sourceContainerID)
	if err != nil || stderr != "" {
		results.outcome = testhelper.ERROR
		return results, fmt.Errorf("ping failed with stderr:%s err:%s", stderr, err)
	}
	results, err = parsePingResult(stdout, stderr)
	return results, err
}

// parsePingResult extracts ping statistics from raw output strings and returns a PingResults struct or an error.
//
// It takes two string arguments: the standard output of a ping command
// and any error output produced by the command. The function parses
// the outputs using regular expressions to find packet loss,
// round‑trip times, and other metrics. If the parsing succeeds,
// it converts numeric values from strings to integers and populates
// a PingResults structure with these fields. On failure, it returns
// an error describing why the output could not be parsed.
func parsePingResult(stdout, stderr string) (results PingResults, err error) {
	re := regexp.MustCompile(ConnectInvalidArgumentRegex)
	matched := re.FindStringSubmatch(stdout)
	// If we find a error log we fail
	if matched != nil {
		results.outcome = testhelper.ERROR
		return results, fmt.Errorf("ping failed with invalid arguments, stdout: %s, stderr: %s", stdout, stderr)
	}
	re = regexp.MustCompile(SuccessfulOutputRegex)
	matched = re.FindStringSubmatch(stdout)
	// If we do not find a successful log, we fail
	if matched == nil {
		results.outcome = testhelper.FAILURE
		return results, fmt.Errorf("ping output did not match successful regex, stdout: %s, stderr: %s", stdout, stderr)
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
