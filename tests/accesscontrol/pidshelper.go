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

package accesscontrol

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
)

const nbProcessesIndex = 2

// getNbOfProcessesInPidNamespace retrieves the number of processes in the Pid namespace.
// Returns:
//   - int :  the number of processes in the PID namespace associated with the specified process ID
//   - error : An error, if any occurred during the execution of the command or parsing of the output.
func getNbOfProcessesInPidNamespace(ctx clientsholder.Context, targetPid int, ch clientsholder.Command) (int, error) {
	cmd := "lsns -p " + strconv.Itoa(targetPid) + " -t pid -n"

	outStr, errStr, err := ch.ExecCommandContainer(ctx, cmd)
	if err != nil {
		return 0, fmt.Errorf("can not execute command: \" %s \", err:%s", cmd, err)
	}
	if errStr != "" {
		return 0, fmt.Errorf("cmd: \" %s \" returned %s", cmd, errStr)
	}

	retValues := strings.Fields(outStr)
	if len(retValues) <= nbProcessesIndex {
		return 0, fmt.Errorf("cmd: \" %s \" returned an invalid value %s", cmd, outStr)
	}
	return strconv.Atoi(retValues[nbProcessesIndex])
}
