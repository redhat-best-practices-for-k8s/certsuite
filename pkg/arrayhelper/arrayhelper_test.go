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

package arrayhelper

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterArray(t *testing.T) {
	stringFilter := func(incomingVar string) bool {
		return strings.Contains(incomingVar, "test")
	}

	testCases := []struct {
		arrayToFilter []string
		expectedArray []string
	}{
		{
			arrayToFilter: []string{"test1", "test2"},
			expectedArray: []string{"test1", "test2"},
		},
		{
			arrayToFilter: []string{"apples", "oranges"},
			expectedArray: []string{},
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedArray, FilterArray(tc.arrayToFilter, stringFilter))
	}
}

func TestArgListToMap(t *testing.T) {
	testCases := []struct {
		argList     []string
		expectedMap map[string]string
	}{
		{
			argList: []string{"key1=value1", `key2="value2"`},
			expectedMap: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			argList: []string{"key1=value1", "key2=value2"},
			expectedMap: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			argList:     []string{},
			expectedMap: map[string]string{},
		},
		{
			argList: []string{"key1=value1", "key2=value2", "key3"},
			expectedMap: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "",
			},
		},
	}

	for _, tc := range testCases {
		assert.True(t, reflect.DeepEqual(tc.expectedMap, ArgListToMap(tc.argList)))
	}
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
		results := Unique(tc.testSlice)
		sort.Strings(results)
		assert.True(t, reflect.DeepEqual(tc.expectedSlice, results))
	}
}
