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

// GetRHCOSMappedVersions Parses a formatted string of RHCOS versions into a mapping
//
// The function receives a multiline string where each line contains a short
// RHCOS version, a slash, and its long-form counterpart. It splits the input by
// newline, trims whitespace, ignores empty lines, then separates each pair on
// the slash to build a map from short to long versions. The resulting map is
// returned along with any error .
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

// GetShortVersionFromLong Retrieves a concise RHCOS version string from its full identifier
//
// The function looks up the provided long-form RHCOS identifier in a preloaded
// mapping of short-to-long versions. If a match is found, it returns the
// corresponding short version; otherwise, it returns a sentinel value
// indicating that the version was not located.
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
