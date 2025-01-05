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

package accesscontrol

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func Test_isContainerCapabilitySet(t *testing.T) {
	type args struct {
		containerCapabilities *corev1.Capabilities
		capability            string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil capabilities",
			args: args{
				capability:            "IPC_LOCK",
				containerCapabilities: nil,
			},
			want: false,
		},
		{
			name: "empty capabilities",
			args: args{
				capability:            "IPC_LOCK",
				containerCapabilities: &corev1.Capabilities{},
			},
			want: false,
		},
		{
			name: "explicitly empty add list",
			args: args{
				capability:            "IPC_LOCK",
				containerCapabilities: &corev1.Capabilities{Add: []corev1.Capability{}},
			},
			want: false,
		},
		{
			name: "explicitly empty drop list",
			args: args{
				capability:            "IPC_LOCK",
				containerCapabilities: &corev1.Capabilities{Drop: []corev1.Capability{}},
			},
			want: false,
		},
		{
			name: "IPC_LOCK not found in any list",
			args: args{
				capability: "IPC_LOCK",
				containerCapabilities: &corev1.Capabilities{
					Add:  []corev1.Capability{"NET_CAP_BINDING"},
					Drop: []corev1.Capability{"SYS_ADMIN", "NET_ADMIN"},
				},
			},
			want: false,
		},
		{
			name: "IPC_LOCK found in the add list",
			args: args{
				capability: "IPC_LOCK",
				containerCapabilities: &corev1.Capabilities{
					Add:  []corev1.Capability{"NET_ADMIN", "IPC_LOCK"},
					Drop: []corev1.Capability{"SYS_ADMIN"}},
			},
			want: true,
		},
		{
			name: "IPC_LOCK appears in the drop list only",
			args: args{
				capability:            "IPC_LOCK",
				containerCapabilities: &corev1.Capabilities{Drop: []corev1.Capability{"SYS_ADMIN", "IPC_LOCK", "NET_ADMIN"}},
			},
			want: false,
		},
		{
			// When set in both add and drop lists, k8s/openshift will compute drop first, then add, which results
			// in the capability to be finally set.
			name: "IPC_LOCK set in both add and drop lists.",
			args: args{
				capability: "IPC_LOCK",
				containerCapabilities: &corev1.Capabilities{
					Add:  []corev1.Capability{"IPC_LOCK"},
					Drop: []corev1.Capability{"SYS_ADMIN", "IPC_LOCK", "NET_ADMIN"},
				},
			},
			want: true,
		},
		{
			name: "ALL capabilities in the add list",
			args: args{
				capability:            "IPC_LOCK",
				containerCapabilities: &corev1.Capabilities{Add: []corev1.Capability{"ALL"}},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isContainerCapabilitySet(tt.args.containerCapabilities, tt.args.capability); got != tt.want {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
