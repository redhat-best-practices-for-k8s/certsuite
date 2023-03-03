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

package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStatefulsetToString(t *testing.T) {
	ss := StatefulSet{
		StatefulSet: &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: "testNS",
			},
		},
	}

	assert.Equal(t, "statefulset: test1 ns: testNS", ss.ToString())
}

func TestIsStatefulSetReady(t *testing.T) {
	generateSS := func(specReplicas *int32, statusReadyReplicas, statusCurrentReplicas, statusUpdatedReplicas int32) *StatefulSet {
		return &StatefulSet{
			StatefulSet: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas: specReplicas,
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas:   statusReadyReplicas,
					UpdatedReplicas: statusUpdatedReplicas,
					CurrentReplicas: statusCurrentReplicas,
				},
			},
		}
	}

	toInt32Ptr := func(num int32) *int32 {
		return &num
	}

	testCases := []struct {
		testSpecReplicas          *int32
		testReadyStatusReplicas   int32
		testUpdatedStatusReplicas int32
		testCurrentStatusReplicas int32
		expectedOutput            bool
	}{
		{
			testSpecReplicas:          toInt32Ptr(10),
			testReadyStatusReplicas:   10,
			testUpdatedStatusReplicas: 10,
			testCurrentStatusReplicas: 10,
			expectedOutput:            true,
		},
		{
			testSpecReplicas:          nil,
			testReadyStatusReplicas:   10,
			testUpdatedStatusReplicas: 10,
			testCurrentStatusReplicas: 10,
			expectedOutput:            false,
		},
		{
			testSpecReplicas:          toInt32Ptr(0),
			testReadyStatusReplicas:   0,
			testUpdatedStatusReplicas: 0,
			testCurrentStatusReplicas: 0,
			expectedOutput:            true,
		},
	}

	for _, tc := range testCases {
		testSS := generateSS(tc.testSpecReplicas, tc.testReadyStatusReplicas, tc.testCurrentStatusReplicas, tc.testUpdatedStatusReplicas)
		assert.Equal(t, tc.expectedOutput, testSS.IsStatefulSetReady())
	}
}
