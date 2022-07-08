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

func TestIsCertified(t *testing.T) {
	checker := OfflineChecker{}
	path, _ := os.Getwd()
	log.Info(path)
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	loadContainersCatalog(path + "/../../")

	// Note: This test case might have to change periodically due to images coming/going from the offline database.
	ans := checker.IsContainerCertified("quay.io", "fujitsu/fujitsu-enterprise-postgres-13-exporter", "ubi8-13-1.3-amd64", "")
	assert.Equal(t, ans, true)
}
