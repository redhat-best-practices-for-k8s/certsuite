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

	"github.com/hashicorp/go-version"
)

// StringInSlice checks a slice for a given string.
func StringInSlice(s []string, str string, contains bool) bool {
	for _, v := range s {
		if !contains {
			if strings.TrimSpace(v) == str {
				return true
			}
		} else {
			if strings.Contains(strings.TrimSpace(v), str) {
				return true
			}
		}
	}
	return false
}

// RemoveDuplicates returns a new slice with unique element in input slice
func RemoveDuplicates(str []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range str {
		if !keys[entry] {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// CompareVersion compare between versions
func CompareVersion(ver1, ver2 string) bool {
	ourKubeVersion, _ := version.NewVersion(ver1)
	kubeVersion := strings.ReplaceAll(ver2, " ", "")[2:]
	if strings.Contains(kubeVersion, "<") {
		kubever := strings.Split(kubeVersion, "<")
		minVersion, _ := version.NewVersion(kubever[0])
		maxVersion, _ := version.NewVersion(kubever[1])
		if ourKubeVersion.GreaterThanOrEqual(minVersion) && ourKubeVersion.LessThan(maxVersion) {
			return true
		}
	} else {
		kubever := strings.Split(kubeVersion, "-")
		minVersion, _ := version.NewVersion(kubever[0])
		if ourKubeVersion.GreaterThanOrEqual(minVersion) {
			return true
		}
	}
	return false
}
