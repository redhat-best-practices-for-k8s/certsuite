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

//nolint:funlen
func TestIsCertified(t *testing.T) {
	checker := OfflineChecker{}
	path, _ := os.Getwd()
	log.Info(path)
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	loadContainersCatalog(path + "/../../")

	// Note: This test cases might have to change periodically due to images coming/going from the offline database.
	testCases := []struct {
		registry                    string
		repository                  string
		tag                         string
		digest                      string
		expectedCertificationStatus bool
	}{
		// Check status based on the tag.
		{
			registry:                    "quay.io",
			repository:                  "fujitsu/fujitsu-enterprise-postgres-13-exporter",
			tag:                         "ubi8-13-1.3-amd64",
			digest:                      "",
			expectedCertificationStatus: true,
		},
		{
			registry:                    "registry.connect.redhat.com",
			repository:                  "rocketchat/rocketchat",
			tag:                         "0.56.0-1",
			digest:                      "",
			expectedCertificationStatus: true,
		},
		// When no tag provided, we'll assume it's the 'latest' one.
		{
			registry:                    "registry.connect.redhat.com",
			repository:                  "rocketchat/rocketchat",
			tag:                         "",
			digest:                      "",
			expectedCertificationStatus: true,
		},
		// Check certification status based on digest only.
		{
			registry:                    "registry.connect.redhat.com",
			repository:                  "rocketchat/rocketchat",
			tag:                         "",
			digest:                      "sha256:b1d5b80d4c119c4316d9fa38a6a21383f30b07b67d8efc762530283a8d070070",
			expectedCertificationStatus: true,
		},
		// Not existing image.
		{
			registry:                    "registry.connect.redhat.com",
			repository:                  "iDoNotExist",
			tag:                         "",
			digest:                      "",
			expectedCertificationStatus: false,
		},
		// Not existing tag.
		{
			registry:                    "registry.connect.redhat.com",
			repository:                  "rocketchat/rocketchat",
			tag:                         "fakeNotCertifiedTag",
			digest:                      "",
			expectedCertificationStatus: false,
		},
		// Empty struct
		{
			registry:                    "",
			repository:                  "",
			tag:                         "",
			digest:                      "",
			expectedCertificationStatus: false,
		},
	}

	for _, tc := range testCases {
		isCertified := checker.IsContainerCertified(tc.registry, tc.repository, tc.tag, tc.digest)
		assert.Equal(t, isCertified, tc.expectedCertificationStatus)
	}
}
