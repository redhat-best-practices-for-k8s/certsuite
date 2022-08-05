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
	networkingv1 "k8s.io/api/networking/v1"
)

func IsNetworkPolicyDenyAll(np *networkingv1.NetworkPolicy) (bool, []networkingv1.PolicyType) {
	if len(np.Spec.PodSelector.MatchLabels) != 0 {
		return false, nil
	}

	// As long as we have decided above that there is no pod selector,
	// we just have to make sure that the policy type is either Ingress or Egress (or both) we can return true.
	// For more information about deny-all policies, there are some good examples on:
	// https://kubernetes.io/docs/concepts/services-networking/network-policies/

	if len(np.Spec.PolicyTypes) == 0 {
		return false, nil
	}

	// Ingress and Egress rules should be "empty" if it is a default rule.
	if np.Spec.Egress != nil {
		return false, nil
	}
	if np.Spec.Ingress != nil {
		return false, nil
	}

	return true, np.Spec.PolicyTypes
}
