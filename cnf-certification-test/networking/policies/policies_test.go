// Copyright (C) 2022 Red Hat, Inc.
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

package policies

import (
	"testing"

	"github.com/stretchr/testify/assert"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//nolint:funlen
func TestIsNetworkPolicyDenyAll(t *testing.T) {
	policyInSlice := func(s []networkingv1.PolicyType, pt networkingv1.PolicyType) bool {
		for _, v := range s {
			if v == pt {
				return true
			}
		}
		return false
	}

	testCases := []struct {
		testNP           networkingv1.NetworkPolicy
		expectedPolicies []networkingv1.PolicyType
		expectedOutput   bool
	}{
		{ // Test #1 - Network Policy with no label selector, no policy types, fails.
			testNP: networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test1",
				},
				Spec: networkingv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"key1": "value1",
						},
					},
				},
			},
			expectedPolicies: nil,
			expectedOutput:   false,
		},
		{ // Test #2 - Network Policy with label selector, and both ingress/egress policy types
			testNP: networkingv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test2",
				},
				Spec: networkingv1.NetworkPolicySpec{
					PodSelector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"key1": "value1",
						},
					},
					PolicyTypes: []networkingv1.PolicyType{
						networkingv1.PolicyTypeEgress,
						networkingv1.PolicyTypeIngress,
					},
				},
			},
			expectedPolicies: []networkingv1.PolicyType{
				networkingv1.PolicyTypeEgress,
				networkingv1.PolicyTypeIngress,
			},
			expectedOutput: true,
		},
	}

	for _, tc := range testCases {
		result, policyResult := IsNetworkPolicyDenyAll(&tc.testNP)
		assert.Equal(t, tc.expectedOutput, result)

		for _, pr := range policyResult {
			assert.True(t, policyInSlice(tc.expectedPolicies, pr))
		}
	}
}
