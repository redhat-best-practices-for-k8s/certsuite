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
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
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

	// The following test cases should cover all of the possible scenarios that determine pass/fail
	// of the AutomountServiceToken test.
	testCases := []struct {
		podSATokenStatus *bool // pod token status is
		saTokenStatus    *bool // returned from the mock API
		expectedResult   bool
		getSAErr         error
		expectedAPICalls int
	}{
		{podSATokenStatus: &trueVar, saTokenStatus: &trueVar, getSAErr: nil, expectedAPICalls: 0, expectedResult: false},                             // FAIL because pod SA token is set to true
		{podSATokenStatus: &trueVar, saTokenStatus: &falseVar, getSAErr: nil, expectedAPICalls: 0, expectedResult: false},                            // FAIL because pod SA token is set to true
		{podSATokenStatus: &falseVar, saTokenStatus: &trueVar, getSAErr: nil, expectedAPICalls: 1, expectedResult: true},                             // PASS because pod SA set to false
		{podSATokenStatus: &falseVar, saTokenStatus: nil, getSAErr: nil, expectedAPICalls: 1, expectedResult: true},                                  // PASS because pod SA set to false
		{podSATokenStatus: &falseVar, saTokenStatus: &trueVar, getSAErr: errors.New("this is an error"), expectedAPICalls: 1, expectedResult: false}, // FAIL because failure to gather SA status from API
		{podSATokenStatus: nil, saTokenStatus: &trueVar, getSAErr: nil, expectedAPICalls: 1, expectedResult: false},                                  // FAIL because pod SA token is set to true and the pod SA is nil
		{podSATokenStatus: nil, saTokenStatus: nil, getSAErr: nil, expectedAPICalls: 1, expectedResult: false},                                       // FAIL because pod SA token is nil and the pod SA is nil
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
		assert.Equal(t, tc.expectedAPICalls, len(mockFuncs.AutomountServiceAccountSetOnSACalls()))
	}
}

//nolint:funlen
func TestTestPodRoleBindings(t *testing.T) {
	testCases := []struct {
		// return values
		getRoleBindingFuncsRet []string
		getRoleBindingFuncsErr error

		// expected results
		expectedAPICalls int
		expectedResult   bool
	}{
		{ // Test Case #1 - Pass with no rolebindingds found
			getRoleBindingFuncsRet: []string{}, // No rolebindings found in other namespaces
			getRoleBindingFuncsErr: nil,

			expectedAPICalls: 1,
			expectedResult:   true,
		},
		{ // Test Case #2 - Fail with rolebindings found
			getRoleBindingFuncsRet: []string{"SA1"}, // rolebindings found in other namespaces
			getRoleBindingFuncsErr: nil,

			expectedAPICalls: 1,
			expectedResult:   false,
		},
		{ // Test Case #3 - Fail with API call failure
			getRoleBindingFuncsRet: []string{"SA1"},
			getRoleBindingFuncsErr: errors.New("this is an error"),

			expectedAPICalls: 1,
			expectedResult:   false,
		},
	}

	for _, tc := range testCases {
		// Assume each test run is going to pass unless GinkgoFail is called.
		sharedResult := true

		// Generate the TestEnvironment
		generateEnv := func() *provider.TestEnvironment {
			return &provider.TestEnvironment{
				Pods: []*v1.Pod{
					{
						Spec: v1.PodSpec{
							NodeName:           "worker01",
							ServiceAccountName: "SA1",
						},
						ObjectMeta: v1meta.ObjectMeta{
							Name:      "testPod",
							Namespace: "testNamespace",
						},
					},
				},
				GinkgoFuncs: &tnf.GinkgoFuncsMock{
					GinkgoAbortSuiteFunc: func(message string, callerSkip ...int) {},
					GinkgoByFunc:         func(text string, callback ...func()) {},
					GinkgoFailFunc: func(message string, callerSkip ...int) {
						sharedResult = false
					},
					GinkgoSkipFunc: func(message string, callerSkip ...int) {},
				},
			}
		}

		// Test the function with mocked internal functions.
		mockFuncs := &rbac.RoleBindingFuncsMock{
			GetRoleBindingsFunc: func(podNamespace, serviceAccountName string) ([]string, error) {
				return tc.getRoleBindingFuncsRet, tc.getRoleBindingFuncsErr
			},
		}
		TestPodRoleBindings(generateEnv(), mockFuncs)
		assert.Equal(t, tc.expectedResult, sharedResult)
		assert.Equal(t, tc.expectedAPICalls, len(mockFuncs.GetRoleBindingsCalls()))
	}
}
