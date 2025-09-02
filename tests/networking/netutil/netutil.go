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

package netutil

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/crclient"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
)

const (
	getListeningPortsCmd = `ss -tulwnH`
	portStateListen      = "LISTEN"
	indexProtocol        = 0
	indexState           = 1
	indexPort            = 4
)

type PortInfo struct {
	PortNumber int32
	Protocol   string
}

func parseListeningPorts(cmdOut string) (map[PortInfo]bool, error) {
	portSet := make(map[PortInfo]bool)

	cmdOut = strings.TrimSuffix(cmdOut, "\n")
	lines := strings.Split(cmdOut, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < indexPort+1 {
			continue
		}
		if fields[indexState] != portStateListen {
			continue
		}
		s := strings.Split(fields[indexPort], ":")
		if len(s) == 0 {
			continue
		}

		port, err := strconv.ParseInt(s[len(s)-1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("string to int conversion error, err: %v", err)
		}
		protocol := strings.ToUpper(fields[indexProtocol])
		portInfo := PortInfo{int32(port), protocol}

		portSet[portInfo] = true
	}

	return portSet, nil
}

func GetListeningPorts(cut *provider.Container) (map[PortInfo]bool, error) {
	outStr, errStr, err := crclient.ExecCommandContainerNSEnter(getListeningPortsCmd, cut)
	if err != nil || errStr != "" {
		return nil, fmt.Errorf("failed to execute command %s on %s, err: %v", getListeningPortsCmd, cut, err)
	}

	return parseListeningPorts(outStr)
}

func GetSSHDaemonPort(cut *provider.Container) (string, error) {
	const findSSHDaemonPort = "ss -tpln | grep sshd | head -1 | awk '{ print $4 }' | awk -F : '{ print $2 }'"
	outStr, errStr, err := crclient.ExecCommandContainerNSEnter(findSSHDaemonPort, cut)
	if err != nil || errStr != "" {
		return "", fmt.Errorf("failed to execute command %s on %s, err: %v", findSSHDaemonPort, cut, err)
	}

	return strings.TrimSpace(outStr), nil
}
