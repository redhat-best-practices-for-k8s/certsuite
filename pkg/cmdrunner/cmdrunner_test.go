// Copyright (C) 2020-2021 Red Hat, Inc.
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

package cmdrunner

import "testing"

func Test_runLocalCommand(t *testing.T) {
	type args struct {
		command string
	}
	tests := []struct {
		name       string
		args       args
		wantOutStr string
		wantErrStr string
		wantErr    bool
	}{
		{
			name:       "echo",
			args:       args{command: "echo test"},
			wantOutStr: "test\n",
			wantErrStr: "",
			wantErr:    false,
		},
		{
			name:       "echobad",
			args:       args{command: "echobad test"},
			wantOutStr: "",
			wantErrStr: "",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOutStr, gotErrStr, err := RunLocalCommand(tt.args.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("runLocalCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOutStr != tt.wantOutStr {
				t.Errorf("runLocalCommand() gotOutStr = %v, want %v", gotOutStr, tt.wantOutStr)
			}
			if gotErrStr != tt.wantErrStr {
				t.Errorf("runLocalCommand() gotErrStr = %v, want %v", gotErrStr, tt.wantErrStr)
			}
		})
	}
}
