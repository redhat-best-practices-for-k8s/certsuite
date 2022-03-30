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
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
)

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
	nbProcesses := strings.Fields(outStr)[nbProcessesIndex]

	return strconv.Atoi(nbProcesses)
}
