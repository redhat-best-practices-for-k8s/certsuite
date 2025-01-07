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

package manageability

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerPortNameFormatCheck(t *testing.T) {
	testCases := []struct {
		portName       string
		expectedOutput bool
	}{
		{
			portName:       "http",
			expectedOutput: true,
		},
		{
			portName:       "tcp-probe",
			expectedOutput: true,
		},
		{
			portName:       "grpc-web-app1",
			expectedOutput: true,
		},
		{
			portName:       "sftp",
			expectedOutput: false,
		},
		{
			portName:       "sctp-endpoint",
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		res := containerPortNameFormatCheck(tc.portName)
		assert.Equal(t, tc.expectedOutput, res)
	}
}
