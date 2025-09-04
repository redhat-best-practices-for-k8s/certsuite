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

package stringhelper

import (
	"fmt"
	"strings"
)

// StringInSlice Checks if a value exists in a string slice
//
// The function iterates over the provided slice, trimming whitespace from each
// element before comparison. If containsCheck is false it tests for exact
// equality; otherwise it checks whether the element contains the target
// substring. It returns true as soon as a match is found, otherwise false.
func StringInSlice[T ~string](s []T, str T, containsCheck bool) bool {
	for _, v := range s {
		if !containsCheck {
			if strings.TrimSpace(string(v)) == string(str) {
				return true
			}
		} else {
			if strings.Contains(strings.TrimSpace(string(v)), string(str)) {
				return true
			}
		}
	}
	return false
}

// SubSlice verifies all elements of one slice exist in another
//
// The function receives two string slices: the main slice and a candidate
// sub-slice. It iterates over each element of the candidate, checking for an
// exact match within the main slice using StringInSlice. If any element is
// missing, it returns false; otherwise it returns true after all checks pass.
func SubSlice(s, sub []string) bool {
	for _, v := range sub {
		if !StringInSlice(s, v, false) {
			return false
		}
	}
	return true
}

// HasAtLeastOneCommonElement verifies whether two string collections contain a shared value
//
// The routine iterates over the second slice and checks each element against
// the first using a helper that compares trimmed strings for equality. If any
// match is found, it immediately returns true; otherwise it completes the loop
// and returns false.
func HasAtLeastOneCommonElement(s1, s2 []string) bool {
	for _, v := range s2 {
		if StringInSlice(s1, v, false) {
			return true
		}
	}
	return false
}

// RemoveEmptyStrings Filters out empty entries from a slice
//
// This function iterates over an input list of strings, selecting only those
// that are not empty. It builds a new slice containing the non-empty values and
// returns it. The original slice is left unchanged.
func RemoveEmptyStrings(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

// PointerToString converts a pointer to its string representation
//
// When the argument is nil, it returns "nil"; otherwise it dereferences the
// pointer and formats the value using standard printing rules. The function
// works for any type thanks to generics, making it useful in log traces or
// debugging output.
func PointerToString[T any](p *T) string {
	if p == nil {
		return "nil"
	} else {
		return fmt.Sprint(*p)
	}
}
