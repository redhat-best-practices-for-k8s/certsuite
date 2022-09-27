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

package netcommons

import (
	"fmt"
	"net"
	"strings"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/netutil"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
	corev1 "k8s.io/api/core/v1"
)

type IPVersion int
type IFType string

const (
	Undefined IPVersion = iota
	IPv4
	IPv6
	IPv4v6
)

const (
	IPv4String             = "IPv4"
	IPv6String             = "IPv6"
	IPv4v6String           = "IPv4v6"
	UndefinedString        = "undefined"
	MULTUS          IFType = "Multus"
	DEFAULT         IFType = "Default"
)

func (version IPVersion) String() string {
	switch version {
	case IPv4:
		return IPv4String
	case IPv6:
		return IPv6String
	case IPv4v6:
		return IPv4v6String
	case Undefined:
		return UndefinedString
	}
	return UndefinedString
}

// netTestContext this is a data structure describing a network test context for a given subnet (e.g. network attachment)
// The test context defines a tester or test initiator, that is initiating the pings. It is selected randomly (first container in the list)
// It also defines a list of destination ping targets corresponding to the other containers IPs on this subnet
type NetTestContext struct {
	// testerContainerNodeOc session context to access the node running the container selected to initiate tests
	TesterContainerNodeName string
	// testerSource is the container select to initiate the ping tests on this given network
	TesterSource ContainerIP
	// ipDestTargets List of containers to be pinged by the testerSource on this given network
	DestTargets []ContainerIP
}

// containerIP holds a container identification and its IP for networking tests.
type ContainerIP struct {
	// ip address of the target container
	IP string
	// targetContainerIdentifier container identifier including namespace, pod name, container name, node name, and container UID
	ContainerIdentifier *provider.Container
}

// String displays the NetTestContext data structure
func (testContext NetTestContext) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("From initiating container: %s\n", testContext.TesterSource.String()))
	if len(testContext.DestTargets) == 0 {
		sb.WriteString("--> No target containers to test for this network")
	}
	for _, target := range testContext.DestTargets {
		sb.WriteString(fmt.Sprintf("--> To target container: %s\n", target.String()))
	}
	return sb.String()
}

// String Displays the ContainerIP data structure
func (cip *ContainerIP) String() string {
	return fmt.Sprintf("%s ( %s )",
		cip.IP,
		cip.ContainerIdentifier.StringLong(),
	)
}

// PrintNetTestContextMap displays the NetTestContext full map
func PrintNetTestContextMap(netsUnderTest map[string]NetTestContext) string {
	var sb strings.Builder
	if len(netsUnderTest) == 0 {
		sb.WriteString("No networks to test.\n")
	}
	for netName, netUnderTest := range netsUnderTest {
		sb.WriteString(fmt.Sprintf("***Test for Network attachment: %s\n", netName))
		sb.WriteString(fmt.Sprintf("%s\n", netUnderTest.String()))
	}
	return sb.String()
}

// PodIPsToStringList converts a list of corev1.PodIP objects into a list of strings
func PodIPsToStringList(ips []corev1.PodIP) (ipList []string) {
	for _, ip := range ips {
		ipList = append(ipList, ip.IP)
	}
	return ipList
}

// GetIPVersion parses a ip address from a string and returns its version
func GetIPVersion(aIP string) (IPVersion, error) {
	ip := net.ParseIP(aIP)
	if ip == nil {
		return Undefined, fmt.Errorf("%s is Not an IPv4 or an IPv6", aIP)
	}
	if ip.To4() != nil {
		return IPv4, nil
	}
	return IPv6, nil
}

// FilterIPListByIPVersion filters a list of ip strings by the provided version
// e.g. a list of mixed ipv4 and ipv6 when filtered with ipv6 version will return a list with
// the ipv6 addresses
func FilterIPListByIPVersion(ipList []string, aIPVersion IPVersion) []string {
	var filteredIPList []string
	for _, aIP := range ipList {
		if ver, _ := GetIPVersion(aIP); aIPVersion == ver {
			filteredIPList = append(filteredIPList, aIP)
		}
	}
	return filteredIPList
}

func FindRogueContainersDeclaringPorts(containers []*provider.Container, portsToTest map[int32]bool) []string {
	var rogueContainers []string
	for _, cut := range containers {
		for _, port := range cut.Ports {
			if portsToTest[port.ContainerPort] {
				tnf.ClaimFilePrintf("%s has declared a port (%d) that has been reserved", cut, port.ContainerPort)
				rogueContainers = append(rogueContainers, cut.String())
			}
		}
	}
	return rogueContainers
}

func FindRoguePodsListeningToPorts(pods []*provider.Pod, portsToTest map[int32]bool) (roguePods []string, failedContainers int) {
	for _, put := range pods {
		cut := put.Containers[0]

		listeningPorts, err := netutil.GetListeningPorts(cut)
		if err != nil {
			tnf.ClaimFilePrintf("Failed to get the listening ports on %s, err: %s", cut, err)
			failedContainers++
			continue
		}
		for port := range listeningPorts {
			if portsToTest[int32(port.PortNumber)] {
				tnf.ClaimFilePrintf("%s has one container listening on port %d that has been reserved", put, port.PortNumber)
				roguePods = append(roguePods, put.String())
			}
		}
	}
	return roguePods, failedContainers
}
