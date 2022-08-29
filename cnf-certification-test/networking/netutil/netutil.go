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

const (
	dumpNFTablesCmd       = "nft list ruleset"
	dumpIPTablesCmd       = "iptables-save"
	dumpIP6TablesCmd      = "ip6tables-save"
	ipTablesLegacyWarning = "# Warning: iptables-legacy tables present, use iptables-legacy-save to see them"
	// https://bugzilla.redhat.com/show_bug.cgi?id=1915027
	openshiftMachineConfigNft = `table ip filter {
		chain INPUT {
			type filter hook input priority filter; policy accept;
		}
	
		chain FORWARD {
			type filter hook forward priority filter; policy accept;
			meta l4proto tcp tcp dport 22623 tcp flags & (fin|syn|rst|ack) == syn counter packets 0 bytes 0 reject
			meta l4proto tcp tcp dport 22624 tcp flags & (fin|syn|rst|ack) == syn counter packets 0 bytes 0 reject
			meta l4proto tcp ip daddr 169.254.169.254 tcp dport != 53 counter packets 0 bytes 0 reject
			meta l4proto udp ip daddr 169.254.169.254 udp dport 53 counter packets 0 bytes 0 reject
		}
	
		chain OUTPUT {
			type filter hook output priority filter; policy accept;
			meta l4proto tcp tcp dport 22623 tcp flags & (fin|syn|rst|ack) == syn counter packets 0 bytes 0 reject
			meta l4proto tcp tcp dport 22624 tcp flags & (fin|syn|rst|ack) == syn counter packets 0 bytes 0 reject
			meta l4proto tcp ip daddr 169.254.169.254 tcp dport != 53 counter packets 0 bytes 0 reject
			meta l4proto udp ip daddr 169.254.169.254 udp dport 53 counter packets 0 bytes 0 reject
		}
	}
	`
	// https://bugzilla.redhat.com/show_bug.cgi?id=1915027
	openshiftMachineConfigIptables = `*filter
	:INPUT ACCEPT [0:0]
	:FORWARD ACCEPT [0:0]
	:OUTPUT ACCEPT [0:0]
	-A FORWARD -p tcp -m tcp --dport 22623 --tcp-flags FIN,SYN,RST,ACK SYN -j REJECT --reject-with icmp-port-unreachable
	-A FORWARD -p tcp -m tcp --dport 22624 --tcp-flags FIN,SYN,RST,ACK SYN -j REJECT --reject-with icmp-port-unreachable
	-A FORWARD -d 169.254.169.254/32 -p tcp -m tcp ! --dport 53 -j REJECT --reject-with icmp-port-unreachable
	-A FORWARD -d 169.254.169.254/32 -p udp -m udp ! --dport 53 -j REJECT --reject-with icmp-port-unreachable
	-A OUTPUT -p tcp -m tcp --dport 22623 --tcp-flags FIN,SYN,RST,ACK SYN -j REJECT --reject-with icmp-port-unreachable
	-A OUTPUT -p tcp -m tcp --dport 22624 --tcp-flags FIN,SYN,RST,ACK SYN -j REJECT --reject-with icmp-port-unreachable
	-A OUTPUT -d 169.254.169.254/32 -p tcp -m tcp ! --dport 53 -j REJECT --reject-with icmp-port-unreachable
	-A OUTPUT -d 169.254.169.254/32 -p udp -m udp ! --dport 53 -j REJECT --reject-with icmp-port-unreachable
	COMMIT
	`
)

func stripSpaceTabLine(in string) string {
	s1 := strings.ReplaceAll(in, "\n", "")
	s2 := strings.ReplaceAll(s1, "\t", "")
	return strings.ReplaceAll(s2, " ", "")
}
func isIPOrNSTablesPresent(cut *provider.Container, command string) (bool, string, error) { //nolint:gocritic
	outStr, errStr, err := crclient.ExecCommandContainerNSEnter(command, cut)
	if err != nil || (errStr != "" && errStr != ipTablesLegacyWarning) {
		return false, outStr, fmt.Errorf("failed to execute command %s on %s, err: %s, errStr: %s", command, cut, err, errStr)
	}

	if errStr == ipTablesLegacyWarning {
		return true, outStr, nil
	}
	if strings.Contains(stripSpaceTabLine(outStr), stripSpaceTabLine(openshiftMachineConfigNft)) ||
		strings.Contains(stripSpaceTabLine(outStr), stripSpaceTabLine(openshiftMachineConfigIptables)) {
		return false, outStr, nil
	}

	return outStr != "", outStr, nil
}

func IsNFTablesPresent(cut *provider.Container) (bool, string, error) { //nolint:gocritic
	return isIPOrNSTablesPresent(cut, dumpNFTablesCmd)
}

func IsIPTablesPresent(cut *provider.Container) (bool, string, error) { //nolint:gocritic
	return isIPOrNSTablesPresent(cut, dumpIPTablesCmd)
}

func IsIP6TablesPresent(cut *provider.Container) (bool, string, error) { //nolint:gocritic
	return isIPOrNSTablesPresent(cut, dumpIP6TablesCmd)
}
