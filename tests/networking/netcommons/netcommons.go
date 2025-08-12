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

// String returns a human‑readable name for the IPVersion.
//
// It converts the IPVersion value to its corresponding string such as
// "IPv4", "IPv6", or "Undefined". The returned string can be used for
// logging or display purposes when working with network interface types.
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

// NetTestContext represents a network test context for a subnet.
//
// NetTestContext holds information about which container initiates the test
// and the list of destination containers to ping on that subnet.
// The TesterSource field is the IP address of the randomly chosen tester
// container, while DestTargets lists the IPs of all other containers in
// the same network attachment. This structure is used by networking tests
// to configure and execute connectivity checks between pods.
type NetTestContext struct {
	// testerContainerNodeOc session context to access the node running the container selected to initiate tests
	TesterContainerNodeName string
	// testerSource is the container select to initiate the ping tests on this given network
	TesterSource ContainerIP
	// ipDestTargets List of containers to be pinged by the testerSource on this given network
	DestTargets []ContainerIP
}

// ContainerIP holds a container identification and its IP for networking tests.
//
// It stores the container identifier, the IP address assigned to that container,
// and the name of the network interface used in tests. This struct is used
// throughout the networking test suite to reference specific containers and
// their networking details when verifying connectivity and configuration.
type ContainerIP struct {
	// ip address of the target container
	IP string
	// targetContainerIdentifier container identifier including namespace, pod name, container name, node name, and container UID
	ContainerIdentifier *provider.Container
	// interfaceName is the interface we want to target for the ping test
	InterfaceName string
}

// String returns a formatted string representation of the NetTestContext.
//
// It serializes the internal state of the context, including interface
// configurations and any reserved ports, into a human‑readable format.
// The returned string can be used for logging or debugging purposes.
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

// String returns a formatted string representation of the ContainerIP.
//
// It builds a concise description of the container's IP address and
// related metadata using the StringLong helper to format the output.
// The returned string can be used for logging or debugging purposes.
func (cip *ContainerIP) String() string {
	return fmt.Sprintf("%s ( %s )",
		cip.IP,
		cip.ContainerIdentifier.StringLong(),
	)
}

// PrintNetTestContextMap displays the NetTestContext full map
//
// It takes a map of string keys to NetTestContext values and returns a formatted
// string representation of that map. The function iterates over each entry,
// converting both key and value to strings, and appends them to an output buffer.
// The resulting string includes the total number of entries in the map and
// provides a readable layout for debugging or logging purposes.
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

// PodIPsToStringList converts a slice of corev1.PodIP objects into a slice of strings.
//
// It iterates over each PodIP in the input slice, extracts the IP address as a string,
// and appends it to a new string slice which is returned. The resulting list contains
// only the IP addresses represented as plain text.
func PodIPsToStringList(ips []corev1.PodIP) (ipList []string) {
	for _, ip := range ips {
		ipList = append(ipList, ip.IP)
	}
	return ipList
}

// GetIPVersion parses a string and returns the IP version of the address.
//
// It attempts to parse the input as an IP address. If parsing succeeds, it
// determines whether the address is IPv4 or IPv6 by checking if the parsed
// address can be represented in 4‑byte form. The function returns an IPVersion
// value (IPv4, IPv6, Undefined) and an error if the string does not contain a
// valid IP address.
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

// FilterIPListByIPVersion returns only the addresses that match a specific IP version.
//
// It takes a slice of string representations of IP addresses and an IPVersion value.
// The function iterates over each address, determines its version using GetIPVersion,
// and appends it to a new slice if it matches the requested version. The resulting
// slice contains all matching addresses in their original order. If no addresses match,
// an empty slice is returned.
func FilterIPListByIPVersion(ipList []string, aIPVersion IPVersion) []string {
	var filteredIPList []string
	for _, aIP := range ipList {
		if ver, _ := GetIPVersion(aIP); aIPVersion == ver {
			filteredIPList = append(filteredIPList, aIP)
		}
	}
	return filteredIPList
}

// findRogueContainersDeclaringPorts identifies containers that expose ports which are not reserved by the system.
//
// It scans a list of containers and checks each container's declared port
// numbers against a map of reserved ports.  For every port that is not in
// the reservation map, it creates a report object describing the rogue
// declaration. The function logs progress and errors using the supplied
// logger. It returns a slice of report objects representing all
// containers with unreserved port declarations.
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

// findRoguePodsListeningToPorts identifies pods that are listening on ports not declared in their containers' specifications.
//
// It takes a slice of pod pointers, a map of port numbers considered safe, a namespace string for logging context,
// and a logger to record diagnostic messages. The function returns a slice of report objects describing any
// detected rogue pods. For each pod it checks the listening ports reported by the system; if a port is not in the
// allowed set or not declared by any container within the pod, a report entry is created. Pods that contain an
// Istio proxy are handled specially to avoid false positives. The returned reports include details such as
// pod name, namespace, offending ports, and a type indicating a rogue listening port situation.
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

// TestReservedPortsUsage checks whether any pods are listening on reserved Istio ports.
//
// It examines the provided mapping of port numbers to booleans, identifies
// rogue pods that are listening on those ports, logs relevant information,
// and returns a slice of report objects summarizing any violations found.
func TestReservedPortsUsage(env *provider.TestEnvironment, reservedPorts map[int32]bool, portsOrigin string, logger *log.Logger) (compliantObjects, nonCompliantObjects []*testhelper.ReportObject) {
	compliantObjectsEntries, nonCompliantObjectsEntries := findRoguePodsListeningToPorts(env.Pods, reservedPorts, portsOrigin, logger)
	compliantObjects = append(compliantObjects, compliantObjectsEntries...)
	nonCompliantObjects = append(nonCompliantObjects, nonCompliantObjectsEntries...)

	return compliantObjects, nonCompliantObjects
}
