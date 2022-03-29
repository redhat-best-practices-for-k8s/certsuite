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

package accesscontrol

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
)

func getPidFromContainer(cut *provider.Container, ctx clientsholder.Context) (int, error) {
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
		return 0, errors.Errorf("Container runtime not supported")
	}

	ch := clientsholder.GetClientsHolder()
	outStr, errStr, err := ch.ExecCommandContainer(ctx, pidCmd)
	if err != nil {
		return 0, errors.Errorf("can't execute command: \" %s \"  on %s err:%s", pidCmd, cut.StringShort(), err)
	}
	if errStr != "" {
		return 0, errors.Errorf("cmd: \" %s \" on %s returned %s", pidCmd, cut.StringShort(), errStr)
	}

	pid, err := strconv.Atoi(strings.TrimSuffix(outStr, "\n"))
	if err != nil {
		return 0, err
	}

	return pid, nil
}

func getNbOfProcessesInPidNamespace(ctx clientsholder.Context, targetPid int) (int, error) {
	cmd := "lsns -p " + strconv.Itoa(targetPid) + " -t pid -n"

	ch := clientsholder.GetClientsHolder()
	outStr, errStr, err := ch.ExecCommandContainer(ctx, cmd)
	if err != nil {
		return 0, errors.Errorf("can't execute command: \" %s \", err:%s", cmd, err)
	}
	if errStr != "" {
		return 0, errors.Errorf("cmd: \" %s \" returned %s", cmd, errStr)
	}

	const nbProcessesIndex = 2
	nbProcesses, err := strconv.Atoi(strings.Fields(outStr)[nbProcessesIndex])
	if err != nil {
		return 0, err
	}

	return nbProcesses, nil
}
