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

// StringInSlice checks whether a given value is present in a slice.
//
// It accepts three arguments: a slice of values, the value to search for,
// and a boolean indicating whether the comparison should ignore case.
// The function returns true if the value exists in the slice, otherwise false.
// When ignoreCase is true, both the target value and each element are trimmed
// of leading/trailing whitespace and compared in a case-insensitive manner.
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

// SubSlice checks whether all elements of the first slice appear in the second.
//
// It returns true if every string in the first argument is found within
// the second slice, otherwise it returns false. The function uses StringInSlice
// internally to perform membership tests.
func SubSlice(s, sub []string) bool {
	for _, v := range sub {
		if !StringInSlice(s, v, false) {
			return false
		}
	}
	return true
}

// HasAtLeastOneCommonElement reports whether two slices share any element.
//
// It returns true if there exists at least one string that appears in both input slices, otherwise false. The function iterates over the first slice and checks each value for membership in the second slice.
func HasAtLeastOneCommonElement(s1, s2 []string) bool {
	for _, v := range s2 {
		if StringInSlice(s1, v, false) {
			return true
		}
	}
	return false
}

// RemoveEmptyStrings returns a new slice with all empty strings removed.
//
// It takes a slice of strings, iterates over each element,
// and appends only non‑empty strings to the result.
// The returned slice contains the original order of non‑empty elements.
func RemoveEmptyStrings(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

// PointerToString returns a string representation of the value pointed to by p.
// It handles nil pointers by returning the literal string "nil". For non‑nil
// pointers, it converts the underlying value to its default string form,
// typically using fmt.Sprint.
//
// The function accepts a single pointer argument of any type and yields a
// string. If the pointer is nil, the result is "nil"; otherwise, the
// dereferenced value is formatted as a string. This helper is useful for
// logging or tracing Kubernetes resources where pointer fields need to be
// displayed in human‑readable form.
func PointerToString[T any](p *T) string {
	if p == nil {
		return "nil"
	} else {
		return fmt.Sprint(*p)
	}
}
