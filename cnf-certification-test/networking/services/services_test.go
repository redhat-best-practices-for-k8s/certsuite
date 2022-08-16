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

package services

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestIsDualStack(t *testing.T) { //nolint:funlen
	type args struct {
		aService *corev1.Service
	}
	tests := []struct {
		name       string
		args       args
		wantResult bool
		wantErr    bool
	}{
		// TODO: Add test cases.
		{
			name:       "dual-stack-ok1",
			args:       args{aService: createService([]string{"1.1.1.1", "fd00:10:96::d789"}, corev1.IPFamilyPolicyPreferDualStack)},
			wantResult: true,
			wantErr:    false,
		},
		{
			name:       "dual-stack-ok2",
			args:       args{aService: createService([]string{"1.1.1.1", "fd00:10:96::d789"}, corev1.IPFamilyPolicyRequireDualStack)},
			wantResult: true,
			wantErr:    false,
		},
		{
			name:       "dual-stack-nok1",
			args:       args{aService: createService([]string{"1.1.1.1"}, corev1.IPFamilyPolicyPreferDualStack)},
			wantResult: false,
			wantErr:    true,
		},
		{
			name:       "dual-stack-nok2",
			args:       args{aService: createService([]string{"1.1.1.1", "2.2.2.2"}, corev1.IPFamilyPolicyPreferDualStack)},
			wantResult: false,
			wantErr:    true,
		},
		{
			name:       "single-stack-ok1",
			args:       args{aService: createService([]string{"fd00:10:96::d789"}, corev1.IPFamilyPolicySingleStack)},
			wantResult: true,
			wantErr:    false,
		},
		{
			name:       "single-stack-nok1",
			args:       args{aService: createService([]string{"1.1.1.1"}, corev1.IPFamilyPolicySingleStack)},
			wantResult: false,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := IsDualStack(tt.args.aService)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsDualStack() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotResult != tt.wantResult {
				t.Errorf("IsDualStack() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func createService(ips []string, aFp corev1.IPFamilyPolicyType) (aService *corev1.Service) {
	aService = &corev1.Service{}
	aService.Name = "test-service"
	aService.Namespace = "tnf"
	aService.Spec.ClusterIP = ips[0]
	aService.Spec.ClusterIPs = ips
	aService.Spec.IPFamilyPolicy = &aFp
	return aService
}
