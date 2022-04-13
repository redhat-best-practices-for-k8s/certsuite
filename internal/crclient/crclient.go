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

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
)

func GetPidFromContainer(cut *provider.Container, ctx clientsholder.Context) (int, error) {
	var pidCmd string

	switch cut.Runtime {
	case "docker": //nolint:goconst // used only once
		pidCmd = "chroot /host docker inspect -f '{{.State.Pid}}' " + cut.UID + " 2>/dev/null"
	case "docker-pullable": //nolint:goconst // used only once
		pidCmd = "chroot /host docker inspect -f '{{.State.Pid}}' " + cut.UID + " 2>/dev/null"
	case "cri-o", "containerd": //nolint:goconst // used only once
		pidCmd = "chroot /host crictl inspect --output go-template --template '{{.info.pid}}' " + cut.UID + " 2>/dev/null"
	default:
		logrus.Debugf("Container runtime %s not supported yet for this test, skipping", cut.Runtime)
		return 0, fmt.Errorf("container runtime not supported")
	}

	ch := clientsholder.GetClientsHolder()
	outStr, errStr, err := ch.ExecCommandContainer(ctx, pidCmd)
	if err != nil {
		return 0, fmt.Errorf("can't execute command: \" %s \"  on %s err:%s", pidCmd, cut, err)
	}
	if errStr != "" {
		return 0, fmt.Errorf("cmd: \" %s \" on %s returned %s", pidCmd, cut, errStr)
	}

	return strconv.Atoi(strings.TrimSuffix(outStr, "\n"))
}

// ExecCommandContainerNSEnter executes a command in the specified container namespace using nsenter
func ExecCommandContainerNSEnter(command string,
	aContainer *provider.Container) (outStr, errStr string, err error) {
	// Getting env
	env := provider.GetTestEnvironment()
	// Getting the debug pod corresponding to the container's node
	debugPod := env.DebugPods[aContainer.NodeName]
	if debugPod == nil {
		err = fmt.Errorf("debug pod not found on Node: %s trying to run command: \" %s \" Namespace: %s Pod: %s container %s err:%s", aContainer.NodeName, command, aContainer.Namespace, aContainer.Podname, aContainer.Data.Name, err)
		return "", "", err
	}
	o := clientsholder.GetClientsHolder()
	ctx := clientsholder.Context{Namespace: debugPod.Namespace, Podname: debugPod.Name, Containername: debugPod.Spec.Containers[0].Name}

	// Get the container PID to build the nsenter command
	containerPid, err := GetPidFromContainer(aContainer, ctx)
	if err != nil {
		return "", "", fmt.Errorf("cannot get PID from: %s, err: %s", aContainer, err)
	}

	// Add the container PID and the specific command to run with nsenter
	nsenterCommand := "nsenter -t " + strconv.Itoa(containerPid) + " -n " + command

	// Run the nsenter command on the debug pod
	outStr, errStr, err = o.ExecCommandContainer(ctx, nsenterCommand)
	if err != nil {
		return "", "", fmt.Errorf("can't execute command: \" %s \"  on %s err:%s", command, aContainer, err)
	}
	return outStr, errStr, err
}
