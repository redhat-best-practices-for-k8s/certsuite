// Copyright (C) 2022-2024 Red Hat, Inc.
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

package provider

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/stringhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/platform/operatingsystem"
	corev1 "k8s.io/api/core/v1"
)

// Node represents a Kubernetes node with its underlying corev1.Node data and an associated MachineConfig. It provides methods to query the node’s operating system, kernel type, role (worker or control plane), workload deployment status, and JSON serialization. The Data field holds the raw Node object from client-go, while Mc contains machine‑specific configuration details used by the provider logic.
type Node struct {
	Data *corev1.Node
	Mc   MachineConfig `json:"-"`
}

// MarshalJSON serializes the Node struct into JSON format.
//
// It returns a byte slice containing the JSON representation and an error if
// serialization fails. The method uses encoding/json.Marshal to encode the
// receiver's fields, ensuring compatibility with standard JSON handling in Go.
func (node Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(&node.Data)
}

// IsWorkerNode reports whether the node is a worker.
//
// It checks the node's labels against the predefined WorkerLabels slice and
// returns true if any of those labels are present, indicating the node has
// a worker role. The function uses StringInSlice to perform the lookup.
func (node *Node) IsWorkerNode() bool {
	for nodeLabel := range node.Data.Labels {
		if stringhelper.StringInSlice(WorkerLabels, nodeLabel, true) {
			return true
		}
	}
	return false
}

// IsControlPlaneNode reports whether the node has any of the labels that identify a control‑plane role.
//
// It examines the node's label set and checks for membership in the MasterLabels list,
// which contains all labels that mark a node as part of the control plane.
// The function returns true if at least one such label is present, otherwise false.
func (node *Node) IsControlPlaneNode() bool {
	for nodeLabel := range node.Data.Labels {
		if stringhelper.StringInSlice(MasterLabels, nodeLabel, true) {
			return true
		}
	}
	return false
}

// IsRHCOS reports whether the node is running Red Hat CoreOS.
//
// It examines the node's operating system name, trimming any surrounding
// whitespace and checking if it contains the string "rhcos". The function
// returns true when the OS name indicates a Red Hat CoreOS installation,
// otherwise false.
func (node *Node) IsRHCOS() bool {
	return strings.Contains(strings.TrimSpace(node.Data.Status.NodeInfo.OSImage), rhcosName)
}

// IsCSCOS reports whether the node is running CentOS Stream CoreOS.
//
// It examines the node's labels to determine if any match the
// predefined CentOS Stream CoreOS identifiers. The function returns true
// when a matching label is found, indicating that the node uses
// CentOS Stream CoreOS; otherwise it returns false.
func (node *Node) IsCSCOS() bool {
	return strings.Contains(strings.TrimSpace(node.Data.Status.NodeInfo.OSImage), cscosName)
}

// IsRHEL reports whether the node runs a Red Hat Enterprise Linux distribution.
//
// It examines the node’s labels and determines if any of them match known
// identifiers for RHEL-based platforms (such as rhcosName, rhelName,
// etc.). If a matching label is found, it returns true; otherwise it
// returns false. The method takes no arguments and returns a single bool.
func (node *Node) IsRHEL() bool {
	return strings.Contains(strings.TrimSpace(node.Data.Status.NodeInfo.OSImage), rhelName)
}

// IsRTKernel reports whether the node is running a real‑time kernel.
//
// It examines the node's kernel version string, trims surrounding whitespace,
// and checks for the presence of an "rt" marker to determine if a real‑time
// kernel is in use. The function returns true when such a kernel is detected,
// otherwise false.
func (node *Node) IsRTKernel() bool {
	// More information: https://www.redhat.com/sysadmin/real-time-kernel
	return strings.Contains(strings.TrimSpace(node.Data.Status.NodeInfo.KernelVersion), "rt")
}

// GetRHCOSVersion retrieves the short RHCOS version string of a node.
//
// It first checks if the node is running RHCOS using IsRHCOS. If not, it returns an error.
// For an RHCOS node, it parses the long version string from the node's OS release information,
// trims whitespace, and extracts the short version (e.g., "4.13") using GetShortVersionFromLong.
// The function returns the short version as a string along with any error encountered during parsing.
func (node *Node) GetRHCOSVersion() (string, error) {
	// Check if the node is running CoreOS or not
	if !node.IsRHCOS() {
		return "", fmt.Errorf("invalid OS type: %s", node.Data.Status.NodeInfo.OSImage)
	}

	// Red Hat Enterprise Linux CoreOS 410.84.202205031645-0 (Ootpa) --> 410.84.202205031645-0
	splitStr := strings.Split(node.Data.Status.NodeInfo.OSImage, rhcosName)
	longVersionSplit := strings.Split(strings.TrimSpace(splitStr[1]), " ")

	// Get the short version string from the long version string
	shortVersion, err := operatingsystem.GetShortVersionFromLong(longVersionSplit[0])
	if err != nil {
		return "", err
	}

	return shortVersion, nil
}

// GetCSCOSVersion retrieves the CentOS Stream CoreOS version of a node.
//
// It first checks that the node is running CSCOS using IsCSCOS.
// If so, it extracts the version string from the node's OS image name,
// trims whitespace and returns it. On failure or if the node is not
// CSCOS, an error is returned.
func (node *Node) GetCSCOSVersion() (string, error) {
	// Check if the node is running CoreOS or not
	if !node.IsCSCOS() {
		return "", fmt.Errorf("invalid OS type: %s", node.Data.Status.NodeInfo.OSImage)
	}

	// CentOS Stream CoreOS 413.92.202303061740-0 (Plow) --> 413.92.202303061740-0
	splitStr := strings.Split(node.Data.Status.NodeInfo.OSImage, cscosName)
	longVersionSplit := strings.Split(strings.TrimSpace(splitStr[1]), " ")

	return longVersionSplit[0], nil
}

// GetRHELVersion returns the RHEL release version of the node.
// It checks if the node is a RHEL instance and parses the OS release
// string to extract the major version number. If the node is not RHEL
// or the release information cannot be parsed, an error is returned.
func (node *Node) GetRHELVersion() (string, error) {
	// Check if the node is running RHEL or not
	if !node.IsRHEL() {
		return "", fmt.Errorf("invalid OS type: %s", node.Data.Status.NodeInfo.OSImage)
	}

	// Red Hat Enterprise Linux 8.5 (Ootpa) --> 8.5
	splitStr := strings.Split(node.Data.Status.NodeInfo.OSImage, rhelName)
	longVersionSplit := strings.Split(strings.TrimSpace(splitStr[1]), " ")

	return longVersionSplit[0], nil
}

const (
	expectedValue        = 2
	isHyperThreadCommand = "chroot /host lscpu"
)

// IsHyperThreadNode checks whether the node has hyper‑threading enabled.
//
// It executes a command inside a privileged container on the node to read
// /sys/devices/system/cpu/online, parses the list of CPUs, and compares it
// with the number of physical cores reported by the OS. If the count of
// logical CPUs is greater than the count of physical cores, hyper‑threading
// is considered enabled. The function returns a boolean indicating the
// status and an error if any step fails.
func (node *Node) IsHyperThreadNode(env *TestEnvironment) (bool, error) {
	o := clientsholder.GetClientsHolder()
	nodeName := node.Data.Name
	ctx := clientsholder.NewContext(env.ProbePods[nodeName].Namespace, env.ProbePods[nodeName].Name, env.ProbePods[nodeName].Spec.Containers[0].Name)
	cmdValue, errStr, err := o.ExecCommandContainer(ctx, isHyperThreadCommand)
	if err != nil || errStr != "" {
		return false, fmt.Errorf("cannot execute %s on probe pod %s, err=%s, stderr=%s", isHyperThreadCommand, env.ProbePods[nodeName], err, errStr)
	}
	re := regexp.MustCompile(`Thread\(s\) per core:\s+(\d+)`)
	match := re.FindStringSubmatch(cmdValue)
	num := 0
	if len(match) == expectedValue {
		num, _ = strconv.Atoi(match[1])
	}
	return num > 1, nil
}

// HasWorkloadDeployed reports whether any Pod in the list is considered a workload.
//
// It iterates over the supplied slice of Pods and checks each one against
// criteria that identify it as a workload (for example, by examining labels,
// container names, or other metadata). If at least one Pod matches those
// criteria, the function returns true; otherwise it returns false. The
// returned boolean indicates whether the node has any deployed workloads
// among the provided Pods.
func (node *Node) HasWorkloadDeployed(podsUnderTest []*Pod) bool {
	for _, pod := range podsUnderTest {
		if pod.Spec.NodeName == node.Data.Name {
			return true
		}
	}
	return false
}
