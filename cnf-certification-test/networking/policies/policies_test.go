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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//nolint:funlen
func TestIsNetworkPolicyCompliant(t *testing.T) {
	testCases := []struct {
		testNP                networkingv1.NetworkPolicy
		expectedPolicies      []networkingv1.PolicyType
		expectedIngressOutput error
		expectedEgressOutput  error
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
			expectedPolicies:      nil,
			expectedIngressOutput: errors.New("PolicyTypes is empty"),
			expectedEgressOutput:  errors.New("PolicyTypes is empty"),
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
			expectedIngressOutput: nil,
			expectedEgressOutput:  nil,
		},
		{ // Test #3 - Network Policy with label selector with no egress policytype
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
						// networkingv1.PolicyTypeEgress,
						networkingv1.PolicyTypeIngress,
					},
				},
			},
			expectedPolicies: []networkingv1.PolicyType{
				networkingv1.PolicyTypeEgress,
				networkingv1.PolicyTypeIngress,
			},
			expectedIngressOutput: nil,
			expectedEgressOutput:  errors.New("deny all network policy not found"),
		},
		{ // Test #4 - Network Policy with label selector with no ingress policytype
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
						// networkingv1.PolicyTypeIngress,
					},
				},
			},
			expectedPolicies: []networkingv1.PolicyType{
				networkingv1.PolicyTypeEgress,
				networkingv1.PolicyTypeIngress,
			},
			expectedIngressOutput: errors.New("deny all network policy not found"),
			expectedEgressOutput:  nil,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedEgressOutput, IsNetworkPolicyCompliant(&tc.testNP, networkingv1.PolicyTypeEgress))
		assert.Equal(t, tc.expectedIngressOutput, IsNetworkPolicyCompliant(&tc.testNP, networkingv1.PolicyTypeIngress))
	}
}
