// Copyright (C) 2020-2022 Red Hat, Inc.
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

package isredhat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRHEL(t *testing.T) {
	testCases := []struct {
		testOutput string
		expected   bool
	}{
		{
			testOutput: "Red Hat Enterprise Linux release 8.5 (Ootpa)",
			expected:   true,
		},
		{
			testOutput: "Unknown Base Image",
			expected:   false,
		},
		{
			testOutput: "CentOS",
			expected:   false,
		},
		{
			testOutput: "",
			expected:   false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, IsRHEL(tc.testOutput))
	}
}
