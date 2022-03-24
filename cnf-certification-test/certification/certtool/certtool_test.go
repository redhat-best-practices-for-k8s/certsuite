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

func TestCompareVersion(t *testing.T) {
	// Note: These values pertain to 'kubeVersion' fields found:
	// https://charts.openshift.io/index.yaml
	testCases := []struct {
		ver1           string
		ver2           string
		expectedOutput bool
	}{
		{
			ver1:           "1.18.1",
			ver2:           ">= 1.19",
			expectedOutput: false,
		},
		{
			ver1:           "1.19.1",
			ver2:           ">= 1.19",
			expectedOutput: true,
		},
		{
			ver1:           "1.19",
			ver2:           ">= 1.16.0 < 1.22.0",
			expectedOutput: true,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, CompareVersion(tc.ver1, tc.ver2))
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
