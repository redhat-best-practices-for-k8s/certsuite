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

package version

var (
	// GitCommit is the latest commit in the current git branch
	GitCommit string
	// GitRelease is the list of tags (if any) applied to the latest commit
	// in the current branch
	GitRelease string
	// GitPreviousRelease is the last release at the date of the latest commit
	// in the current branch
	GitPreviousRelease string
)

// GetGitVersion returns the git display version: the latest previously released
// build in case this build is not released. Otherwise display the build version
func GetGitVersion() string {
	// gitDisplayRelease is a string used to hold the text to display
	// the version on screen and in the claim file
	var gitDisplayRelease string
	if GitRelease == "" {
		gitDisplayRelease = "Unreleased build post " + GitPreviousRelease
	} else {
		gitDisplayRelease = GitRelease
	}

	return gitDisplayRelease + " ( " + GitCommit + " )"
}
