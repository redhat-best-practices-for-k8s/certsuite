// Copyright (C) 2022-2024 Red Hat, Inc.
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

package operatingsystem

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRHCOSMappedVersionsFromFile(t *testing.T) {
	testCases := []struct {
		expectedOutput map[string]string
		filename       string
		expectedErr    error
	}{
		{
			filename: "testdata/versionMapTest.txt",
			expectedOutput: map[string]string{
				"4.9.21":  "49.84.202202081504-0",
				"4.9.25":  "49.84.202203112054-0",
				"4.10.14": "410.84.202205031645-0",
			},
			expectedErr: nil,
		},
		{
			filename: "testdata/wrongfile.txt",
			expectedOutput: map[string]string{
				"4.9.21":  "49.84.202202081504-0",
				"4.9.25":  "49.84.202203112054-0",
				"4.10.14": "410.84.202205031645-0",
			},
			expectedErr: errors.New("this is an error"),
		},
	}

	for _, tc := range testCases {
		// read the relative path file
		// var content embed.FS
		// file, err := content.ReadFile(tc.filename)
		file, err := os.ReadFile(tc.filename)

		if tc.expectedErr != nil {
			assert.Error(t, err)
		} else {
			result, err := GetRHCOSMappedVersions(string(file))
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedOutput, result)
		}
	}
}

func TestGetShortVersionFromLong(t *testing.T) {
	testCases := []struct {
		testLongVersion      string
		expectedShortVersion string
		expectedErr          error
	}{
		{ // Test Case #1 - valid version found
			testLongVersion:      "49.84.202202081504-0",
			expectedShortVersion: "4.9.21",
			expectedErr:          nil,
		},
		{ // Test Case #2 - invalid long version, not found in file.
			testLongVersion:      "1.3.1337",
			expectedShortVersion: "version-not-found",
			expectedErr:          nil,
		},
	}

	for _, tc := range testCases {
		result, err := GetShortVersionFromLong(tc.testLongVersion)
		if tc.expectedErr != nil {
			assert.Error(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, tc.expectedShortVersion, result)
	}
}
