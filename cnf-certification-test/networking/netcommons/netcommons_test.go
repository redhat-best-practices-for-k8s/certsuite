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

package netcommons

import (
	"reflect"
	"testing"
)

func Test_getIPVersion(t *testing.T) {
	type args struct {
		aIP string
	}
	tests := []struct {
		name    string
		args    args
		want    IPVersion
		wantErr bool
	}{
		{name: "GoodIPv4",
			args:    args{aIP: "2.2.2.2"},
			want:    IPv4,
			wantErr: false,
		},
		{name: "GoodIPv6",
			args:    args{aIP: "fd00:10:244:1::3"},
			want:    IPv6,
			wantErr: false,
		},
		{name: "BadIPv4",
			args:    args{aIP: "2.hfh.2.2"},
			want:    "",
			wantErr: true,
		},
		{name: "BadIPv6",
			args:    args{aIP: "fd00:10:ono;ogmo:1::3"},
			want:    "",
			wantErr: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetIPVersion(tt.args.aIP)
			if (err != nil) != tt.wantErr {
				t.Errorf("getIPVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getIPVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterIPListByIPVersion(t *testing.T) {
	type args struct {
		ipList     []string
		aIPVersion IPVersion
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "okIpv4",
			args: args{ipList: []string{"1.1.1.1", "2.2.2.2", "fd00:10:244:1::3"}, aIPVersion: IPv4},
			want: []string{"1.1.1.1", "2.2.2.2"},
		},
		{
			name: "okIpv6",
			args: args{ipList: []string{"1.1.1.1", "2.2.2.2", "fd00:10:244:1::3"}, aIPVersion: IPv6},
			want: []string{"fd00:10:244:1::3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterIPListByIPVersion(tt.args.ipList, tt.args.aIPVersion); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterIPListByIPVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
