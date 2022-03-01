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

package icmp

import (
	"reflect"
	"testing"

	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
)

func Test_parsePingResult(t *testing.T) { //nolint:funlen
	type args struct {
		stdout string
		stderr string
	}
	tests := []struct {
		name        string
		args        args
		wantResults PingResults
		wantErr     bool
	}{
		{
			name: "pingOk",
			args: args{stdout: `PING 8.8.8.8 (8.8.8.8) 56(84) bytes of data.
		 64 bytes from 8.8.8.8: icmp_seq=1 ttl=57 time=36.1 ms
		 64 bytes from 8.8.8.8: icmp_seq=2 ttl=57 time=32.6 ms
		 64 bytes from 8.8.8.8: icmp_seq=3 ttl=57 time=35.9 ms
		 64 bytes from 8.8.8.8: icmp_seq=4 ttl=57 time=38.2 ms
		 64 bytes from 8.8.8.8: icmp_seq=5 ttl=57 time=36.0 ms
		 
		 --- 8.8.8.8 ping statistics ---
		 5 packets transmitted, 5 received, 0% packet loss, time 4005ms
		 rtt min/avg/max/mdev = 32.593/35.761/38.212/1.802 ms`, stderr: ""},
			wantResults: PingResults{outcome: testhelper.SUCCESS, transmitted: 5, received: 5, errors: 0},
			wantErr:     false,
		},
		{
			name: "pingErrorPacket",
			args: args{stdout: `PING 192.168.1.1 (192.168.1.1) 56(84) bytes of data.
			64 bytes from 192.168.1.1: icmp_seq=1 ttl=61 time=1.79 ms
			64 bytes from 192.168.1.1: icmp_seq=2 ttl=61 time=3.37 ms
			64 bytes from 192.168.1.1: icmp_seq=3 ttl=61 time=2.14 ms
			64 bytes from 192.168.1.1: icmp_seq=4 ttl=61 time=3.62 ms
			From 10.0.2.2 icmp_seq=5 Destination Net Unreachable
			From 10.0.2.2 icmp_seq=6 Destination Net Unreachable
			From 10.0.2.2 icmp_seq=7 Destination Net Unreachable
			From 10.0.2.2 icmp_seq=8 Destination Net Unreachable
			64 bytes from 192.168.1.1: icmp_seq=9 ttl=61 time=297 ms
			64 bytes from 192.168.1.1: icmp_seq=10 ttl=61 time=258 ms
			64 bytes from 192.168.1.1: icmp_seq=11 ttl=61 time=276 ms
			64 bytes from 192.168.1.1: icmp_seq=12 ttl=61 time=1.58 ms
			64 bytes from 192.168.1.1: icmp_seq=13 ttl=61 time=445 ms
			64 bytes from 192.168.1.1: icmp_seq=14 ttl=61 time=3.57 ms
			64 bytes from 192.168.1.1: icmp_seq=15 ttl=61 time=60.5 ms
			64 bytes from 192.168.1.1: icmp_seq=16 ttl=61 time=585 ms
			64 bytes from 192.168.1.1: icmp_seq=17 ttl=61 time=155 ms
			64 bytes from 192.168.1.1: icmp_seq=18 ttl=61 time=20.4 ms
			64 bytes from 192.168.1.1: icmp_seq=19 ttl=61 time=26.7 ms
			64 bytes from 192.168.1.1: icmp_seq=20 ttl=61 time=2.14 ms
			
			--- 192.168.1.1 ping statistics ---
			20 packets transmitted, 16 received, +4 errors, 20% packet loss, time 19118ms
			rtt min/avg/max/mdev = 1.582/134.079/585.861/179.394 ms`, stderr: ""},
			wantResults: PingResults{outcome: testhelper.ERROR, transmitted: 20, received: 16, errors: 4},
			wantErr:     false,
		},
		{
			name: "pingIncorrectIp",
			args: args{stdout: `connect: Invalid argument
			command terminated with exit code 2`, stderr: ""},
			wantResults: PingResults{outcome: testhelper.ERROR, transmitted: 0, received: 0, errors: 0},
			wantErr:     true,
		},
		{
			name: "pingPassingPacketLoss",
			args: args{stdout: `PING 192.168.1.5 (192.168.1.5) 56(84) bytes of data.
			64 bytes from 192.168.1.5: icmp_seq=1 ttl=61 time=14.8 ms
			64 bytes from 192.168.1.5: icmp_seq=2 ttl=61 time=11.2 ms
			64 bytes from 192.168.1.5: icmp_seq=3 ttl=61 time=10.9 ms
			64 bytes from 192.168.1.5: icmp_seq=5 ttl=61 time=9.68 ms
			64 bytes from 192.168.1.5: icmp_seq=6 ttl=61 time=4.55 ms
			64 bytes from 192.168.1.5: icmp_seq=7 ttl=61 time=3.38 ms
			64 bytes from 192.168.1.5: icmp_seq=8 ttl=61 time=3.67 ms
			64 bytes from 192.168.1.5: icmp_seq=9 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=10 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=11 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=12 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=13 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=14 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=15 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=16 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=17 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=18 ttl=61 time=3.77 ms
			64 bytes from 192.168.1.5: icmp_seq=19 ttl=61 time=3.77 ms
			
			--- 192.168.1.5 ping statistics ---
			20 packets transmitted, 19 received, 5% packet loss, time 19297ms
			rtt min/avg/max/mdev = 3.381/7.772/14.867/4.167 ms`, stderr: ""},
			wantResults: PingResults{outcome: testhelper.SUCCESS, transmitted: 20, received: 19, errors: 0},
			wantErr:     false,
		},
		{
			name: "pingFailingPacketLoss",
			args: args{stdout: `
			PING 192.168.1.2 (192.168.1.2) 56(84) bytes of data.
			
			--- 192.168.1.2 ping statistics ---
			1 packets transmitted, 0 received, 100% packet loss, time 0ms
			
			command terminated with exit code 1`, stderr: ""},
			wantResults: PingResults{outcome: testhelper.FAILURE, transmitted: 1, received: 0, errors: 0},
			wantErr:     false,
		},
		{
			name: "pingHostnameNoPacketLoss",
			args: args{stdout: `PING www.google.com (172.217.12.132) 56(84) bytes of data.
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=1 ttl=61 time=25.4 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=2 ttl=61 time=27.1 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=3 ttl=61 time=26.7 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=4 ttl=61 time=24.2 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=5 ttl=61 time=28.0 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=6 ttl=61 time=37.0 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=7 ttl=61 time=21.6 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=8 ttl=61 time=30.6 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=9 ttl=61 time=27.3 ms
			64 bytes from lga34s19-in-f4.1e100.net (172.217.12.132): icmp_seq=10 ttl=61 time=27.9 ms
			
			--- www.google.com ping statistics ---
			10 packets transmitted, 10 received, 0% packet loss, time 9014ms
			rtt min/avg/max/mdev = 21.650/27.619/37.003/3.885 ms`, stderr: ""},
			wantResults: PingResults{outcome: testhelper.SUCCESS, transmitted: 10, received: 10, errors: 0},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResults, err := parsePingResult(tt.args.stdout, tt.args.stderr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePingResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResults, tt.wantResults) {
				t.Errorf("parsePingResult() = %v, want %v", gotResults, tt.wantResults)
			}
		})
	}
}
