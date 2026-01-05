// Copyright (C) 2020-2026 Red Hat, Inc.
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

package phasecheck

import (
	"testing"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_isOperatorSucceeded(t *testing.T) {
	type args struct {
		csv *v1alpha1.ClusterServiceVersion
	}
	tests := []struct {
		name        string
		args        args
		wantIsReady bool
	}{
		{name: "ok",
			args: args{csv: &v1alpha1.ClusterServiceVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aCSV",
					Namespace: "aNamespace",
				},
				Status: v1alpha1.ClusterServiceVersionStatus{
					Phase: v1alpha1.CSVPhaseSucceeded,
				},
			}},
			wantIsReady: true,
		},
		{name: "nok",
			args: args{csv: &v1alpha1.ClusterServiceVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aCSV",
					Namespace: "aNamespace",
				},
				Status: v1alpha1.ClusterServiceVersionStatus{
					Phase: v1alpha1.CSVPhaseInstalling,
				},
			}},
			wantIsReady: false,
		},
		{name: "nok",
			args: args{csv: &v1alpha1.ClusterServiceVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aCSV",
					Namespace: "aNamespace",
				},
				Status: v1alpha1.ClusterServiceVersionStatus{
					Phase: v1alpha1.CSVPhaseDeleting,
				},
			}},
			wantIsReady: false,
		},
		{name: "nok",
			args: args{csv: &v1alpha1.ClusterServiceVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aCSV",
					Namespace: "aNamespace",
				},
				Status: v1alpha1.ClusterServiceVersionStatus{
					Phase: v1alpha1.CSVPhaseInstallReady,
				},
			}},

			wantIsReady: false,
		},
		{name: "nok",
			args: args{csv: &v1alpha1.ClusterServiceVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aCSV",
					Namespace: "aNamespace",
				},
				Status: v1alpha1.ClusterServiceVersionStatus{
					Phase: v1alpha1.CSVPhaseUnknown,
				},
			}},
			wantIsReady: false,
		},
		{name: "nok",
			args: args{csv: &v1alpha1.ClusterServiceVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aCSV",
					Namespace: "aNamespace",
				},
				Status: v1alpha1.ClusterServiceVersionStatus{
					Phase: v1alpha1.CSVPhasePending,
				},
			}},
			wantIsReady: false,
		},
		{name: "nok",
			args: args{csv: &v1alpha1.ClusterServiceVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aCSV",
					Namespace: "aNamespace",
				},
				Status: v1alpha1.ClusterServiceVersionStatus{
					Phase: v1alpha1.CSVPhaseReplacing,
				},
			}},
			wantIsReady: false,
		},
		{name: "nok",
			args: args{csv: &v1alpha1.ClusterServiceVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aCSV",
					Namespace: "aNamespace",
				},
				Status: v1alpha1.ClusterServiceVersionStatus{
					Phase: v1alpha1.CSVPhaseAny,
				},
			}},
			wantIsReady: false,
		},
		{name: "nok",
			args: args{csv: &v1alpha1.ClusterServiceVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aCSV",
					Namespace: "aNamespace",
				},
				Status: v1alpha1.ClusterServiceVersionStatus{
					Phase: v1alpha1.CSVPhaseFailed,
				},
			}},

			wantIsReady: false,
		},
		{name: "nok",
			args: args{csv: &v1alpha1.ClusterServiceVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aCSV",
					Namespace: "aNamespace",
				},
				Status: v1alpha1.ClusterServiceVersionStatus{
					Phase: v1alpha1.CSVPhaseNone,
				},
			}},

			wantIsReady: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIsReady := isOperatorPhaseSucceeded(tt.args.csv); gotIsReady != tt.wantIsReady {
				t.Errorf("isOperatorSucceeded() = %v, want %v", gotIsReady, tt.wantIsReady)
			}
		})
	}
}

func Test_isOperatorFailedOrUnknown(t *testing.T) {
	type args struct {
		csv *v1alpha1.ClusterServiceVersion
	}
	tests := []struct {
		name        string
		args        args
		wantIsReady bool
	}{
		{name: "ok",
			args: args{csv: &v1alpha1.ClusterServiceVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aCSV",
					Namespace: "aNamespace",
				},
				Status: v1alpha1.ClusterServiceVersionStatus{
					Phase: v1alpha1.CSVPhaseFailed,
				},
			}},
			wantIsReady: true,
		},
		{name: "ok",
			args: args{csv: &v1alpha1.ClusterServiceVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aCSV",
					Namespace: "aNamespace",
				},
				Status: v1alpha1.ClusterServiceVersionStatus{
					Phase: v1alpha1.CSVPhaseUnknown,
				},
			}},
			wantIsReady: true,
		},
		{name: "nok",
			args: args{csv: &v1alpha1.ClusterServiceVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aCSV",
					Namespace: "aNamespace",
				},
				Status: v1alpha1.ClusterServiceVersionStatus{
					Phase: v1alpha1.CSVPhaseDeleting,
				},
			}},
			wantIsReady: false,
		},
		{name: "nok",
			args: args{csv: &v1alpha1.ClusterServiceVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "aCSV",
					Namespace: "aNamespace",
				},
				Status: v1alpha1.ClusterServiceVersionStatus{
					Phase: v1alpha1.CSVPhaseInstallReady,
				},
			}},

			wantIsReady: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotIsReady := isOperatorPhaseFailedOrUnknown(tt.args.csv); gotIsReady != tt.wantIsReady {
				t.Errorf("isOperatorFailedOrUnknown() = %v, want %v", gotIsReady, tt.wantIsReady)
			}
		})
	}
}
