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

// BaseImageInfo holds the command execution context used to inspect a container image.
//
// It contains a client holder that provides methods to run commands inside a
// container and a context representing the OpenShift cluster configuration.
// These fields are required by the TestContainerIsRedHatRelease method,
// which determines whether the running container is based on a Red‑Hat release.
type BaseImageInfo struct {
	ClientHolder clientsholder.Command
	OCPContext   clientsholder.Context
}

// NewBaseImageTester creates a BaseImageInfo instance using the provided command executor and context.
//
// It accepts a Command interface used to run shell commands and a Context that supplies
// configuration or environment information for the test. The function executes the necessary
// checks against the base image, populates a BaseImageInfo struct with the results,
// and returns a pointer to that struct. If an error occurs during command execution,
// it is wrapped inside the returned BaseImageInfo structure for further handling.
func NewBaseImageTester(client clientsholder.Command, ctx clientsholder.Context) *BaseImageInfo {
	return &BaseImageInfo{
		ClientHolder: client,
		OCPContext:   ctx,
	}
}

// TestContainerIsRedHatRelease determines if the container image is a Red Hat Enterprise Linux release.
//
// It runs a command inside the container to inspect its OS information, logs details for debugging,
// and checks whether the detected distribution matches known RHEL identifiers.
// The function returns true if the image is identified as RHEL, otherwise false along with any error encountered.
func (b *BaseImageInfo) TestContainerIsRedHatRelease() (bool, error) {
	output, err := b.runCommand(`if [ -e /etc/redhat-release ]; then cat /etc/redhat-release; else echo \"Unknown Base Image\"; fi`)
	log.Info("Output from /etc/redhat-release: %q", output)
	if err != nil {
		return false, err
	}
	return IsRHEL(output), nil
}

// IsRHEL reports whether the given operating system name matches a Red Hat Enterprise Linux release.
//
// It checks the input string against predefined regular expressions for non‑Red Hat based systems and
// standard RHEL versions, returning true if either pattern finds a match.
// The function returns a single boolean value indicating presence of an RHEL release.
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

// runCommand executes a shell command inside the container associated with the BaseImageInfo instance.
//
// It takes a single string argument representing the command to run and returns the command's
// standard output as a string along with an error if the execution fails or if there is
// any issue capturing the output. The function internally uses ExecCommandContainer
// to perform the command execution within the container environment. If the command
// exits with a non‑zero status, runCommand returns an error created via errors.New.
// The returned string contains whatever was written to stdout by the executed command.
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
