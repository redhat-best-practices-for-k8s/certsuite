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

package netcommons

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/networking/netutil"
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

// IPVersion.String Returns the textual form of an IP version
//
// The method examines the value of the receiver and maps each predefined
// constant to its corresponding string. It covers IPv4, IPv6, combined
// IPv4/IPv6, and an undefined case. If none match, it defaults to the undefined
// string.
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

// NetTestContext Describes a network test setup for a subnet
//
// This structure holds information about which container initiates ping tests
// on a given network, the node it runs on, and the list of target containers to
// be pinged. The tester source is chosen randomly from available containers. It
// provides a string representation that lists the initiating container followed
// by all destination targets.
type NetTestContext struct {
	// testerContainerNodeOc session context to access the node running the container selected to initiate tests
	TesterContainerNodeName string
	// testerSource is the container select to initiate the ping tests on this given network
	TesterSource ContainerIP
	// ipDestTargets List of containers to be pinged by the testerSource on this given network
	DestTargets []ContainerIP
}

// ContainerIP Formats a container's IP address with its identifier
//
// This method returns a human‑readable representation that combines the
// container’s IP address and a long form of its identifier. It concatenates
// the two strings with parentheses to clearly separate the network address from
// the container details. The output is useful for logging or debugging
// networking tests.
type ContainerIP struct {
	// ip address of the target container
	IP string
	// targetContainerIdentifier container identifier including namespace, pod name, container name, node name, and container UID
	ContainerIdentifier *provider.Container
	// interfaceName is the interface we want to target for the ping test
	InterfaceName string
}

// NetTestContext.String Formats the network test context for display
//
// This method builds a multi-line string describing the container that
// initiates the tests and each target container it will communicate with. It
// first writes the source container, then lists all destination containers or
// indicates when none are present. The resulting string is returned for logging
// or debugging purposes.
func (testContext *NetTestContext) String() string {
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

// ContainerIP.String Formats the container IP address with its identifier
//
// This method constructs a string that shows the IP address followed by the
// long form of the container’s identifier in parentheses. It uses formatting
// utilities to combine the two pieces into a single readable representation,
// which is returned as a string.
func (cip *ContainerIP) String() string {
	return fmt.Sprintf("%s ( %s )",
		cip.IP,
		cip.ContainerIdentifier.StringLong(),
	)
}

// PrintNetTestContextMap Formats a map of network test contexts into a readable string
//
// This function iterates over a mapping from network names to NetTestContext
// objects, building a multi-line string that begins with a header for each
// network and then includes the detailed output of the context’s String
// method. If no networks are provided it returns a short message indicating
// there is nothing to test. The resulting string is used by other parts of the
// test suite to log or display current test conditions.
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

// PodIPsToStringList Transforms a slice of PodIP structures into plain IP address strings
//
// The function iterates over each corev1.PodIP element, extracts the IP string
// field, and appends it to a new slice. It returns this list of string
// addresses for use elsewhere in the package. The operation is linear in the
// number of input items and requires no additional dependencies beyond standard
// Go append.
func PodIPsToStringList(ips []corev1.PodIP) (ipList []string) {
	for _, ip := range ips {
		ipList = append(ipList, ip.IP)
	}
	return ipList
}

// GetIPVersion determines whether a string represents an IPv4 or IPv6 address
//
// The function attempts to parse the input as an IP address using the standard
// library. If parsing fails, it reports that the string is not a valid IP. It
// then distinguishes between IPv4 and IPv6 by checking if the parsed address
// can be converted to a four‑byte form; the result is returned along with any
// error.
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

// FilterIPListByIPVersion Selects addresses matching a specified IP version
//
// The function receives a slice of string IPs and an IP version to filter by.
// It iterates over the list, determines each address’s version, and keeps
// only those that match the requested type. The resulting slice contains only
// IPv4 or IPv6 addresses as requested.
func FilterIPListByIPVersion(ipList []string, aIPVersion IPVersion) []string {
	var filteredIPList []string
	for _, aIP := range ipList {
		if ver, _ := GetIPVersion(aIP); aIPVersion == ver {
			filteredIPList = append(filteredIPList, aIP)
		}
	}
	return filteredIPList
}

// findRogueContainersDeclaringPorts identifies containers that declare prohibited ports
//
// The function scans a list of containers, checking each declared port against
// a set of reserved ports. For every match it records a non‑compliant report;
// otherwise it logs compliance and creates a compliant report object. It
// returns slices of these report objects for further processing.
func findRogueContainersDeclaringPorts(containers []*provider.Container, portsToTest map[int32]bool, portsOrigin string, logger *log.Logger) (compliantObjects, nonCompliantObjects []*testhelper.ReportObject) {
	for _, cut := range containers {
		logger.Info("Testing Container %q", cut)
		for _, port := range cut.Ports {
			if portsToTest[port.ContainerPort] {
				logger.Error("%q declares %s reserved port %d (%s)", cut, portsOrigin, port.ContainerPort, port.Protocol)
				nonCompliantObjects = append(nonCompliantObjects,
					testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name,
						fmt.Sprintf("Container declares %s reserved port in %v", portsOrigin, portsToTest), false).
						SetType(testhelper.DeclaredPortType).
						AddField(testhelper.PortNumber, strconv.Itoa(int(port.ContainerPort))).
						AddField(testhelper.PortProtocol, string(port.Protocol)))
			} else {
				logger.Info("%q does not declare any %s reserved port", cut, portsOrigin)
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

// findRoguePodsListeningToPorts Detects pods that are listening on or declaring reserved ports
//
// The function iterates over each pod, checking its containers for declared
// ports and actual listening sockets against a set of prohibited port numbers.
// It logs detailed information and generates report objects indicating
// compliance status for both container declarations and pod-level listening
// behavior. Non‑compliant pods are reported with the specific port and
// protocol that violates the reservation rules.
func findRoguePodsListeningToPorts(pods []*provider.Pod, portsToTest map[int32]bool, portsOrigin string, logger *log.Logger) (compliantObjects, nonCompliantObjects []*testhelper.ReportObject) {
	for _, put := range pods {
		logger.Info("Testing Pod %q", put)
		compliantObjectsEntries, nonCompliantObjectsEntries := findRogueContainersDeclaringPorts(put.Containers, portsToTest, portsOrigin, logger)
		nonCompliantPortFound := len(nonCompliantObjectsEntries) > 0
		compliantObjects = append(compliantObjects, compliantObjectsEntries...)
		nonCompliantObjects = append(nonCompliantObjects, nonCompliantObjectsEntries...)
		cut := put.Containers[0]
		listeningPorts, err := netutil.GetListeningPorts(cut)
		if err != nil {
			logger.Error("Failed to get the listening ports on %q, err: %v", cut, err)
			nonCompliantObjects = append(nonCompliantObjects,
				testhelper.NewPodReportObject(cut.Namespace, put.Name,
					fmt.Sprintf("Failed to get the listening ports on pod, err: %v", err), false))
			continue
		}
		for port := range listeningPorts {
			if ok := portsToTest[port.PortNumber]; ok {
				// If pod contains an "istio-proxy" container, we need to make sure that the ports returned
				// overlap with the known istio ports
				if put.ContainsIstioProxy() && ReservedIstioPorts[port.PortNumber] {
					logger.Info("%q was found to be listening to port %d due to istio-proxy being present. Ignoring.", put, port.PortNumber)
					continue
				}

				logger.Error("%q has one container (%q) listening on port %d (%s) that has been reserved", put, cut.Name, port.PortNumber, port.Protocol)
				nonCompliantObjects = append(nonCompliantObjects,
					testhelper.NewPodReportObject(cut.Namespace, put.Name,
						fmt.Sprintf("Pod Listens to %s reserved port in %v", portsOrigin, portsToTest), false).
						SetType(testhelper.ListeningPortType).
						AddField(testhelper.PortNumber, strconv.Itoa(int(port.PortNumber))).
						AddField(testhelper.PortProtocol, port.Protocol))
				nonCompliantPortFound = true
			} else {
				logger.Info("%q listens in %s unreserved port %d (%s)", put, portsOrigin, port.PortNumber, port.Protocol)
				compliantObjects = append(compliantObjects,
					testhelper.NewPodReportObject(cut.Namespace, put.Name,
						fmt.Sprintf("Pod Listens to port not in %s reserved port %v", portsOrigin, portsToTest), true).
						SetType(testhelper.ListeningPortType).
						AddField(testhelper.PortNumber, strconv.Itoa(int(port.PortNumber))).
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

// TestReservedPortsUsage checks pods for listening on or declaring reserved ports
//
// The function receives a test environment, a map of port numbers that are
// considered reserved, an origin label for those ports, and a logger. It scans
// all pods in the environment to find any containers listening on or declaring
// these ports, excluding known Istio proxy exceptions. The result is two slices
// of report objects indicating compliant and non‑compliant findings.
func TestReservedPortsUsage(env *provider.TestEnvironment, reservedPorts map[int32]bool, portsOrigin string, logger *log.Logger) (compliantObjects, nonCompliantObjects []*testhelper.ReportObject) {
	compliantObjectsEntries, nonCompliantObjectsEntries := findRoguePodsListeningToPorts(env.Pods, reservedPorts, portsOrigin, logger)
	compliantObjects = append(compliantObjects, compliantObjectsEntries...)
	nonCompliantObjects = append(nonCompliantObjects, nonCompliantObjectsEntries...)

	return compliantObjects, nonCompliantObjects
}
