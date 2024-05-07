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
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform/operatingsystem"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/stringhelper"
	corev1 "k8s.io/api/core/v1"
)

type Node struct {
	Data *corev1.Node
	Mc   MachineConfig `json:"-"`
}

func (node Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(&node.Data)
}

func (node *Node) IsWorkerNode() bool {
	for nodeLabel := range node.Data.Labels {
		if stringhelper.StringInSlice(WorkerLabels, nodeLabel, true) {
			return true
		}
	}
	return false
}

func (node *Node) IsMasterNode() bool {
	for nodeLabel := range node.Data.Labels {
		if stringhelper.StringInSlice(MasterLabels, nodeLabel, true) {
			return true
		}
	}
	return false
}

func (node *Node) IsRHCOS() bool {
	return strings.Contains(strings.TrimSpace(node.Data.Status.NodeInfo.OSImage), rhcosName)
}

func (node *Node) IsCSCOS() bool {
	return strings.Contains(strings.TrimSpace(node.Data.Status.NodeInfo.OSImage), cscosName)
}

func (node *Node) IsRHEL() bool {
	return strings.Contains(strings.TrimSpace(node.Data.Status.NodeInfo.OSImage), rhelName)
}

func (node *Node) IsRTKernel() bool {
	// More information: https://www.redhat.com/sysadmin/real-time-kernel
	return strings.Contains(strings.TrimSpace(node.Data.Status.NodeInfo.KernelVersion), "rt")
}

func (node *Node) GetRHCOSVersion() (string, error) {
	// Check if the node is running CoreOS or not
	if !node.IsRHCOS() {
		return "", fmt.Errorf("invalid OS type: %s", node.Data.Status.NodeInfo.OSImage)
	}

	path, err := os.Getwd()
	if err != nil {
		return "", err
	}

	//TODO: remove this workaround once the rhcos_version_map file can be provided instead of searched for.
	var filePath string
	tokens := strings.Split(path, "/")
	if tokens[len(tokens)-1] == "cnf-certification-test" {
		filePath = fmt.Sprintf(rhcosRelativePath, path)
	} else {
		filePath = fmt.Sprintf(rhcosRelativePath, path+"/cnf-certification-test")
	}

	// Red Hat Enterprise Linux CoreOS 410.84.202205031645-0 (Ootpa) --> 410.84.202205031645-0
	splitStr := strings.Split(node.Data.Status.NodeInfo.OSImage, rhcosName)
	longVersionSplit := strings.Split(strings.TrimSpace(splitStr[1]), " ")

	// Get the short version string from the long version string
	shortVersion, err := operatingsystem.GetShortVersionFromLong(longVersionSplit[0], filePath)
	if err != nil {
		return "", err
	}

	return shortVersion, nil
}

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

func (node *Node) IsHyperThreadNode(env *TestEnvironment) (bool, error) {
	o := clientsholder.GetClientsHolder()
	nodeName := node.Data.Name
	ctx := clientsholder.NewContext(env.DebugPods[nodeName].Namespace, env.DebugPods[nodeName].Name, env.DebugPods[nodeName].Spec.Containers[0].Name)
	cmdValue, errStr, err := o.ExecCommandContainer(ctx, isHyperThreadCommand)
	if err != nil || errStr != "" {
		return false, fmt.Errorf("cannot execute %s on debug pod %s, err=%s, stderr=%s", isHyperThreadCommand, env.DebugPods[nodeName], err, errStr)
	}
	re := regexp.MustCompile(`Thread\(s\) per core:\s+(\d+)`)
	match := re.FindStringSubmatch(cmdValue)
	num := 0
	if len(match) == expectedValue {
		num, _ = strconv.Atoi(match[1])
	}
	return num > 1, nil
}

func (node *Node) HasWorkloadDeployed(podsUnderTest []*Pod) bool {
	for _, pod := range podsUnderTest {
		if pod.Spec.NodeName == node.Data.Name {
			return true
		}
	}
	return false
}
