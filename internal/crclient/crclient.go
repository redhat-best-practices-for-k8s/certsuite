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
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
)

// ExecCommandContainerNSEnter executes a command in the specified container namespace using nsenter
func ExecCommandContainerNSEnter(command string,
	aContainer *provider.Container,
	env *provider.TestEnvironment) (outStr, errStr string, err error) {
	// Getting the debug pod corresponding to the container's node
	debugPod := env.DebugPods[aContainer.NodeName]
	if debugPod == nil {
		err = errors.Errorf("debug pod not found on Node: %s trying to run command: \" %s \" Namespace: %s Pod: %s container %s err:%s", aContainer.NodeName, command, aContainer.Namespace, aContainer.Podname, aContainer.Data.Name, err)
		return "", "", err
	}
	o := clientsholder.GetClientsHolder()
	ctx := clientsholder.Context{Namespace: debugPod.Namespace, Podname: debugPod.Name, Containername: debugPod.Spec.Containers[0].Name}

	// Starting to build nsenter command based on the container runtime: getting the container PID
	var nsenterCommand string
	switch aContainer.Runtime {
	case "docker": //nolint:goconst // used only once
		nsenterCommand = "PID=`chroot /host docker inspect -f '{{.State.Pid}}' " + aContainer.UID + " 2>/dev/null`"
	case "docker-pullable": //nolint:goconst // used only once
		nsenterCommand = "PID=`chroot /host docker inspect -f '{{.State.Pid}}' " + aContainer.UID + " 2>/dev/null`"
	case "cri-o", "containerd": //nolint:goconst // used only once
		nsenterCommand = "PID=`chroot /host crictl inspect --output go-template --template '{{.info.pid}}' " + aContainer.UID + " 2>/dev/null`"
	default:
		logrus.Debugf("Container runtime %s not supported yet for this test, skipping", aContainer.Runtime)
		return "", "", errors.Errorf("Container runtime not supported")
	}

	// Adding the nsenter command with the container PID
	nsenterCommand = nsenterCommand + ";nsenter -t $PID -n " + command

	// Run the nsenter command on on the debug pod
	outStr, errStr, err = o.ExecCommandContainer(ctx, nsenterCommand)
	if err != nil {
		return "", "", errors.Errorf("can't execute command: \" %s \"  on %s err:%s", command, aContainer, err)
	}
	return outStr, errStr, err
}
