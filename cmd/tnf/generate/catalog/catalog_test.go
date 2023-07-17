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

package catalog

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/arrayhelper"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
)

func TestNewCommand(t *testing.T) {
	assert.NotNil(t, NewCommand())
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
	assert.Nil(t, runGenerateMarkdownCmd(nil, nil))
}

func TestUnique(t *testing.T) {
	testCases := []struct {
		testSlice     []string
		expectedSlice []string
	}{
		{
			testSlice:     []string{"one", "two", "three"},
			expectedSlice: []string{"one", "two", "three"},
		},
		{
			testSlice:     []string{"one", "two", "three", "three"},
			expectedSlice: []string{"one", "two", "three"},
		},
		{
			testSlice:     []string{},
			expectedSlice: []string{},
		},
	}

	for _, tc := range testCases {
		sort.Strings(tc.expectedSlice)
		results := arrayhelper.Unique(tc.testSlice)
		sort.Strings(results)
		assert.True(t, reflect.DeepEqual(tc.expectedSlice, results))
	}
}

func TestGetSuitesFromIdentifiers(t *testing.T) {
	testCases := []struct {
		testKeys       []claim.Identifier
		expectedSuites []string
	}{
		{
			testKeys: []claim.Identifier{
				{
					Id:    "helloworld/test",
					Suite: "helloworld",
				},
				{
					Id:    "helloworld2/test2",
					Suite: "helloworld2",
				},
			},
			expectedSuites: []string{"helloworld", "helloworld2"},
		},
	}

	for _, tc := range testCases {
		sort.Strings(tc.expectedSuites)
		results := GetSuitesFromIdentifiers(tc.testKeys)
		sort.Strings(results)
		assert.True(t, reflect.DeepEqual(tc.expectedSuites, results))
	}
}
