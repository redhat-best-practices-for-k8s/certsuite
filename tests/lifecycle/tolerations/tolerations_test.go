// Copyright (C) 2022-2024 Red Hat, Inc.
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
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestIsTolerationModified(t *testing.T) {
	getInt64Pointer := func(val int64) *int64 {
		return &val
	}

	testCases := []struct {
		testToleration corev1.Toleration
		expectedOutput bool
		qosClass       corev1.PodQOSClass
	}{
		{ // Test Case #1 - default not-ready toleration
			testToleration: corev1.Toleration{
				Key:               "node.kubernetes.io/not-ready",
				Operator:          corev1.TolerationOpExists,
				Effect:            corev1.TaintEffectNoExecute,
				TolerationSeconds: getInt64Pointer(300),
			},
			expectedOutput: false,
			qosClass:       corev1.PodQOSGuaranteed,
		},
		{ // Test Case #2 - default unreachable toleration
			testToleration: corev1.Toleration{
				Key:               "node.kubernetes.io/unreachable",
				Operator:          corev1.TolerationOpExists,
				Effect:            corev1.TaintEffectNoExecute,
				TolerationSeconds: getInt64Pointer(300),
			},
			expectedOutput: false,
			qosClass:       corev1.PodQOSGuaranteed,
		},
		{ // Test Case #3 - modified unreachable toleration
			testToleration: corev1.Toleration{
				Key:               "node.kubernetes.io/unreachable",
				Operator:          corev1.TolerationOpExists,
				Effect:            corev1.TaintEffectNoExecute,
				TolerationSeconds: getInt64Pointer(350), // modified from 300
			},
			expectedOutput: true,
			qosClass:       corev1.PodQOSGuaranteed,
		},
		{ // Test Case #4 - modified unreachable toleration
			testToleration: corev1.Toleration{
				Key:               "node.kubernetes.io/unreachable",
				Operator:          corev1.TolerationOpEqual, // modified from exists
				Effect:            corev1.TaintEffectNoExecute,
				TolerationSeconds: getInt64Pointer(300),
			},
			expectedOutput: true,
			qosClass:       corev1.PodQOSGuaranteed,
		},
		{ // Test Case #5 - missing effect
			testToleration: corev1.Toleration{
				Key:      "node.kubernetes.io/unreachable",
				Operator: corev1.TolerationOpExists,
				// Effect:            corev1.TaintEffectNoExecute,
				TolerationSeconds: getInt64Pointer(300),
			},
			expectedOutput: false,
			qosClass:       corev1.PodQOSGuaranteed,
		},
		{ // Test Case #6 - example from QE and DCI - this should pass only if qosClass is
			// different than BestEffort, which is the case
			testToleration: corev1.Toleration{
				Key:      "node.kubernetes.io/memory-pressure",
				Operator: corev1.TolerationOpExists,
				Effect:   corev1.TaintEffectNoSchedule,
			},
			expectedOutput: false,
			qosClass:       corev1.PodQOSGuaranteed,
		},
		{ // Test Case #7 - example from QE and DCI - however, if qosClass is BestEffort, it
			// must be considered as a modified toleration
			testToleration: corev1.Toleration{
				Key:      "node.kubernetes.io/memory-pressure",
				Operator: corev1.TolerationOpExists,
				Effect:   corev1.TaintEffectNoSchedule,
			},
			expectedOutput: true,
			qosClass:       corev1.PodQOSBestEffort,
		},
		{ // Test Case #8 - Custom toleration - fails due to using a NoExecute taint
			testToleration: corev1.Toleration{
				Key:               "custom-toleration/test1",
				Operator:          corev1.TolerationOpExists,
				Effect:            corev1.TaintEffectNoExecute,
				TolerationSeconds: getInt64Pointer(300),
			},
			expectedOutput: true,
			qosClass:       corev1.PodQOSGuaranteed,
		},
		{ // Test Case #9 - not-ready toleration added with 301 (not default) toleration seconds
			testToleration: corev1.Toleration{
				Key:               "node.kubernetes.io/not-ready",
				Operator:          corev1.TolerationOpExists,
				Effect:            corev1.TaintEffectNoExecute,
				TolerationSeconds: getInt64Pointer(301),
			},
			expectedOutput: true,
			qosClass:       corev1.PodQOSGuaranteed,
		},
		{ // Test Case #10 - default not-ready toleration with unset toleration seconds, modified from default so failure
			testToleration: corev1.Toleration{
				Key:      "node.kubernetes.io/not-ready",
				Operator: corev1.TolerationOpExists,
				Effect:   corev1.TaintEffectNoExecute,
				// TolerationSeconds: getInt64Pointer(300),
			},
			expectedOutput: true,
			qosClass:       corev1.PodQOSGuaranteed,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, IsTolerationModified(tc.testToleration, tc.qosClass))
	}
}

func TestIsTolerationDefault(t *testing.T) {
	testCases := []struct {
		testToleration corev1.Toleration
		expectedOutput bool
	}{
		{
			testToleration: corev1.Toleration{
				Key: "node.kubernetes.io/test1",
			},
			expectedOutput: true,
		},
		{
			testToleration: corev1.Toleration{
				Key: "this.is.a.test/test1",
			},
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, IsTolerationDefault(tc.testToleration))
	}
}
