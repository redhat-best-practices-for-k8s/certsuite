// Copyright (C) 2022-2026 Red Hat, Inc.
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

// GetShortVersionFromLong returns the first matching OCP version for a given RHCOS long version.
// Deprecated: Use GetAllShortVersionsFromLong instead, as the same RHCOS version may map to multiple OCP versions.
func GetShortVersionFromLong(longVersion string) (string, error) {
	versions, err := GetAllShortVersionsFromLong(longVersion)
	if err != nil {
		return "", err
	}
	if len(versions) == 0 {
		return NotFoundStr, nil
	}
	return versions[0], nil
}

// GetAllShortVersionsFromLong returns all matching OCP versions for a given RHCOS long version.
// The same RHCOS build can be used by multiple OCP versions (e.g., 9.6.20251125-1 maps to both 4.19.20 and 4.20.6).
// This function returns all matching versions to allow proper compatibility checking.
func GetAllShortVersionsFromLong(longVersion string) ([]string, error) {
	capturedVersions, err := GetRHCOSMappedVersions(rhcosVersionMap)
	if err != nil {
		return nil, err
	}

	var matchingVersions []string
	// search through all available rhcos versions for matches
	for s, l := range capturedVersions {
		if l == longVersion {
			matchingVersions = append(matchingVersions, s)
		}
	}

	return matchingVersions, nil
}
