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
package offlinecheck

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIsOperatorCertified(t *testing.T) {
	t.Skip() // TODO: Offline certification tests that need the DB should be moved to the OCT repo
	validator := OfflineValidator{}
	name := "zoperator.v0.3.6"
	ocpversion := "4.6"
	channel := "alpha"
	path, _ := os.Getwd()
	log.Info(path)
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	_ = loadOperatorsCatalog(path + "/../../")
	assert.True(t, validator.IsOperatorCertified(name, ocpversion, channel))
	name = "falcon-alpha"
	assert.False(t, validator.IsOperatorCertified(name, ocpversion, channel))

	assert.True(t, validator.IsOperatorCertified("artifactory-ha-operator.v1.2.0", "4.9", "alpha"))
}
