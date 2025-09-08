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

// PortInfo Describes a network port with number and protocol
//
// This structure holds the numeric value of a listening port and the transport
// protocol used, such as TCP or UDP. It is used to identify unique ports in
// mappings returned by functions that parse command output for listening
// sockets.
type PortInfo struct {
	PortNumber int32
	Protocol   string
}

// parseListeningPorts parses command output into a map of listening ports
//
// The function takes the raw string from a network command and splits it line
// by line, extracting protocol and port number when the state indicates LISTEN.
// It converts the numeric part to an integer, normalizes the protocol name to
// upper case, and stores each unique pair in a map keyed by PortInfo with a
// boolean value of true. Errors during conversion cause an immediate return
// with an error message.
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

// GetListeningPorts Retrieves the set of ports currently listening inside a container
//
// The function runs an nsenter command inside the target container to list open
// sockets, then parses the output into a map keyed by port information. It
// returns this map along with any error that occurs during execution or
// parsing.
func GetListeningPorts(cut *provider.Container) (map[PortInfo]bool, error) {
	outStr, errStr, err := crclient.ExecCommandContainerNSEnter(getListeningPortsCmd, cut)
	if err != nil || errStr != "" {
		return nil, fmt.Errorf("failed to execute command %s on %s, err: %v", getListeningPortsCmd, cut, err)
	}

	return parseListeningPorts(outStr)
}

// GetSSHDaemonPort Retrieves the SSH daemon listening port within a container
//
// This function runs a shell command inside the specified container to locate
// the sshd process and extract its bound TCP port. It executes the command via
// nsenter, handles any execution errors or non‑empty stderr output, and
// returns the trimmed port number as a string. If the command fails or returns
// no output, an error is returned.
func GetSSHDaemonPort(cut *provider.Container) (string, error) {
	const findSSHDaemonPort = "ss -tpln | grep sshd | head -1 | awk '{ print $4 }' | awk -F : '{ print $2 }'"
	outStr, errStr, err := crclient.ExecCommandContainerNSEnter(findSSHDaemonPort, cut)
	if err != nil || errStr != "" {
		return "", fmt.Errorf("failed to execute command %s on %s, err: %v", findSSHDaemonPort, cut, err)
	}

	return strings.TrimSpace(outStr), nil
}
