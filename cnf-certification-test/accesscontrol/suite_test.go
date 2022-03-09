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

package accesscontrol

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/rbac"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//nolint:funlen
func TestTestAutomountServiceToken(t *testing.T) {
	generateEnv := func(tokenStatus *bool) *provider.TestEnvironment {
		return &provider.TestEnvironment{
			Pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						NodeName:                     "worker01",
						AutomountServiceAccountToken: tokenStatus,
						ServiceAccountName:           "SA1",
					},
					ObjectMeta: v1meta.ObjectMeta{
						Name:      "testPod",
						Namespace: "testNamespace",
					},
				},
			},
		}
	}

	falseVar := false
	trueVar := true

	testCases := []struct {
		podSATokenStatus *bool // pod token status is
		saTokenStatus    *bool // returned from the mock API
		expectedResult   bool
		getSAErr         error
	}{
		{podSATokenStatus: &trueVar, saTokenStatus: &trueVar, getSAErr: nil, expectedResult: false},                             // FAIL because pod SA token is set to true
		{podSATokenStatus: &trueVar, saTokenStatus: &falseVar, getSAErr: nil, expectedResult: false},                            // FAIL because pod SA token is set to true
		{podSATokenStatus: &falseVar, saTokenStatus: &trueVar, getSAErr: nil, expectedResult: true},                             // PASS because pod SA set to false
		{podSATokenStatus: &falseVar, saTokenStatus: nil, getSAErr: nil, expectedResult: true},                                  // PASS because pod SA set to false
		{podSATokenStatus: &falseVar, saTokenStatus: &trueVar, getSAErr: errors.New("this is an error"), expectedResult: false}, // FAIL because failure to gather SA status from API
		{podSATokenStatus: nil, saTokenStatus: &trueVar, getSAErr: nil, expectedResult: false},                                  // FAIL because pod SA token is set to true and the pod SA is nil
		{podSATokenStatus: nil, saTokenStatus: nil, getSAErr: nil, expectedResult: false},                                       // FAIL because pod SA token is nil and the pod SA is nil
	}

	for _, tc := range testCases {
		sharedResult := false

		// Test the function with mocked internal functions.
		mockFuncs := &rbac.AutomountTokenFuncsMock{
			AutomountServiceAccountSetOnSAFunc: func(serviceAccountName, podNamespace string) (*bool, error) {
				return tc.saTokenStatus, tc.getSAErr
			},
			SetTestingResultFunc: func(result bool) {
				sharedResult = result
			},
		}
		TestAutomountServiceToken(generateEnv(tc.podSATokenStatus), mockFuncs)
		assert.Equal(t, tc.expectedResult, sharedResult)
	}
}
