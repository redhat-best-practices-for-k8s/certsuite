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

package platform

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform/nodetainted"
	clientsholder "github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	v1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//nolint:funlen
func TestTestTainted(t *testing.T) {
	generateEnv := func(acceptedModule string) *provider.TestEnvironment {
		return &provider.TestEnvironment{
			DebugPods: map[string]*v1.Pod{
				"debug-pod-01": {
					Spec: v1.PodSpec{
						NodeName: "worker01",
						Containers: []v1.Container{
							{},
						},
					},
					ObjectMeta: v1meta.ObjectMeta{
						Name:      "testPod",
						Namespace: "testNamespace",
					},
				},
			},
			Config: configuration.TestConfiguration{
				AcceptedKernelTaints: []configuration.AcceptedKernelTaintsInfo{
					{
						Module: acceptedModule,
					},
				},
			},
		}
	}

	testCases := []struct {
		// Spoofed responses
		taintInfoFuncRet       string
		taintInfoFuncErr       error
		modulesFromNodeFuncRet []string
		outOfTreeFuncRet       []string
		modulesInTreeFuncRet   bool

		// expected calls to funcs
		expectedGetKernelTaintInfoCalls int
		expectedGetModulesFromNodeCalls int
		expectedModuleInTreeCalls       int
		expectedGetOutOfTreeModules     int

		// environment vars
		acceptedModule string

		// results
		expectedResult bool
	}{
		{ // Test Case #1 - Tainted node, no 'module was loaded' however it failed because of another taint
			taintInfoFuncRet: "512", // kernel issued warning
			taintInfoFuncErr: nil,

			modulesFromNodeFuncRet: []string{}, // no modules pulled because `kernel issued warning` isn't a taint about modules
			outOfTreeFuncRet:       []string{},
			modulesInTreeFuncRet:   false, // unused
			acceptedModule:         "",

			expectedGetKernelTaintInfoCalls: 1,
			expectedGetModulesFromNodeCalls: 1,
			expectedModuleInTreeCalls:       0,
			expectedGetOutOfTreeModules:     1,

			expectedResult: false, // Fail because of a tainted node
		},
		{ // Test Case #2 - Un-tainted node, but fail to parse uint
			taintInfoFuncRet: "",
			taintInfoFuncErr: nil,

			modulesFromNodeFuncRet: []string{},
			outOfTreeFuncRet:       []string{},
			modulesInTreeFuncRet:   false, // unused
			acceptedModule:         "",

			expectedGetKernelTaintInfoCalls: 1,
			expectedGetModulesFromNodeCalls: 0,
			expectedModuleInTreeCalls:       0,
			expectedGetOutOfTreeModules:     0,

			expectedResult: false, // Fail because failure to parse uint
		},
		{ // Test Case #3 - Tainted node, but an accepted module
			taintInfoFuncRet: "4096",
			taintInfoFuncErr: nil,

			modulesFromNodeFuncRet: []string{"vboxsf"},
			outOfTreeFuncRet:       []string{"vboxsf"},
			modulesInTreeFuncRet:   false,
			acceptedModule:         "vboxsf", // acceptable

			expectedGetKernelTaintInfoCalls: 1,
			expectedGetModulesFromNodeCalls: 1,
			expectedModuleInTreeCalls:       1,
			expectedGetOutOfTreeModules:     1,

			expectedResult: true, // Pass because accepted module
		},
		{ // Test Case #4 - Tainted node, not accepted
			taintInfoFuncRet: "4096",
			taintInfoFuncErr: nil,

			modulesFromNodeFuncRet: []string{"vboxsf"},
			outOfTreeFuncRet:       []string{"vboxsf"},
			modulesInTreeFuncRet:   false,
			acceptedModule:         "", // no modules accepted

			expectedGetKernelTaintInfoCalls: 1,
			expectedGetModulesFromNodeCalls: 1,
			expectedModuleInTreeCalls:       1,
			expectedGetOutOfTreeModules:     1,

			expectedResult: false, // Fail because not-accepted
		},
		{ // Test Case #5 - Tainted node with multiple taints different reasons including `module was loaded`
			// Kernel is Tainted for following reasons:
			// * Proprietary module was loaded (#0)
			// * Kernel issued warning (#9)
			// * Externally-built ('out-of-tree') module was loaded  (#12)
			taintInfoFuncRet: "4609",
			taintInfoFuncErr: nil,

			modulesFromNodeFuncRet: []string{"vboxsf"},
			outOfTreeFuncRet:       []string{"vboxsf"},
			modulesInTreeFuncRet:   false,
			acceptedModule:         "vboxsf", // no modules accepted

			expectedGetKernelTaintInfoCalls: 1,
			expectedGetModulesFromNodeCalls: 1,
			expectedModuleInTreeCalls:       1,
			expectedGetOutOfTreeModules:     1,

			expectedResult: false, // Fail because module is accepted but there are other reasons for taints
		},
		{ // Test Case #6 - Failure to gather taint info from node
			taintInfoFuncRet: "",
			taintInfoFuncErr: errors.New("this is an error"),

			modulesFromNodeFuncRet: []string{},
			outOfTreeFuncRet:       []string{},
			modulesInTreeFuncRet:   false, // unused
			acceptedModule:         "",

			expectedGetKernelTaintInfoCalls: 1,
			expectedGetModulesFromNodeCalls: 0,
			expectedModuleInTreeCalls:       0,
			expectedGetOutOfTreeModules:     0,

			expectedResult: false, // Fail because failed to query node's taint info
		},
		{ // Test Case #7 - Un-tainted node
			taintInfoFuncRet: "0",
			taintInfoFuncErr: nil,

			modulesFromNodeFuncRet: []string{},
			outOfTreeFuncRet:       []string{},
			modulesInTreeFuncRet:   false, // unused
			acceptedModule:         "",

			expectedGetKernelTaintInfoCalls: 1,
			expectedGetModulesFromNodeCalls: 0,
			expectedModuleInTreeCalls:       0,
			expectedGetOutOfTreeModules:     0,

			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		sharedResult := false
		mockFuncs := &nodetainted.TaintedFuncsMock{
			GetKernelTaintInfoFunc: func(ctx clientsholder.Context) (string, error) {
				return tc.taintInfoFuncRet, tc.taintInfoFuncErr
			},
			GetModulesFromNodeFunc: func(ctx clientsholder.Context) []string {
				return tc.modulesFromNodeFuncRet
			},
			GetOutOfTreeModulesFunc: func(modules []string, ctx clientsholder.Context) []string {
				return tc.outOfTreeFuncRet
			},
			ModuleInTreeFunc: func(moduleName string, ctx clientsholder.Context) bool {
				return tc.modulesInTreeFuncRet
			},
			SetTestingResultFunc: func(result bool) {
				sharedResult = result
			},
		}

		// Run the tests
		testTainted(generateEnv(tc.acceptedModule), mockFuncs)

		// Assertions
		assert.Equal(t, tc.expectedResult, sharedResult)
		assert.Equal(t, tc.expectedGetKernelTaintInfoCalls, len(mockFuncs.GetKernelTaintInfoCalls()))
	}
}
