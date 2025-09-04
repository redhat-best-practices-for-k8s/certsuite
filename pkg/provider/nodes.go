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

// Node Encapsulates a Kubernetes node with optional machine configuration
//
// This structure holds a reference to the underlying corev1.Node object,
// providing convenient access to node metadata and status information. It
// optionally includes a MachineConfig for nodes managed by OpenShift, enabling
// retrieval of configuration details such as kernel settings or custom
// annotations. The struct’s methods offer helpers for OS detection, role
// identification, workload presence, and JSON serialization.
type Node struct {
	Data *corev1.Node
	Mc   MachineConfig `json:"-"`
}

// Node.MarshalJSON Serializes the node's internal data to JSON
//
// The method calls the standard library’s Marshal function with a pointer to
// the node’s Data field. It produces a byte slice containing the JSON
// representation of that data and returns any error encountered during
// marshaling.
func (node Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(&node.Data)
}

// Node.IsWorkerNode Determines if a node is considered a worker by inspecting its labels
//
// This method iterates over all labels attached to the node and checks each one
// against a predefined list of worker-identifying label patterns. It uses a
// helper that performs a substring match, allowing flexible recognition of
// common worker label conventions. The function returns true if any matching
// label is found; otherwise it returns false.
func (node *Node) IsWorkerNode() bool {
	for nodeLabel := range node.Data.Labels {
		if stringhelper.StringInSlice(WorkerLabels, nodeLabel, true) {
			return true
		}
	}
	return false
}

// Node.IsControlPlaneNode Determines whether the node is a control‑plane instance
//
// The method inspects each label on the node’s data and checks if any match
// known master labels using a string containment helper. If a matching label is
// found, it returns true; otherwise it returns false.
func (node *Node) IsControlPlaneNode() bool {
	for nodeLabel := range node.Data.Labels {
		if stringhelper.StringInSlice(MasterLabels, nodeLabel, true) {
			return true
		}
	}
	return false
}

// Node.IsRHCOS Determines whether a node runs Red Hat CoreOS
//
// The method examines the operating system image field of the node's status
// information, removing any surrounding whitespace before searching for the
// predefined CoreOS identifier string. If that identifier is present, it
// returns true; otherwise, it returns false. This check is used by other
// functions to confirm OS compatibility before proceeding with further
// operations.
func (node *Node) IsRHCOS() bool {
	return strings.Contains(strings.TrimSpace(node.Data.Status.NodeInfo.OSImage), rhcosName)
}

// Node.IsCSCOS Determines whether the node runs CoreOS
//
// This method inspects the operating system image string from the node’s
// status information, trims surrounding whitespace, and checks if it contains
// the CoreOS identifier. It returns true when the identifier is present,
// indicating a CoreOS or CentOS Stream CoreOS environment; otherwise it returns
// false.
func (node *Node) IsCSCOS() bool {
	return strings.Contains(strings.TrimSpace(node.Data.Status.NodeInfo.OSImage), cscosName)
}

// Node.IsRHEL checks whether the node’s OS image is a Red Hat Enterprise Linux release
//
// The method trims any surrounding whitespace from the node’s OS image string
// and then looks for the RHEL identifier within it. If the identifier is
// present, it returns true; otherwise it returns false. This boolean result is
// used by other functions to decide whether RHEL‑specific logic should be
// applied.
func (node *Node) IsRHEL() bool {
	return strings.Contains(strings.TrimSpace(node.Data.Status.NodeInfo.OSImage), rhelName)
}

// Node.IsRTKernel Indicates if the node uses a real‑time kernel
//
// This method examines the node's kernel version string, trims whitespace, and
// checks for the presence of "rt" to determine whether a real‑time kernel is
// installed. It returns true when the substring is found, otherwise false.
func (node *Node) IsRTKernel() bool {
	// More information: https://www.redhat.com/sysadmin/real-time-kernel
	return strings.Contains(strings.TrimSpace(node.Data.Status.NodeInfo.KernelVersion), "rt")
}

// Node.GetRHCOSVersion Retrieves the short RHCOS version string from a node's OS image
//
// The function first verifies that the node is running Red Hat Enterprise Linux
// CoreOS, returning an error if not. It then parses the OSImage field to
// extract the long version identifier and converts it into the corresponding
// short version using a helper routine. The resulting short version string is
// returned alongside any potential errors.
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

// Node.GetCSCOSVersion Retrieves the CoreOS version string from a node's OS image
//
// The function first verifies that the node is running CoreOS by checking its
// status. If not, it returns an error indicating an unsupported OS type. When
// valid, it parses the OSImage field to extract and return the CoreOS release
// identifier as a string.
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

// Node.GetRHELVersion Retrieves the major and minor RHEL version from a node
//
// The method first verifies that the node reports an OS image containing
// "RHEL"; if not, it returns an error indicating the OS type is invalid. It
// then splits the OS image string on the RHEL identifier, trims any surrounding
// whitespace, and extracts the leading numeric part of the remaining string as
// the version. The extracted version string is returned along with a nil error
// when successful.
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

// Node.IsHyperThreadNode Determines if the node supports hyper‑threading
//
// The method runs a predefined command inside a probe pod on the node to query
// CPU core information. It parses the output for the number of threads per core
// and returns true when more than one thread is reported, indicating
// hyper‑threading support. Errors from execution or parsing are returned
// alongside the boolean result.
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

// Node.HasWorkloadDeployed Determines whether any of a set of pods are running on this node
//
// The method walks through each pod in the provided slice and inspects its spec
// to see if the node name matches the current node’s name. If it finds a
// match, it immediately returns true; otherwise, after checking all pods it
// returns false.
func (node *Node) HasWorkloadDeployed(podsUnderTest []*Pod) bool {
	for _, pod := range podsUnderTest {
		if pod.Spec.NodeName == node.Data.Name {
			return true
		}
	}
	return false
}
