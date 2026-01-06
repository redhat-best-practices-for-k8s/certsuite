// Copyright (C) 2020-2026 Red Hat, Inc.
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

type BaseImageInfo struct {
	ClientHolder clientsholder.Command
	OCPContext   clientsholder.Context
}

func NewBaseImageTester(client clientsholder.Command, ctx clientsholder.Context) *BaseImageInfo {
	return &BaseImageInfo{
		ClientHolder: client,
		OCPContext:   ctx,
	}
}

func (b *BaseImageInfo) TestContainerIsRedHatRelease() (bool, error) {
	output, err := b.runCommand(`if [ -e /etc/redhat-release ]; then cat /etc/redhat-release; else echo \"Unknown Base Image\"; fi`)
	log.Info("Output from /etc/redhat-release: %q", output)
	if err != nil {
		return false, err
	}
	return IsRHEL(output), nil
}

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
