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

package crclient

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
)

func GetPidFromContainer(cut *provider.Container, ctx clientsholder.Context) (int, error) {
	var pidCmd string

	switch cut.Runtime {
	case "docker":
		pidCmd = "chroot /host docker inspect -f '{{.State.Pid}}' " + cut.UID + " 2>/dev/null"
	case "docker-pullable":
		pidCmd = "chroot /host docker inspect -f '{{.State.Pid}}' " + cut.UID + " 2>/dev/null"
	case "cri-o", "containerd":
		pidCmd = "chroot /host crictl inspect --output go-template --template '{{.info.pid}}' " + cut.UID + " 2>/dev/null"
	default:
		logrus.Debugf("Container runtime %s not supported yet for this test, skipping", cut.Runtime)
		return 0, fmt.Errorf("container runtime %s not supported", cut.Runtime)
	}

	ch := clientsholder.GetClientsHolder()
	outStr, errStr, err := ch.ExecCommandContainer(ctx, pidCmd)
	if err != nil {
		return 0, fmt.Errorf("cannot execute command: \" %s \"  on %s err:%s", pidCmd, cut, err)
	}
	if errStr != "" {
		return 0, fmt.Errorf("cmd: \" %s \" on %s returned %s", pidCmd, cut, errStr)
	}

	return strconv.Atoi(strings.TrimSuffix(outStr, "\n"))
}

// To get the pid namespace of the container
func GetContainerPidNamespace(testContainer *provider.Container, env *provider.TestEnvironment) (string, error) {
	// Get the container pid
	nodeName := testContainer.NodeName
	debugPod := env.DebugPods[nodeName]
	if debugPod == nil {
		ginkgo.Fail(fmt.Sprintf("Debug pod not found on Node: %s", nodeName))
	}
	ocpContext := clientsholder.NewContext(debugPod.Namespace, debugPod.Name, debugPod.Spec.Containers[0].Name)
	pid, err := GetPidFromContainer(testContainer, ocpContext)
	if err != nil {
		return "", fmt.Errorf("unable to get container process id due to: %v", err)
	}
	logrus.Debugf("Obtained process id for %s is %d", testContainer, pid)

	command := fmt.Sprintf("lsns -p %d -t pid -n", pid)
	stdout, stderr, err := ExecCommandContainerNSEnter(command, testContainer)
	if err != nil || stderr != "" {
		return "", fmt.Errorf("unable to run nsenter due to : %v", err)
	}
	return strings.Fields(stdout)[0], nil
}

// To get pid of all processes running in the pid namespace from the target container
func GetPidsFromPidNamespace(pidNamespace string, container *provider.Container) []int {
	const newLine = "\n"
	const command = "ps -e -o pidns,pid,args"

	stdout, stderr, err := ExecCommandContainerNSEnter(command, container)
	if err != nil || stderr != "" {
		logrus.Errorf("unable to run nsenter due to : %v", err)
	}

	var pids []int
	for _, line := range strings.Split(strings.TrimSuffix(stdout, newLine), newLine) {
		if line == "" {
			continue
		}
		tokens := strings.Fields(line)
		if tokens[0] == pidNamespace {
			pid, err := strconv.Atoi(tokens[1])
			if err != nil {
				logrus.Errorf("err converting %s by strconv %v", tokens[1], err)
				continue
			}
			pids = append(pids, pid)
		}
	}

	return pids
}

func GetContainerPids(container *provider.Container, env *provider.TestEnvironment) ([]int, error) {
	pidNs, err := GetContainerPidNamespace(container, env)
	if err != nil {
		return nil, fmt.Errorf("could not get the containers' pid namespace, err: %v", err)
	}

	return GetPidsFromPidNamespace(pidNs, container), nil
}

// ExecCommandContainerNSEnter executes a command in the specified container namespace using nsenter
func ExecCommandContainerNSEnter(command string,
	aContainer *provider.Container) (outStr, errStr string, err error) {
	// Getting env
	env := provider.GetTestEnvironment()
	// Getting the debug pod corresponding to the container's node
	debugPod := env.DebugPods[aContainer.NodeName]
	if debugPod == nil {
		err = fmt.Errorf("debug pod not found on Node: %s trying to run command: \" %s \" Namespace: %s Pod: %s container %s err:%s", aContainer.NodeName, command, aContainer.Namespace, aContainer.Podname, aContainer.Name, err)
		return "", "", err
	}
	o := clientsholder.GetClientsHolder()
	ctx := clientsholder.NewContext(debugPod.Namespace, debugPod.Name, debugPod.Spec.Containers[0].Name)

	// Get the container PID to build the nsenter command
	containerPid, err := GetPidFromContainer(aContainer, ctx)
	if err != nil {
		return "", "", fmt.Errorf("cannot get PID from: %s, err: %v", aContainer, err)
	}

	// Add the container PID and the specific command to run with nsenter
	nsenterCommand := "nsenter -t " + strconv.Itoa(containerPid) + " -n " + command

	// Run the nsenter command on the debug pod
	outStr, errStr, err = o.ExecCommandContainer(ctx, nsenterCommand)
	if err != nil {
		return "", "", fmt.Errorf("cannot execute command: \" %s \"  on %s err:%s", command, aContainer, err)
	}
	return outStr, errStr, err
}
