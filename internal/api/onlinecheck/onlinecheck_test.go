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
package onlinecheck

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
)

func TestGetRequest(t *testing.T) {
	checker := NewOnlineValidator()
	_, err := checker.GetRequest("http://non-existingurl.com")
	assert.NotNil(t, err)
	_, err = checker.GetRequest(redhatCatalogPingURL)
	assert.Nil(t, err)
}

func TestCreateContainerCatalogQueryURL(t *testing.T) {
	testCases := []struct {
		testContainerImageID configuration.ContainerImageIdentifier
		expectedResult       string
	}{
		{
			testContainerImageID: configuration.ContainerImageIdentifier{
				Name:       "name1",
				Repository: "repo1",
				Tag:        "tag1",
				Digest:     "digest1",
			},
			expectedResult: apiCatalogByRepositoriesBaseEndPoint + "/repo1/name1/images?filter=architecture==amd64;image_id==digest1",
		},
		{
			testContainerImageID: configuration.ContainerImageIdentifier{
				Name:       "name1",
				Repository: "repo1",
				Tag:        "tag1",
				// Digest:     "digest1",
			},
			expectedResult: apiCatalogByRepositoriesBaseEndPoint + "/repo1/name1/images?filter=architecture==amd64;repositories.repository==repo1/name1;repositories.tags.name==tag1",
		},
		{
			testContainerImageID: configuration.ContainerImageIdentifier{
				Name:       "name1",
				Repository: "repo1",
				// Tag:        "tag1",
				// Digest:     "digest1",
			},
			expectedResult: apiCatalogByRepositoriesBaseEndPoint + "/repo1/name1/images?filter=architecture==amd64;repositories.repository==repo1/name1;repositories.tags.name==latest",
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedResult, CreateContainerCatalogQueryURL(tc.testContainerImageID))
	}
}
