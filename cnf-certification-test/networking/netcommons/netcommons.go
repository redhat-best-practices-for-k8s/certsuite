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

package netcommons

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/networking/netutil"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
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

func FindRogueContainersDeclaringPorts(containers []*provider.Container, portsToTest map[int32]bool, portsOrigin string) (compliantObjects, nonCompliantObjects []*testhelper.ReportObject) {
	for _, cut := range containers {
		for _, port := range cut.Ports {
			if portsToTest[port.ContainerPort] {
				log.Debug("%s has declared a port (%d) that has been reserved", cut, port.ContainerPort)
				nonCompliantObjects = append(nonCompliantObjects,
					testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name,
						fmt.Sprintf("Container declares %s reserved port in %v", portsOrigin, portsToTest), false).
						SetType(testhelper.DeclaredPortType).
						AddField(testhelper.PortNumber, strconv.Itoa(int(port.ContainerPort))).
						AddField(testhelper.PortProtocol, string(port.Protocol)))
			} else {
				compliantObjects = append(compliantObjects,
					testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name,
						fmt.Sprintf("Container does not declare %s reserved port in %v", portsOrigin, portsToTest), true).
						SetType(testhelper.DeclaredPortType).
						AddField(testhelper.PortNumber, strconv.Itoa(int(port.ContainerPort))).
						AddField(testhelper.PortProtocol, string(port.Protocol)))
			}
		}
	}
	return compliantObjects, nonCompliantObjects
}

var ReservedIstioPorts = map[int32]bool{
	// https://istio.io/latest/docs/ops/deployment/requirements/#ports-used-by-istio
	15090: true, // Envoy Prometheus telemetry
	15053: true, // DNS port, if capture is enabled
	15021: true, // Health checks
	15020: true, // Merged Prometheus telemetry from Istio agent, Envoy, and application
	15009: true, // HBONE port for secure networks
	15008: true, // HBONE mTLS tunnel port
	15006: true, // Envoy inbound
	15004: true, // Debug port
	15001: true, // Envoy outbound
	15000: true, // Envoy admin port (commands/diagnostics)
}

func FindRoguePodsListeningToPorts(pods []*provider.Pod, portsToTest map[int32]bool, portsOrigin string) (compliantObjects, nonCompliantObjects []*testhelper.ReportObject) {
	for _, put := range pods {
		compliantObjectsEntries, nonCompliantObjectsEntries := FindRogueContainersDeclaringPorts(put.Containers, portsToTest, portsOrigin)
		nonCompliantPortFound := len(nonCompliantObjectsEntries) > 0
		compliantObjects = append(compliantObjects, compliantObjectsEntries...)
		nonCompliantObjects = append(nonCompliantObjects, nonCompliantObjectsEntries...)
		cut := put.Containers[0]
		listeningPorts, err := netutil.GetListeningPorts(cut)
		if err != nil {
			log.Debug("Failed to get the listening ports on %s, err: %v", cut, err)
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewPodReportObject(cut.Namespace, put.Name,
					fmt.Sprintf("Failed to get the listening ports on pod, err: %v", err), false))
			continue
		}
		for port := range listeningPorts {
			if ok := portsToTest[int32(port.PortNumber)]; ok {
				// If pod contains an "istio-proxy" container, we need to make sure that the ports returned
				// overlap with the known istio ports
				if put.ContainsIstioProxy() && ReservedIstioPorts[int32(port.PortNumber)] {
					log.Debug("%s was found to be listening to port %d due to istio-proxy being present. Ignoring.", put, port.PortNumber)
					continue
				}
				log.Debug("%s has one container (%s) listening on port %d that has been reserved", put, cut.Name, port.PortNumber)
				nonCompliantObjects = append(nonCompliantObjects,
					testhelper.NewPodReportObject(cut.Namespace, put.Name,
						fmt.Sprintf("Pod Listens to %s reserved port in %v", portsOrigin, portsToTest), false).
						SetType(testhelper.ListeningPortType).
						AddField(testhelper.PortNumber, strconv.Itoa(port.PortNumber)).
						AddField(testhelper.PortProtocol, port.Protocol))
				nonCompliantPortFound = true
			} else {
				compliantObjects = append(compliantObjects,
					testhelper.NewPodReportObject(cut.Namespace, put.Name,
						fmt.Sprintf("Pod Listens to port not in %s reserved port %v", portsOrigin, portsToTest), true).
						SetType(testhelper.ListeningPortType).
						AddField(testhelper.PortNumber, strconv.Itoa(port.PortNumber)).
						AddField(testhelper.PortProtocol, port.Protocol))
			}
		}
		if nonCompliantPortFound {
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewPodReportObject(cut.Namespace, put.Name,
					fmt.Sprintf("Pod listens to or its containers declares some %s reserved port in %v", portsOrigin, portsToTest), false))
			continue
		}
		compliantObjects = append(compliantObjects,
			testhelper.NewPodReportObject(cut.Namespace, put.Name,
				fmt.Sprintf("Pod does not listen to or declare any %s reserved port in %v", portsOrigin, portsToTest), true))
	}
	return compliantObjects, nonCompliantObjects
}

func TestReservedPortsUsage(env *provider.TestEnvironment, reservedPorts map[int32]bool, portsOrigin string) (compliantObjects, nonCompliantObjects []*testhelper.ReportObject) {
	compliantObjectsEntries, nonCompliantObjectsEntries := FindRoguePodsListeningToPorts(env.Pods, reservedPorts, portsOrigin)
	compliantObjects = append(compliantObjects, compliantObjectsEntries...)
	nonCompliantObjects = append(nonCompliantObjects, nonCompliantObjectsEntries...)

	return compliantObjects, nonCompliantObjects
}
