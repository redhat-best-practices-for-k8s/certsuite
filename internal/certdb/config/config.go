// Copyright (C) 2020-2022 Red Hat, Inc.
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
package config

import (
	"time"

	"github.com/sirupsen/logrus"
)

const (
	registryRedhatIo        = "registry.redhat.io"
	registryAccessRedhatCom = "registry.access.redhat.com"
)

// HardcodedRegistryMapping is map holding hardcoded entries in pyxis
var HardcodedRegistryMapping = map[string]string{registryRedhatIo: registryAccessRedhatCom}

// determines certification status for Redhat images
func IsRegistryRedhatOnlyImages(registry, publishedDate string) bool {
	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	date, err := time.Parse("2006-01-02T15:04:05.999999-07:00", publishedDate)
	if err != nil {
		logrus.Errorf("could not parse image published date, container is not certified, err=%s", err)
		return false
	}

	return (registry == registryRedhatIo || registry == registryAccessRedhatCom) && date.After(oneYearAgo)
}
