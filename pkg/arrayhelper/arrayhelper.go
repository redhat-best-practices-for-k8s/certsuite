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

// ArgListToMap Converts key=value strings into a map
//
// It receives an array of strings, each representing a kernel argument or
// configuration pair. For every entry it removes surrounding quotes, splits on
// the first equals sign, and stores the key with its corresponding in a new
// map. The resulting map is returned for further processing.
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

// FilterArray Filters elements of a slice based on a predicate
//
// It iterates over each string in the input slice, applies the provided
// function to decide if an element should be kept, and collects those that
// satisfy the condition into a new slice which is then returned.
func FilterArray(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

// Unique Eliminates duplicate strings from a slice
//
// The function receives a slice of strings and returns a new slice containing
// each distinct element exactly once. It builds a map to track seen values,
// then collects the unique keys into a result slice. The order of elements is
// not preserved.
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
