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

package isredhat

import (
	"errors"
	"regexp"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
)

const (
	// NotRedHatBasedRegex is the expected output for a container that is not based on Red Hat technologies.
	NotRedHatBasedRegex = `(?m)Unknown Base Image`
	// VersionRegex is regular expression expected for a container based on Red Hat technologies.
	VersionRegex = `(?m)Red Hat Enterprise Linux( Server)? release (\d+\.\d+)`
)

// BaseImageInfo provides utilities for inspecting a container’s base image
//
// The struct holds a command executor and context, enabling it to run commands
// inside a container. It offers methods such as TestContainerIsRedHatRelease,
// which checks the presence of /etc/redhat-release to determine if the image is
// RHEL-based, returning a boolean and error. The helper runCommand executes
// arbitrary shell commands via the client holder, handling errors and capturing
// output.
type BaseImageInfo struct {
	ClientHolder clientsholder.Command
	OCPContext   clientsholder.Context
}

// NewBaseImageTester Creates a new instance of the base image tester
//
// The function accepts a client holder and a contextual object representing a
// Kubernetes pod or container. It constructs and returns a pointer to a struct
// that stores these inputs for subsequent checks on the container's base image.
// No additional processing occurs during construction.
func NewBaseImageTester(client clientsholder.Command, ctx clientsholder.Context) *BaseImageInfo {
	return &BaseImageInfo{
		ClientHolder: client,
		OCPContext:   ctx,
	}
}

// BaseImageInfo.TestContainerIsRedHatRelease Checks if the container image is a Red Hat release
//
// The method runs a shell command inside the container to read
// /etc/redhat-release or report an unknown base image, logs the output, and
// then uses IsRHEL to determine whether the image matches known Red Hat
// patterns. It returns true when the image is confirmed as a Red Hat release,
// otherwise false, along with any execution error that occurs.
func (b *BaseImageInfo) TestContainerIsRedHatRelease() (bool, error) {
	output, err := b.runCommand(`if [ -e /etc/redhat-release ]; then cat /etc/redhat-release; else echo \"Unknown Base Image\"; fi`)
	log.Info("Output from /etc/redhat-release: %q", output)
	if err != nil {
		return false, err
	}
	return IsRHEL(output), nil
}

// IsRHEL determines whether the provided string signifies a Red Hat based release
//
// The function examines the supplied text for patterns that indicate a
// non‑Red Hat base image and immediately returns false if such patterns are
// found. If no negative matches occur, it logs the content of
// /etc/redhat-release and checks against a regular expression describing
// official Red Hat releases, returning true when a match is detected.
func IsRHEL(output string) bool {
	// If the 'Unknown Base Image' string appears, return false.
	notRedHatRegex := regexp.MustCompile(NotRedHatBasedRegex)
	matchNotRedhat := notRedHatRegex.FindAllString(output, -1)
	if len(matchNotRedhat) > 0 {
		return false
	}

	// /etc/redhat-release exists. check if it matches the regex for an official build.
	log.Info("redhat-release was found to be: %s", output)
	redHatVersionRegex := regexp.MustCompile(VersionRegex)
	matchVersion := redHatVersionRegex.FindAllString(output, -1)
	return len(matchVersion) > 0
}

// BaseImageInfo.runCommand Executes a shell command inside a container
//
// The method runs the supplied command in the container using the client
// holder, capturing both standard output and error streams. If execution fails
// or an error string is returned, it logs the issue and propagates an error to
// the caller. On success, it returns the command's output as a string.
func (b *BaseImageInfo) runCommand(cmd string) (string, error) {
	output, outerr, err := b.ClientHolder.ExecCommandContainer(b.OCPContext, cmd)
	if err != nil {
		log.Error("can not execute command on container, err: %v", err)
		return "", err
	}
	if outerr != "" {
		log.Error("Error when running baseimage command, err: %v", outerr)
		return "", errors.New(outerr)
	}
	return output, nil
}
