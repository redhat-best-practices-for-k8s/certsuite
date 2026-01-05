// Copyright (C) 2021-2024 Red Hat, Inc.
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

// Package diagnostic provides a test suite which gathers OpenShift cluster information.
package diagnostics

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestGetVersionOcClient(t *testing.T) {
	assert.Equal(t, "n/a, (not using oc or kubectl client)", GetVersionOcClient())
}

func TestGetHWJsonOutput(t *testing.T) {
	type TestingJSON struct {
		Testing string `json:"testing"`
	}

	testCases := []struct {
		execStdout string
		execStderr string
		execErr    error

		expectedErr    error
		expectedResult TestingJSON
	}{
		{
			execStdout: `{"testing":"hello world"}`,
			execStderr: "",
			execErr:    nil,

			expectedErr: nil,
			expectedResult: TestingJSON{
				Testing: "hello world",
			},
		},
		{
			execStdout: `{"testing":"hello world"}`,
			execStderr: "this is an error",
			execErr:    nil,

			expectedErr:    errors.New("command does not matter failed with error err: <nil>, stderr: this is an error"),
			expectedResult: TestingJSON{},
		},
		{
			execStdout: `{"testing":"hello world"}`,
			execStderr: "this is an error",
			execErr:    errors.New("this is an error2"),

			expectedErr:    errors.New("command does not matter failed with error err: <nil>, stderr: this is an error"),
			expectedResult: TestingJSON{},
		},
	}

	for _, tc := range testCases {
		result, err := getHWJsonOutput(&corev1.Pod{
			Spec: corev1.PodSpec{
				// Note: We do not actually care about the podname
				// for this test, but the function uses it to build the
				// context .
				Containers: []corev1.Container{
					{
						Name: testhelper.PodName,
					},
				},
			},
		}, &clientsholder.MockCommand{
			ExecFunc: func(context clientsholder.Context, s string) (string, string, error) {
				return tc.execStdout, tc.execStderr, nil
			},
		}, "does not matter")

		tj := TestingJSON{}
		tjBytes, _ := json.Marshal(result)
		_ = json.Unmarshal(tjBytes, &tj)

		assert.Equal(t, tc.expectedErr, err)
		assert.Equal(t, tc.expectedResult, tj)
	}
}

func TestGetHWTextOutput(t *testing.T) {
	testCases := []struct {
		execStdout string
		execStderr string
		execErr    error

		expectedErr    error
		expectedResult []string
	}{
		{
			execStdout: "hello\nworld",
			execStderr: "",
			execErr:    nil,

			expectedErr:    nil,
			expectedResult: []string{"hello", "world"},
		},
		{
			execStdout: `{"testing":"hello world"}`,
			execStderr: "this is an error",
			execErr:    nil,

			expectedErr:    errors.New("command lspci failed with error err: <nil>, stderr: this is an error"),
			expectedResult: nil,
		},
		{
			execStdout: `{"testing":"hello world"}`,
			execStderr: "this is an error",
			execErr:    errors.New("this is an error2"),

			expectedErr:    errors.New("command lspci failed with error err: <nil>, stderr: this is an error"),
			expectedResult: nil,
		},
	}

	for _, tc := range testCases {
		result, err := getHWTextOutput(&corev1.Pod{
			Spec: corev1.PodSpec{
				// Note: We do not actually care about the podname
				// for this test, but the function uses it to build the
				// context .
				Containers: []corev1.Container{
					{
						Name: testhelper.PodName,
					},
				},
			},
		}, &clientsholder.MockCommand{
			ExecFunc: func(context clientsholder.Context, s string) (string, string, error) {
				return tc.execStdout, tc.execStderr, nil
			},
		}, lspciCommand)

		assert.Equal(t, tc.expectedErr, err)
		assert.Equal(t, tc.expectedResult, result)
	}
}
