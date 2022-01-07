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

package version

import (
	"encoding/json"
	"os"
	"path"
)

var (
	defaultVersionFile = path.Join("..", "version.json")
)

// Version refers to the `test-network-function` version tag.
type Version struct {
	// Tag is the Git tag for the version.
	Tag string `json:"tag" yaml:"tag"`
}

// GetVersion extracts the test-network-function version.
func GetVersion() (*Version, error) {
	contents, err := os.ReadFile(defaultVersionFile)
	if err != nil {
		return nil, err
	}
	version := &Version{}
	err = json.Unmarshal(contents, version)
	return version, err
}
