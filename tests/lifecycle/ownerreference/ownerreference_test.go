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

package ownerreference_test

import (
	"strings"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle/ownerreference"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRunTest(t *testing.T) {
	testCases := []struct {
		podKind        string
		expectedResult int
	}{
		{
			podKind:        "StatefulSet",
			expectedResult: testhelper.SUCCESS,
		},
		{
			podKind:        "ReplicaSet",
			expectedResult: testhelper.SUCCESS,
		},
		{
			podKind:        "NotARealKind",
			expectedResult: testhelper.FAILURE,
		},
	}

	for _, tc := range testCases {
		testPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "testpod",
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind: tc.podKind,
					},
				},
			},
			Spec: corev1.PodSpec{},
		}

		ownerRef := ownerreference.NewOwnerReference(testPod)
		assert.NotNil(t, ownerRef)
		var logArchive strings.Builder
		log.SetupLogger(&logArchive, "INFO")
		ownerRef.RunTest(log.GetLogger())
		assert.Equal(t, tc.expectedResult, ownerRef.GetResults())
	}
}
