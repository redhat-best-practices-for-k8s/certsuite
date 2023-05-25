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

package configuration

import (
	"reflect"
	"testing"
)

func Test_createLabels(t *testing.T) {
	type args struct {
		labelStrings []string
	}
	tests := []struct {
		name             string
		args             args
		wantLabelObjects []LabelObject
	}{
		{
			name:             "ok",
			args:             args{labelStrings: []string{"test-network-function.com/generic: target"}},
			wantLabelObjects: []LabelObject{{LabelKey: "test-network-function.com/generic", LabelValue: "target"}},
		},
		{
			name:             "ok1",
			args:             args{labelStrings: []string{"test-network-function.com/generic   : 1"}},
			wantLabelObjects: []LabelObject{{LabelKey: "test-network-function.com/generic", LabelValue: "1"}},
		},
		{
			name: "nok",
			args: args{labelStrings: []string{"test-network-function.com/generic= target"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotLabelObjects := createLabels(tt.args.labelStrings); !reflect.DeepEqual(gotLabelObjects, tt.wantLabelObjects) {
				t.Errorf("createLabels() = %v, want %v", gotLabelObjects, tt.wantLabelObjects)
			}
		})
	}
}
