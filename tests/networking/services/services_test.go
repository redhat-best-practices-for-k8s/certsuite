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

package services

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/tests/networking/netcommons"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetServiceIPVersion(t *testing.T) {
	type args struct {
		aService *corev1.Service
	}
	tests := []struct {
		name       string
		args       args
		wantResult netcommons.IPVersion
		wantErr    bool
	}{
		// TODO: Add test cases.
		{
			name:       "dual-stack-ok1",
			args:       args{aService: createService([]string{"1.1.1.1", "fd00:10:96::d789"}, corev1.IPFamilyPolicyPreferDualStack)},
			wantResult: netcommons.IPv4v6,
			wantErr:    false,
		},
		{
			name:       "dual-stack-ok2",
			args:       args{aService: createService([]string{"1.1.1.1", "fd00:10:96::d789"}, corev1.IPFamilyPolicyRequireDualStack)},
			wantResult: netcommons.IPv4v6,
			wantErr:    false,
		},
		{
			name:       "dual-stack-nok1",
			args:       args{aService: createService([]string{"1.1.1.1"}, corev1.IPFamilyPolicyPreferDualStack)},
			wantResult: netcommons.Undefined,
			wantErr:    true,
		},
		{
			name:       "dual-stack-nok2",
			args:       args{aService: createService([]string{"1.1.1.1", "2.2.2.2"}, corev1.IPFamilyPolicyPreferDualStack)},
			wantResult: netcommons.Undefined,
			wantErr:    true,
		},
		{
			name:       "single-stack-ipv6",
			args:       args{aService: createService([]string{"fd00:10:96::d789"}, corev1.IPFamilyPolicySingleStack)},
			wantResult: netcommons.IPv6,
			wantErr:    false,
		},
		{
			name:       "single-stack-ipv4",
			args:       args{aService: createService([]string{"1.1.1.1"}, corev1.IPFamilyPolicySingleStack)},
			wantResult: netcommons.IPv4,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := GetServiceIPVersion(tt.args.aService)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetServiceIPVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotResult != tt.wantResult {
				t.Errorf("GetServiceIPVersion() = %v, want %v", gotResult, tt.wantResult)
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

func TestToStringSlice(t *testing.T) {
	testCases := []struct {
		testSvcs []*corev1.Service
		expected string
	}{
		{
			testSvcs: []*corev1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-1",
						Namespace: "tnf",
					},
					Spec: corev1.ServiceSpec{
						ClusterIP:  "192.168.1.1",
						ClusterIPs: []string{"192.168.1.1"},
					},
				},
			},
			expected: "Service ns: tnf, name: test-service-1 ClusterIP:192.168.1.1 ClusterIPs: [192.168.1.1]\n",
		},
		{
			testSvcs: []*corev1.Service{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-service-1",
						Namespace: "tnf",
					},
					Spec: corev1.ServiceSpec{
						ClusterIP:  "192.168.1.2",
						ClusterIPs: []string{"192.168.1.2"},
					},
				},
			},
			expected: "Service ns: tnf, name: test-service-1 ClusterIP:192.168.1.2 ClusterIPs: [192.168.1.2]\n",
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, ToStringSlice(tc.testSvcs))
	}
}
