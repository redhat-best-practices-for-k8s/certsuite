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

package accesscontrol

import (
	"errors"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/stretchr/testify/assert"
)

func TestGetNbOfProcessesInPidNamespace(t *testing.T) {
	testCases := []struct {
		testPID               int
		expectedErr           error
		expectedResult        int
		expectedExecFuncCalls int

		// func results
		execOutStr string
		execErrStr string
		execErr    error
	}{
		{ // Test Case #1 - Failure because the exec only returns one value 8080
			testPID:        1337,
			expectedErr:    errors.New(`cmd: " lsns -p 1337 -t pid -n " returned an invalid value 8080`),
			expectedResult: 0,

			execOutStr:            "8080",
			execErrStr:            "",
			execErr:               nil,
			expectedExecFuncCalls: 1,
		},
		{ // Test Case #2 - Pass
			testPID:        1337,
			expectedErr:    nil,
			expectedResult: 8082,

			execOutStr:            "8080 8081 8082",
			execErrStr:            "",
			execErr:               nil,
			expectedExecFuncCalls: 1,
		},
		{ // Test Case #3 - Failure - Error performing exec
			testPID:        1337,
			expectedErr:    errors.New("can not execute command: \" lsns -p 1337 -t pid -n \", err:this is an error"),
			expectedResult: 0,

			execOutStr:            "8080 8081 8082",
			execErrStr:            "",
			execErr:               errors.New("this is an error"),
			expectedExecFuncCalls: 1,
		},
		{ // Test Case #4 - Failure - StdErr after running exec
			testPID:        1337,
			expectedErr:    errors.New("cmd: \" lsns -p 1337 -t pid -n \" returned this is an error"),
			expectedResult: 0,

			execOutStr:            "8080 8081 8082",
			execErrStr:            "this is an error",
			execErr:               nil,
			expectedExecFuncCalls: 1,
		},
	}

	for _, tc := range testCases {
		// Setup a mock version of the clientsHolder so we can "run" commands
		ch := &clientsholder.CommandMock{
			ExecCommandContainerFunc: func(context clientsholder.Context, s string) (string, string, error) {
				return tc.execOutStr, tc.execErrStr, tc.execErr
			},
		}

		result, err := getNbOfProcessesInPidNamespace(clientsholder.NewContext("testNamespace", "testPod", testhelper.ContainerName), tc.testPID, ch)

		// assertions
		assert.Equal(t, tc.expectedResult, result)
		assert.Equal(t, tc.expectedErr, err)
		assert.Equal(t, tc.expectedExecFuncCalls, len(ch.ExecCommandContainerCalls()))
	}
}
