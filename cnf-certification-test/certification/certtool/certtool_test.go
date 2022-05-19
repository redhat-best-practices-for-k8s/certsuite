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

package certtool

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/internal/api"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
)

func TestGetContainerCertificationRequestFunction(t *testing.T) {
	// Mock out the CertAPIClient
	defer func() {
		CertAPIClient = api.CertAPIClient{}
	}()
	CertAPIClient = &api.CertAPIClientFuncsMock{
		GetContainerCatalogEntryFunc: func(id configuration.ContainerImageIdentifier) (*api.ContainerCatalogEntry, error) {
			return &api.ContainerCatalogEntry{
				ID: id.Name,
			}, nil
		},
	}

	testCases := []struct {
		imageID configuration.ContainerImageIdentifier
	}{
		{
			imageID: configuration.ContainerImageIdentifier{
				Name:       "testID",
				Repository: "testRepo",
				Tag:        "testTag",
				Digest:     "testDigest",
			},
		},
	}

	// Run the test
	for _, tc := range testCases {
		result := GetContainerCertificationRequestFunction(tc.imageID)
		id, err := result()
		assert.Nil(t, err)
		myResult := id.(*api.ContainerCatalogEntry)
		assert.Equal(t, tc.imageID.Name, myResult.ID)
	}
}

func TestGetOperatorCertificationRequestFunction(t *testing.T) {
	// Mock out the CertAPIClient
	defer func() {
		CertAPIClient = api.CertAPIClient{}
	}()

	testCases := []struct {
		opCertFuncRet    bool
		opCertFuncRetErr error
	}{
		{
			opCertFuncRet:    true,
			opCertFuncRetErr: nil,
		},
		{
			opCertFuncRet:    false,
			opCertFuncRetErr: nil,
		},
		{
			opCertFuncRet:    true,
			opCertFuncRetErr: errors.New("this is an error"),
		},
	}

	for _, tc := range testCases {
		// Mock out the return values for the API func
		CertAPIClient = &api.CertAPIClientFuncsMock{
			IsOperatorCertifiedFunc: func(org, packageName, version string) (bool, error) {
				return tc.opCertFuncRet, tc.opCertFuncRetErr
			},
		}

		resultFunc := GetOperatorCertificationRequestFunction("", "", "")
		certifiedResult, err := resultFunc()
		assert.Equal(t, tc.opCertFuncRetErr, err)
		testResult := certifiedResult.(bool)
		assert.Equal(t, tc.opCertFuncRet, testResult)
	}
}

func TestWaitForCertificationRequestToSuccess(t *testing.T) {
	testCases := []struct {
		testFunc   func() (interface{}, error)
		expectedID string
		timeout    time.Duration
	}{
		{
			testFunc: func() (interface{}, error) {
				return &api.ContainerCatalogEntry{
					ID: "id1",
				}, nil
			},
			expectedID: "id1",
			timeout:    time.Second * 5,
		},
		{
			testFunc: func() (interface{}, error) {
				return &api.ContainerCatalogEntry{
					ID: "id1",
				}, nil
			},
			expectedID: "id1",
			timeout:    -1,
		},
		{
			testFunc: func() (interface{}, error) {
				return &api.ContainerCatalogEntry{
					ID: "id1",
				}, errors.New("this is an error")
			},
			expectedID: "id1",
			timeout:    -1,
		},
	}

	for _, tc := range testCases {
		result := WaitForCertificationRequestToSuccess(tc.testFunc, tc.timeout)
		if tc.timeout > time.Second*0 {
			assert.Equal(t, tc.expectedID, result.(*api.ContainerCatalogEntry).ID)
		} else {
			// Note: Cannot cast the result here because the interface returned is nil.
			assert.Nil(t, result)
		}
	}
}

//nolint:funlen
func TestGetContainersToQuery(t *testing.T) {
	generateEnv := func(checkDiscovered bool, CIDs []configuration.ContainerImageIdentifier, CIIs []*provider.Container) *provider.TestEnvironment {
		return &provider.TestEnvironment{
			Config: configuration.TestConfiguration{
				CertifiedContainerInfo:                      CIDs,
				CheckDiscoveredContainerCertificationStatus: checkDiscovered,
			},
			Containers: CIIs,
		}
	}

	testCases := []struct {
		testEnv     *provider.TestEnvironment
		expectedMap map[configuration.ContainerImageIdentifier]bool
	}{
		{ // Test Case #1 - Different images in the map
			testEnv: generateEnv(true, []configuration.ContainerImageIdentifier{
				{
					Name: "image1",
				},
			}, []*provider.Container{
				{
					ContainerImageIdentifier: configuration.ContainerImageIdentifier{
						Name: "image2",
					},
				},
			}),
			expectedMap: map[configuration.ContainerImageIdentifier]bool{
				{Name: "image1"}: true,
				{Name: "image2"}: true,
			},
		},
		{ // Test Case 2 - Map is overwritten with image1
			testEnv: generateEnv(true, []configuration.ContainerImageIdentifier{
				{
					Name: "image1",
				},
			}, []*provider.Container{
				{
					ContainerImageIdentifier: configuration.ContainerImageIdentifier{
						Name: "image1",
					},
				},
			}),
			expectedMap: map[configuration.ContainerImageIdentifier]bool{
				{Name: "image1"}: true,
			},
		},
		{ // Test Case 3 - Empty map
			testEnv:     generateEnv(true, []configuration.ContainerImageIdentifier{}, []*provider.Container{}),
			expectedMap: map[configuration.ContainerImageIdentifier]bool{},
		},
		{ // Test Case 4 - CheckDiscoveredContainerCertificationStatus is false
			testEnv: generateEnv(false, []configuration.ContainerImageIdentifier{
				{
					Name: "image1",
				},
			}, []*provider.Container{
				{ContainerImageIdentifier: configuration.ContainerImageIdentifier{Name: "image2"}},
				{ContainerImageIdentifier: configuration.ContainerImageIdentifier{Name: "image3"}},
			}),
			expectedMap: map[configuration.ContainerImageIdentifier]bool{
				{Name: "image1"}: true,
			},
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedMap, GetContainersToQuery(tc.testEnv))
	}
}
