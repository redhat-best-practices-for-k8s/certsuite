// Copyright (C) 2022-2023 Red Hat, Inc.
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
	networkingv1 "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//nolint:gocritic // unnamed results
func IsNetworkPolicyCompliant(np *networkingv1.NetworkPolicy, policyType networkingv1.PolicyType) (bool, string) {
	// As long as we have decided above that there is no pod selector,
	// we just have to make sure that the policy type is either Ingress or Egress (or both) we can return true.
	// For more information about deny-all policies, there are some good examples on:
	// https://kubernetes.io/docs/concepts/services-networking/network-policies/

	if len(np.Spec.PolicyTypes) == 0 {
		return false, "empty policy types"
	}

	// Ingress and Egress rules should be "empty" if it is a default rule.
	if policyType == networkingv1.PolicyTypeEgress {
		if np.Spec.Egress != nil || len(np.Spec.Egress) > 0 {
			return false, "egress spec not empty for default egress rule"
		}
	}

	if policyType == networkingv1.PolicyTypeIngress {
		if np.Spec.Ingress != nil || len(np.Spec.Ingress) > 0 {
			return false, "ingress spec not empty for default ingress rule"
		}
	}

	policyTypeFound := false
	// Look through the returned policies to see if they match the desired policyType
	for _, p := range np.Spec.PolicyTypes {
		if p == policyType {
			policyTypeFound = true
			break
		}
	}

	return policyTypeFound, ""
}

func LabelsMatch(podSelectorLabels v1.LabelSelector, podLabels map[string]string) bool {
	labelMatch := false

	// When the pod selector label is empty, it will always match the pod
	if podSelectorLabels.Size() == 0 {
		return true
	}

	for psLabelKey, psLabelValue := range podSelectorLabels.MatchLabels {
		for podLabelKey, podLabelValue := range podLabels {
			if psLabelKey == podLabelKey && psLabelValue == podLabelValue {
				labelMatch = true
				break
			}
		}
		if labelMatch {
			break
		}
	}

	return labelMatch
}
