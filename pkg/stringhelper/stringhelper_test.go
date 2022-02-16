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

package stringhelper

import (
	"testing"

	"gotest.tools/assert"
)

func TestStringInSlice(t *testing.T) {
	testCases := []struct {
		testSlice  []string
		testString string
		expected   bool
	}{
		{
			testSlice: []string{
				"apples",
				"bananas",
				"oranges",
			},
			testString: "apples",
			expected:   true,
		},
		{
			testSlice: []string{
				"apples",
				"bananas",
				"oranges",
			},
			testString: "tacos",
			expected:   false,
		},
		{
			testSlice: []string{
				"intree: Y",
				"intree: N",
				"outoftree: Y",
			},
			testString: "intree:",
			expected:   false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, StringInSlice(tc.testSlice, tc.testString))
	}
}

func TestSubStringInSlice(t *testing.T) {
	testCases := []struct {
		testSlice  []string
		testString string
		expected   bool
	}{
		{
			testSlice: []string{
				"intree: Y",
				"intree: N",
				"outoftree: Y",
			},
			testString: "intree:",
			expected:   true,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, SubStringInSlice(tc.testSlice, tc.testString))
	}
}

//nolint:funlen
func Test_hasIntersection(t *testing.T) {
	type args struct {
		s   []string
		str []string
	}
	tests := []struct {
		name             string
		args             args
		wantFullMatch    bool
		wantPartialMatch bool
	}{
		{
			args: args{[]string{
				"apples",
				"bananas",
				"oranges",
			},
				[]string{"apples"}},
			wantFullMatch:    true,
			wantPartialMatch: false,
		},
		{
			args: args{[]string{
				"apples",
				"bananas",
				"oranges",
			},
				[]string{"tacos"}},
			wantFullMatch:    false,
			wantPartialMatch: false,
		},
		{
			args: args{[]string{
				"intree: Y",
				"intree: N",
				"outoftree: Y",
			},
				[]string{"intree:"}},
			wantFullMatch:    false,
			wantPartialMatch: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFullMatch, gotPartialMatch := hasIntersection(tt.args.s, tt.args.str)
			if gotFullMatch != tt.wantFullMatch {
				t.Errorf("hasIntersection() gotFullMatch = %v, want %v", gotFullMatch, tt.wantFullMatch)
			}
			if gotPartialMatch != tt.wantPartialMatch {
				t.Errorf("hasIntersection() gotPartialMatch = %v, want %v", gotPartialMatch, tt.wantPartialMatch)
			}
		})
	}
}
