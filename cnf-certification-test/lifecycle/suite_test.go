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

package lifecycle

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

//nolint:funlen
func TestTestPodTaintsTolerationHelper(t *testing.T) {
	pod := v1.Pod{}
	pod.Spec.Tolerations = append(pod.Spec.Tolerations,
		v1.Toleration{
			Key:      "test",
			Value:    "test",
			Operator: v1.TolerationOpEqual,
			Effect:   v1.TaintEffectNoExecute},
	)
	tolerations := getPodTaintsTolerations(&pod)
	assert.NotEmpty(t, tolerations)

	pod.Spec.Tolerations = []v1.Toleration{}
	pod.Spec.Tolerations = append(pod.Spec.Tolerations,
		v1.Toleration{
			Key:      "test",
			Value:    "test",
			Operator: v1.TolerationOpEqual,
			Effect:   v1.TaintEffectNoSchedule},
	)
	tolerations = getPodTaintsTolerations(&pod)
	assert.NotEmpty(t, tolerations)

	pod.Spec.Tolerations = []v1.Toleration{}
	pod.Spec.Tolerations = append(pod.Spec.Tolerations,
		v1.Toleration{
			Key:      "test",
			Value:    "test",
			Operator: v1.TolerationOpEqual,
			Effect:   v1.TaintEffectPreferNoSchedule},
	)
	tolerations = getPodTaintsTolerations(&pod)
	assert.NotEmpty(t, tolerations)

	pod.Spec.Tolerations = []v1.Toleration{}
	pod.Spec.Tolerations = append(pod.Spec.Tolerations,
		v1.Toleration{
			Key:      "test",
			Value:    "test",
			Operator: v1.TolerationOpEqual,
			Effect:   v1.TaintNodeDiskPressure},
		v1.Toleration{
			Key:      "test",
			Value:    "test",
			Operator: v1.TolerationOpEqual,
			Effect:   v1.TaintNodeMemoryPressure},
	)
	tolerations = getPodTaintsTolerations(&pod)
	assert.Empty(t, tolerations)
}
