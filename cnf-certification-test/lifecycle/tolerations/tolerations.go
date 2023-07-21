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

package tolerations

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

var (
	nonCompliantTolerations  = []corev1.TaintEffect{corev1.TaintEffectNoExecute, corev1.TaintEffectNoSchedule, corev1.TaintEffectPreferNoSchedule}
	tolerationSecondsDefault = 300
)

func IsTolerationModified(t corev1.Toleration, qosClass corev1.PodQOSClass) bool {
	const (
		notReadyStr       = "node.kubernetes.io/not-ready"
		unreachableStr    = "node.kubernetes.io/unreachable"
		memoryPressureStr = "node.kubernetes.io/memory-pressure"
	)
	// Check each of the tolerations to make sure they are the default tolerations added by k8s:
	// tolerations:
	// - effect: NoExecute
	//   key: node.kubernetes.io/not-ready
	//   operator: Exists
	//   tolerationSeconds: 300
	// - effect: NoExecute
	//   key: node.kubernetes.io/unreachable
	//   operator: Exists
	//   tolerationSeconds: 300
	// # this last one, only if QoS class for the pod is different than BestEffort
	// - effect: NoSchedule
	//   key: node.kubernetes.io/memory-pressure
	//   operator: Exists

	// Short circuit.  Anything that is not 'node.kubernetes.io' is considered a modified toleration immediately.
	if !IsTolerationDefault(t) {
		return true
	}

	// Happy Path - This is detecting a default toleration
	if t.Effect == corev1.TaintEffectNoExecute {
		if t.Key == notReadyStr || t.Key == unreachableStr {
			// 300 seconds is the default, return false for not modified
			if t.Operator == corev1.TolerationOpExists && t.TolerationSeconds != nil && *t.TolerationSeconds == int64(tolerationSecondsDefault) {
				return false
			}

			// Toleration seconds has been modified, return true.
			return true
		}
	} else if t.Effect == corev1.TaintEffectNoSchedule {
		// If toleration is NoSchedule - node.kubernetes.io/memory-pressure - Exists and the QoS class for
		// the pod is different than BestEffort, it is also a default toleration added by k8s
		if (t.Key == memoryPressureStr) &&
			(t.Operator == corev1.TolerationOpExists) &&
			(qosClass != corev1.PodQOSBestEffort) {
			return false
		}
	}

	// Check through the list of non-compliant tolerations to see if anything snuck by the above short circuit
	for _, nct := range nonCompliantTolerations {
		if t.Effect == nct {
			return true
		}
	}

	return false
}

func IsTolerationDefault(t corev1.Toleration) bool {
	return strings.Contains(t.Key, "node.kubernetes.io")
}
