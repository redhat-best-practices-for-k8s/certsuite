// Copyright (C) 2020-2023 Red Hat, Inc.
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
package offlinecheck

import "testing"

func TestCompareVersion(t *testing.T) {
	type args struct {
		version    string
		constraint string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ok",
			args: args{version: "v1.25.4+18eadca", constraint: "v1.25.x"},
			want: true,
		},
		{
			name: "failed",
			args: args{version: "v1.25.4+18eadca", constraint: "v1.24.x"},
			want: false,
		},
		{
			name: "superior",
			args: args{version: "v1.25.4+18eadca", constraint: ">= v1.23.x"},
			want: true,
		},
		{
			name: "inferior",
			args: args{version: "v1.25.4+18eadca", constraint: "< v2.2.x"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompareVersion(tt.args.version, tt.args.constraint); got != tt.want {
				t.Errorf("CompareVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
