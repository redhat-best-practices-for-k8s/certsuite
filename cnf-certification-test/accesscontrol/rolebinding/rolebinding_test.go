// Copyright (C) 2021 Red Hat, Inc.
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

package rolebinding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleOutOfNamespace(t *testing.T) {
	testCases := []struct {
		testRoleNS             string
		testPodNS              string
		testRoleName           string
		testServiceAccountName string
		expectedOutOfNS        bool
	}{
		{ // Test Case #1 - Pod and Role are in the same namespace.
			testRoleNS:             "ns1",
			testPodNS:              "ns1",
			testRoleName:           "sa1",
			testServiceAccountName: "sa1",

			expectedOutOfNS: false,
		},
		{ // Test Case #2 - Pod and Role are in different namespaces.
			testRoleNS:             "ns1",
			testPodNS:              "ns2",
			testRoleName:           "sa1",
			testServiceAccountName: "sa1",

			expectedOutOfNS: true,
		},
		{ // Test Case #3 - Pod, Role names don't match and are in different namespaces.
			testRoleNS:             "ns1",
			testPodNS:              "ns2",
			testRoleName:           "sa1",
			testServiceAccountName: "sa2",

			expectedOutOfNS: false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutOfNS, roleOutOfNamespace(tc.testRoleNS, tc.testPodNS, tc.testRoleName, tc.testServiceAccountName))
	}
}
