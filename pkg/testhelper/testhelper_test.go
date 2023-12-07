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
	"strings"
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

func TestSkipIfEmptyFuncs(t *testing.T) {
	testCases := []struct {
		map1, slice1           interface{}
		skippedAny, skippedAll bool
	}{
		{ // Test Case #1 - Skip because objects is empty, no panic because []string type
			slice1:     []string{},
			map1:       make(map[string]string),
			skippedAny: true,
			skippedAll: true,
		},
		{ // Test Case #2 - Skip because objects is empty, no panic because map type
			slice1:     make(map[string]string),
			map1:       make(map[string]string),
			skippedAny: true,
			skippedAll: true,
		},
		{ // Test Case #3 - No skip because objects is populated, no panic because []string type
			slice1:     []string{"test"},
			map1:       make(map[string]string),
			skippedAny: true,
			skippedAll: false,
		},
		{ // Test Case #4 - Multiple objects
			slice1:     []string{"test1", "test2"},
			map1:       make(map[string]string),
			skippedAny: true,
			skippedAll: false,
		},
		{ // Test Case #5 - Nil objects panic
			slice1:     nil,
			map1:       nil,
			skippedAny: true,
			skippedAll: true,
		},
	}

	for _, tc := range testCases {
		result := false

		if tc.slice1 == nil || tc.map1 == nil {
			assert.Panics(t, func() {
				SkipIfEmptyAll(func(s string, i ...int) {}, NewSkipObject(tc.slice1, "tc.slice1"), NewSkipObject(tc.map1, "tc.map1"))
			})
			assert.Panics(t, func() {
				SkipIfEmptyAny(func(s string, i ...int) {}, NewSkipObject(tc.slice1, "tc.slice1"), NewSkipObject(tc.map1, "tc.map1"))
			})
		} else {
			SkipIfEmptyAll(func(s string, i ...int) {
				if strings.Contains(s, "Test skipped") {
					result = true
				} else {
					result = false
				}
			}, NewSkipObject(tc.slice1, "tc.slice1"), NewSkipObject(tc.map1, "tc.map1"))
			assert.Equal(t, tc.skippedAll, result)

			SkipIfEmptyAny(func(s string, i ...int) {
				if strings.Contains(s, "Test skipped") {
					result = true
				} else {
					result = false
				}
			}, NewSkipObject(tc.slice1, "tc.slice1"), NewSkipObject(tc.map1, "tc.map1"))
			assert.Equal(t, tc.skippedAny, result)
		}
	}
}
