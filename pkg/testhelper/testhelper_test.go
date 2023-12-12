// Copyright (C) 2020-2023 Red Hat, Inc.
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

package testhelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResultToString(t *testing.T) {
	testCases := []struct {
		input          int
		expectedResult string
	}{
		{input: SUCCESS, expectedResult: "SUCCESS"},
		{input: FAILURE, expectedResult: "FAILURE"},
		{input: ERROR, expectedResult: "ERROR"},
		{input: 1337, expectedResult: ""},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedResult, ResultToString(tc.input))
	}
}
