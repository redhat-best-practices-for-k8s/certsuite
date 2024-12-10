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

package stringhelper

import (
	"fmt"
	"strings"
)

// StringInSlice checks a slice for a given string.
func StringInSlice[T ~string](s []T, str T, contains bool) bool {
	for _, v := range s {
		if !contains {
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

// SubSlice checks if a slice's elements all exist within a slice
func SubSlice(s, sub []string) bool {
	for _, v := range sub {
		if !StringInSlice(s, v, false) {
			return false
		}
	}
	return true
}

// checks that at least one element is common to both slices
func HasAtLeastOneCommonElement(s1, s2 []string) bool {
	for _, v := range s2 {
		if StringInSlice(s1, v, false) {
			return true
		}
	}
	return false
}

func RemoveEmptyStrings(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

// PointerToString returns the default string representation of the value pointer by p, mainly
// used in log traces to print k8s resources' pointer fields.
// If p is a nil pointer, no matter the type, it will return the string "nil".
//
// # Example 1
//
//	var b* bool
//	PointerToString(b) -> returns "nil"
//
// # Example 2
//
//	b := true
//	bTrue := &b
//	PointerToString(bTrue) -> returns "true"
//
// # Example 3
//
//	var num *int
//	PointerToString(num) -> returns "nil"
//
// # Example 4
//
//	num := 1984
//	num1984 := &num
//	PointerToString(num1984) -> returns "1984"
func PointerToString[T any](p *T) string {
	if p == nil {
		return "nil"
	} else {
		return fmt.Sprint(*p)
	}
}
