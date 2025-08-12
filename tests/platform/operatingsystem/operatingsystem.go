// Copyright (C) 2022-2024 Red Hat, Inc.
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

package operatingsystem

import (
	_ "embed"
	"strings"
)

const (
	NotFoundStr = "version-not-found"
)

//go:embed files/rhcos_version_map
var rhcosVersionMap string

// GetRHCOSMappedVersions parses a string of RHCOS version mappings and returns them as a map.
//
// It expects the input to contain lines in the form "key: value". The function splits the
// input into lines, trims whitespace around keys and values, and builds a map from key to value.
// If parsing succeeds it returns the map and nil error; otherwise it returns an empty map and an error.
func GetRHCOSMappedVersions(rhcosVersionMap string) (map[string]string, error) {
	capturedInfo := make(map[string]string)

	// Example: Translate `Red Hat Enterprise Linux CoreOS 410.84.202205031645-0 (Ootpa)` into a RHCOS version number
	// and long-form counterpart

	/// Example lines from the captured file
	// 4.9.21 / 49.84.202202081504-0
	// 4.9.25 / 49.84.202203112054-0
	// 4.10.14 / 410.84.202205031645-0

	versions := strings.Split(rhcosVersionMap, "\n")
	for _, v := range versions {
		// Skip any empty lines
		if strings.TrimSpace(v) == "" {
			continue
		}

		// Split on the / and capture the line into the map
		splitVersion := strings.Split(v, "/")
		capturedInfo[strings.TrimSpace(splitVersion[0])] = strings.TrimSpace(splitVersion[1])
	}

	return capturedInfo, nil
}

// GetShortVersionFromLong extracts the short OS version from a full version string.
//
// It accepts a long-form operating system version string, looks up a
// corresponding short form in the rhcosVersionMap, and returns the
// short version string along with an error if the input is empty or not found.
func GetShortVersionFromLong(longVersion string) (string, error) {
	capturedVersions, err := GetRHCOSMappedVersions(rhcosVersionMap)
	if err != nil {
		return "", err
	}

	// search through all available rhcos versions for a match
	for s, l := range capturedVersions {
		if l == longVersion {
			return s, nil
		}
	}

	// return "version-not-found" if the short version cannot be found
	return NotFoundStr, nil
}
