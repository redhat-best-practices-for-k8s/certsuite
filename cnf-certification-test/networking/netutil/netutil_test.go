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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseListeningPorts(t *testing.T) {
	testCases := []struct {
		inputStr               string
		expectedListeningPorts map[PortInfo]bool
	}{
		{
			inputStr:               "tcp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\n",
			expectedListeningPorts: map[PortInfo]bool{{PortNumber: 8080, Protocol: "TCP"}: true},
		},
		{
			inputStr:               "",
			expectedListeningPorts: map[PortInfo]bool{},
		},
		{
			inputStr:               "\n",
			expectedListeningPorts: map[PortInfo]bool{},
		},
		{
			inputStr:               "tcp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\ntcp LISTEN 0      128    0.0.0.0:7878 0.0.0.0:*\n",
			expectedListeningPorts: map[PortInfo]bool{{PortNumber: 8080, Protocol: "TCP"}: true, {PortNumber: 7878, Protocol: "TCP"}: true},
		},
		{
			inputStr:               "udp LISTEN 0      128    0.0.0.0:8080 0.0.0.0:*\nudp LISTEN 0      128    0.0.0.0:7878 0.0.0.0:*\n",
			expectedListeningPorts: map[PortInfo]bool{{PortNumber: 8080, Protocol: "UDP"}: true, {PortNumber: 7878, Protocol: "UDP"}: true},
		},
		{
			inputStr:               "tcp LISTEN 0      128    [::]:22\n",
			expectedListeningPorts: map[PortInfo]bool{{PortNumber: 22, Protocol: "TCP"}: true},
		},
	}
	for _, tc := range testCases {
		listeningPorts, err := parseListeningPorts(tc.inputStr)
		if assert.NoError(t, err) {
			assert.Equal(t, tc.expectedListeningPorts, listeningPorts)
		}
	}
}

var nonZeroOpenshiftMachineConfigIPtables = `*filter
:INPUT ACCEPT [876:9889]
:FORWARD ACCEPT [333:0]
:OUTPUT ACCEPT [0:999]
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

func Test_zeroCounters(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name    string
		args    args
		wantOut string
	}{
		{
			name:    "ok",
			args:    args{in: stripSpaceTabLine(nonZeroOpenshiftMachineConfigIPtables)},
			wantOut: stripSpaceTabLine(openshiftMachineConfigIPtables),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOut := zeroCounters(tt.args.in); gotOut != tt.wantOut {
				t.Errorf("zeroCounters() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}
