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

package arrayhelper

import (
	"strings"
)

// ArgListToMap converts a slice of strings formatted as key=value into a map.
//
// It iterates over each string in the input slice, splits it on the first '='
// character, and assigns the resulting key and value to a new map. The function
// returns this map containing all parsed key/value pairs. If an element does not
// contain '=', it is ignored. This utility is useful for parsing command-line
// arguments or configuration options into a dictionary.
func ArgListToMap(lst []string) map[string]string {
	retval := make(map[string]string)
	for _, arg := range lst {
		arg = strings.ReplaceAll(arg, `"`, ``)
		splitArgs := strings.Split(arg, "=")
		if len(splitArgs) == 1 {
			retval[splitArgs[0]] = ""
		} else {
			retval[splitArgs[0]] = splitArgs[1]
		}
	}
	return retval
}

// FilterArray filters a slice of strings using the provided predicate.
//
// It iterates over each element in the input slice and appends those
// for which the predicate function returns true to a new slice.
// The resulting slice is returned, leaving the original unchanged.
func FilterArray(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

// Unique removes duplicate strings from a slice and returns a new slice containing only unique values.
//
// It takes a slice of strings, iterates over each element, and builds a new slice that contains
// each string exactly once while preserving the original order of first occurrences.
// The function does not modify the input slice.
func Unique(slice []string) []string {
	// create a map with all the values as key
	uniqMap := make(map[string]struct{})
	for _, v := range slice {
		uniqMap[v] = struct{}{}
	}

	// turn the map keys into a slice
	uniqSlice := make([]string, 0, len(uniqMap))
	for v := range uniqMap {
		uniqSlice = append(uniqSlice, v)
	}
	return uniqSlice
}
