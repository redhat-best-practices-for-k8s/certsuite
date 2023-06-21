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
package offlinecheck

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const containersDBJSON = `{
	"sha256:d2f388e163a5126f7112757f0475c8c4e036fe00c76ef8a7fd50848fafdcb96b":{
	   "_id":"",
	   "architecture":"amd64",
	   "certified":false,
	   "image_id":"sha256:d2f388e163a5126f7112757f0475c8c4e036fe00c76ef8a7fd50848fafdcb96b",
	   "docker_image_id":"",
	   "repositories":[
		  {
			 "registry":"quay.io",
			 "repository":"fujitsu/fujitsu-enterprise-postgres-13-exporter",
			 "tags":[
				{
				   "name":"ubi8-13-1.0-amd64"
				}
			 ]
		  },
		  {
			 "registry":"registry.rhc4tp.openshift.com",
			 "repository":"ospid-613f8666333748670463a91b/partner-build-service",
			 "tags":[
				{
				   "name":"ubi8-13-1.0-amd64"
				}
			 ]
		  }
	   ]
	},
	"sha256:fa8f2136aed9daf4c5a805068a87dd274016b8dddae36bc0b02e18b391690493":{
	   "_id":"",
	   "architecture":"amd64",
	   "certified":false,
	   "image_id":"sha256:fa8f2136aed9daf4c5a805068a87dd274016b8dddae36bc0b02e18b391690493",
	   "docker_image_id":"",
	   "repositories":[
		  {
			 "registry":"registry.connect.redhat.com",
			 "repository":"rocketchat/rocketchat",
			 "tags":[
				{
				   "name":"0.56.0-1"
				},
				{
				   "name":"latest"
				}
			 ]
		  }
	   ]
	}
 }`

func loadContainersDB() error {
	bytes, err := io.ReadAll(strings.NewReader(containersDBJSON))
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &containerdb)
	if err != nil {
		return err
	}

	return nil
}

func TestIsCertified(t *testing.T) {
	validator := OfflineValidator{}

	assert.NoError(t, loadContainersDB())

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
			tag:                         "ubi8-13-1.0-amd64",
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
			digest:                      "sha256:fa8f2136aed9daf4c5a805068a87dd274016b8dddae36bc0b02e18b391690493",
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
		isCertified := validator.IsContainerCertified(tc.registry, tc.repository, tc.tag, tc.digest)
		assert.Equal(t, tc.expectedCertificationStatus, isCertified)
	}
}
