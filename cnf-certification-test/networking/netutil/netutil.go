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

package netutil

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/test-network-function/cnf-certification-test/internal/crclient"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
)

const (
	getListeningPortsCmd = `ss -tulwnH`
	portStateListen      = "LISTEN"
	indexProtocol        = 0
	indexState           = 1
	indexPort            = 4
)

type PortInfo struct {
	PortNumber int
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

		port, err := strconv.Atoi(s[len(s)-1])
		if err != nil {
			return nil, fmt.Errorf("string to int conversion error, err: %s", err)
		}
		protocol := strings.ToUpper(fields[indexProtocol])
		portInfo := PortInfo{port, protocol}

		portSet[portInfo] = true
	}

	return portSet, nil
}

func GetListeningPorts(cut *provider.Container) (map[PortInfo]bool, error) {
	outStr, errStr, err := crclient.ExecCommandContainerNSEnter(getListeningPortsCmd, cut)
	if err != nil || errStr != "" {
		return nil, fmt.Errorf("failed to execute command %s on %s, err: %s", getListeningPortsCmd, cut, err)
	}

	return parseListeningPorts(outStr)
}
