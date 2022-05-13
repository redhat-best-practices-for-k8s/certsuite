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

package catalog

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	assert.NotNil(t, NewCommand())
}

func TestCmdJoin(t *testing.T) {
	testCases := []struct {
		testElems []string
		testSep   string
		expected  string
	}{
		{
			testElems: []string{"this", "is", "a", "test"},
			testSep:   ".",
			expected:  "`this`.`is`.`a`.`test`",
		},
		{
			testElems: []string{},
			testSep:   ".",
			expected:  "",
		},
		{
			testElems: []string{"this"},
			testSep:   ".",
			expected:  "`this`",
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, cmdJoin(tc.testElems, tc.testSep))
	}
}

func TestEmitTextFromFile(t *testing.T) {
	testCases := []struct {
		filename    string
		expectedErr error
	}{
		{
			filename:    "testdata/testFile1.txt",
			expectedErr: nil,
		},
		{
			filename:    "testdata/unknown.txt",
			expectedErr: errors.New("this is an error"),
		},
	}

	for _, tc := range testCases {
		if tc.expectedErr != nil {
			assert.NotNil(t, emitTextFromFile(tc.filename))
		} else {
			assert.Nil(t, emitTextFromFile(tc.filename))
		}
	}
}

func TestRunGenerateMarkdownCmd(t *testing.T) {
	assert.NotNil(t, runGenerateMarkdownCmd(nil, nil))
}
