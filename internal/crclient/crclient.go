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

package crclient

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
)

const PsRegex = `(?m)^(\d+?)\s+?(\d+?)\s+?(.*?)$`

type Process struct {
	PidNs, Pid int
	Args       string
}

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

func GetContainerProcesses(container *provider.Container, env *provider.TestEnvironment) ([]*Process, error) {
	pidNs, err := GetContainerPidNamespace(container, env)
	if err != nil {
		return nil, fmt.Errorf("could not get the containers' pid namespace, err: %v", err)
	}

	return GetPidsFromPidNamespace(pidNs, container)
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

func GetPidsFromPidNamespace(pidNamespace string, container *provider.Container) (p []*Process, err error) {
	const command = "ps -e -o pidns,pid,args"

	stdout, stderr, err := ExecCommandContainerNSEnter(command, container)
	if err != nil || stderr != "" {
		err = fmt.Errorf("unable to run nsenter due to : %v", err)
		return p, err
	}
	re := regexp.MustCompile(PsRegex)
	matches := re.FindAllStringSubmatch(stdout, -1)
	// If we do not find a successful log, we fail
	for _, v := range matches {
		// Matching only the right PidNs
		if pidNamespace != v[1] {
			continue
		}
		aPidNs, err := strconv.Atoi(v[1])
		if err != nil {
			logrus.Errorf("could not convert string %s to integer, err=%s", v[1], err)
			continue
		}
		aPid, err := strconv.Atoi(v[2])
		if err != nil {
			logrus.Errorf("could not convert string %s to integer, err=%s", v[2], err)
			continue
		}
		p = append(p, &Process{PidNs: aPidNs, Pid: aPid, Args: v[3]})
	}
	return p, nil
}
