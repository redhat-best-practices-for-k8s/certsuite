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

package provider

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
)

func Test_isDaemonSetReady(t *testing.T) {
	type args struct {
		status *appsv1.DaemonSetStatus
	}
	tests := []struct {
		name        string
		args        args
		wantIsReady bool
	}{
		{name: "daemonsetReady",
			args: args{status: &appsv1.DaemonSetStatus{
				CurrentNumberScheduled: 4, NumberMisscheduled: 0, DesiredNumberScheduled: 4,
				NumberReady: 4, ObservedGeneration: 0, UpdatedNumberScheduled: 0,
				NumberAvailable: 4, NumberUnavailable: 0, CollisionCount: nil, Conditions: nil,
			},
			}, wantIsReady: true,
		},
		{name: "daemonsetNotReady1",
			args: args{status: &appsv1.DaemonSetStatus{
				CurrentNumberScheduled: 4, NumberMisscheduled: 0, DesiredNumberScheduled: 4,
				NumberReady: 4, ObservedGeneration: 0, UpdatedNumberScheduled: 0,
				NumberAvailable: 3, NumberUnavailable: 0, CollisionCount: nil, Conditions: nil,
			},
			}, wantIsReady: false,
		},
		{name: "daemonsetNotReady2",
			args: args{status: &appsv1.DaemonSetStatus{
				CurrentNumberScheduled: 4, NumberMisscheduled: 1, DesiredNumberScheduled: 4,
				NumberReady: 4, ObservedGeneration: 0, UpdatedNumberScheduled: 0,
				NumberAvailable: 4, NumberUnavailable: 0, CollisionCount: nil, Conditions: nil,
			},
			}, wantIsReady: false,
		},
		{name: "daemonsetNotReady3",
			args: args{status: &appsv1.DaemonSetStatus{
				CurrentNumberScheduled: 4, NumberMisscheduled: 0, DesiredNumberScheduled: 4,
				NumberReady: 3, ObservedGeneration: 0, UpdatedNumberScheduled: 0,
				NumberAvailable: 4, NumberUnavailable: 0, CollisionCount: nil, Conditions: nil,
			},
			}, wantIsReady: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIsReady := isDaemonSetReady(tt.args.status); gotIsReady != tt.wantIsReady {
				t.Errorf("isDaemonSetReady() = %v, want %v", gotIsReady, tt.wantIsReady)
			}
		})
	}
}
