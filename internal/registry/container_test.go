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
package registry

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIsCertified(t *testing.T) {
	path, _ := os.Getwd()
	log.Info(path)
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	loadContainersCatalog(path + "/../")
	ans := IsCertified("registry.connect.redhat.com", "bitnami/nodejs", "11.14.0-rhel-7-r5-5", "")
	assert.Equal(t, true, ans)

	ans = IsCertified("registry.connect.redhat.com", "nearform/nearform-s2i-nodejs10", "10.1.0", "")
	assert.Equal(t, true, ans)
}
