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
			name: "pingOk",
			args: args{stdout: `PING 8.8.8.8 (8.8.8.8) 56(84) bytes of data.
		 64 bytes from 8.8.8.8: icmp_seq=1 ttl=57 time=36.1 ms
		 64 bytes from 8.8.8.8: icmp_seq=2 ttl=57 time=32.6 ms
		 64 bytes from 8.8.8.8: icmp_seq=3 ttl=57 time=35.9 ms
		 64 bytes from 8.8.8.8: icmp_seq=4 ttl=57 time=38.2 ms

		 
		 --- 8.8.8.8 ping statistics ---
		 5 packets transmitted, 4 received, 0% packet loss, time 4005ms
		 rtt min/avg/max/mdev = 32.593/35.761/38.212/1.802 ms`, stderr: ""},
			wantResults: PingResults{outcome: testhelper.SUCCESS, transmitted: 5, received: 4, errors: 0},
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
