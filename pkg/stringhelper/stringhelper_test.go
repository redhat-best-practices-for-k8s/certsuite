// Copyright (C) 2020-2026 Red Hat, Inc.
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

package stringhelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

type otherString string

func TestStringInSlice(t *testing.T) {
	testCases := []struct {
		testSlice       []string
		testString      string
		containsFeature bool
		expected        bool
	}{
		{
			testSlice: []string{
				"apples",
				"bananas",
				"oranges",
			},
			testString:      "apples",
			containsFeature: false,
			expected:        true,
		},
		{
			testSlice: []string{
				"apples",
				"bananas",
				"oranges",
			},
			testString:      "tacos",
			containsFeature: false,
			expected:        false,
		},
		{
			testSlice: []string{
				"intree: Y",
				"intree: N",
				"outoftree: Y",
			},
			testString:      "intree:",
			containsFeature: true, // Note: Turn 'on' the contains check
			expected:        true,
		},
		{
			testSlice: []string{
				"intree: Y",
				"intree: N",
				"outoftree: Y",
			},
			testString:      "intree:",
			containsFeature: false, // Note: Turn 'off' the contains check
			expected:        false,
		},
		{
			testSlice: []string{
				"oneapple",
			},
			testString:      "apple",
			containsFeature: false, // Note: Turn 'off' the contains check
			expected:        false,
		},
		{
			testSlice: []string{
				"apples",
			},
			testString:      "twoapples",
			containsFeature: false, // Note: Turn 'off' the contains check
			expected:        false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, StringInSlice(tc.testSlice, tc.testString, tc.containsFeature))
	}
}

func TestStringInSlice_other(t *testing.T) {
	testCases := []struct {
		testSlice       []otherString
		testString      otherString
		containsFeature bool
		expected        bool
	}{
		{
			testSlice: []otherString{
				"apples",
				"bananas",
				"oranges",
			},
			testString:      "apples",
			containsFeature: false,
			expected:        true,
		},
		{
			testSlice: []otherString{
				"apples",
				"bananas",
				"oranges",
			},
			testString:      "tacos",
			containsFeature: false,
			expected:        false,
		},
		{
			testSlice: []otherString{
				"intree: Y",
				"intree: N",
				"outoftree: Y",
			},
			testString:      "intree:",
			containsFeature: true, // Note: Turn 'on' the contains check
			expected:        true,
		},
		{
			testSlice: []otherString{
				"intree: Y",
				"intree: N",
				"outoftree: Y",
			},
			testString:      "intree:",
			containsFeature: false, // Note: Turn 'off' the contains check
			expected:        false,
		},
		{
			testSlice: []otherString{
				"intreeintreeintree",
			},
			testString:      "intree",
			containsFeature: false, // Note: Turn 'off' the contains check
			expected:        false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, StringInSlice(tc.testSlice, tc.testString, tc.containsFeature))
	}
}

func TestRemoveEmptyStrings(t *testing.T) {
	testCases := []struct {
		testSlice     []string
		expectedSlice []string
	}{
		{
			testSlice:     []string{"one", "two", "three", "", ""},
			expectedSlice: []string{"one", "two", "three"},
		},
		{ // returns a nil slice if the contents of the incoming slice are empty
			testSlice:     []string{"", ""},
			expectedSlice: nil,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedSlice, RemoveEmptyStrings(tc.testSlice))
	}
}

func TestSubSlice(t *testing.T) {
	testCases := []struct {
		testSliceA     []string
		testSliceB     []string
		expectedOutput bool
	}{
		{ // Test #1 - SliceB exists in SliceA
			testSliceA:     []string{"one", "two", "three"},
			testSliceB:     []string{"one", "two"},
			expectedOutput: true,
		},
		{ // Test #2 - SliceB does not exist in SliceA
			testSliceA:     []string{"one", "two", "three"},
			testSliceB:     []string{"four", "five"},
			expectedOutput: false,
		},
		{ // Test #3 - Same slices, return true
			testSliceA:     []string{"one", "two", "three"},
			testSliceB:     []string{"one", "two", "three"},
			expectedOutput: true,
		},
		{ // Test Case #4 - Empty SliceA
			testSliceA:     []string{},
			testSliceB:     []string{"one", "two", "three"},
			expectedOutput: false,
		},
		{ // Test #5 - SliceB's elements exist out of order in SliceA
			testSliceA:     []string{"one", "two", "three"},
			testSliceB:     []string{"two", "one"},
			expectedOutput: true,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, SubSlice(tc.testSliceA, tc.testSliceB))
	}
}

func TestPointerToString(t *testing.T) {
	const wantNil = "nil"

	var want string

	// pointer to bool
	var boolPointer *bool
	want = wantNil
	if got := PointerToString(boolPointer); got != want {
		t.Errorf("PointerToString() = %v, want %v", got, want)
	}

	boolPointer = ptr.To(true)
	want = "true"
	if got := PointerToString(boolPointer); got != want {
		t.Errorf("PointerToString() = %v, want %v", got, want)
	}

	// pointer to number
	var numPointer *int64
	want = wantNil
	if got := PointerToString(numPointer); got != want {
		t.Errorf("PointerToString() = %v, want %v", got, want)
	}

	numPointer = ptr.To(int64(1984))
	want = "1984"
	if got := PointerToString(numPointer); got != want {
		t.Errorf("PointerToString() = %v, want %v", got, want)
	}

	// pointer to string
	var stringPointer *string
	want = "nil"
	if got := PointerToString(stringPointer); got != want {
		t.Errorf("PointerToString() = %v, want %v", got, want)
	}

	stringPointer = ptr.To("hello, world!")
	want = "hello, world!"
	if got := PointerToString(stringPointer); got != want {
		t.Errorf("PointerToString() = %v, want %v", got, want)
	}
}

func TestHasAtLeastOneCommonElement(t *testing.T) {
	testCases := []struct {
		slice1   []string
		slice2   []string
		expected bool
	}{
		{
			slice1:   []string{"one", "two", "three"},
			slice2:   []string{"one", "two"},
			expected: true,
		},
		{
			slice1:   []string{"one", "two", "three"},
			slice2:   []string{"four", "five"},
			expected: false,
		},
		{
			slice1:   []string{"one", "two", "three"},
			slice2:   []string{"one", "two", "three"},
			expected: true,
		},
		{
			slice1:   []string{},
			slice2:   []string{"one", "two", "three"},
			expected: false,
		},
		{
			slice1:   []string{"one", "two", "three"},
			slice2:   []string{"two", "one"},
			expected: true,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, HasAtLeastOneCommonElement(tc.slice1, tc.slice2))
	}
}
