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

package stringhelper

import (
	"strings"
)

// StringInSlice checks a slice for a given string.
func StringInSlice(s []string, str string) bool {
	fullMatch, _ := hasIntersection(s, []string{str})
	return fullMatch
}

// StringInSlicePartialMatch checks a slice for a given substring.
func SubStringInSlice(s []string, str string) bool {
	fullMatch, partialMatch := hasIntersection(s, []string{str})
	return fullMatch || partialMatch
}

func hasIntersection(s, str []string) (fullMatch, partialMatch bool) {
	for _, aString := range str {
		for _, aStringInSlice := range s {
			if strings.TrimSpace(aStringInSlice) == aString {
				fullMatch = true
				partialMatch = false
				return fullMatch, partialMatch
			}
			if strings.Contains(strings.TrimSpace(aStringInSlice), aString) {
				fullMatch = false
				partialMatch = true
				return fullMatch, partialMatch
			}
		}
	}
	return false, false
}
