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

package certification

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
)

func TestGetContainersToQuery(t *testing.T) {
	generateEnv := func(certStatus bool) *provider.TestEnvironment {
		return &provider.TestEnvironment{
			Config: configuration.TestConfiguration{
				CheckDiscoveredContainerCertificationStatus: certStatus,
				CertifiedContainerInfo: []configuration.ContainerImageIdentifier{
					{
						Name:       "test2",
						Repository: "repo1",
					},
				},
			},
			Containers: []*provider.Container{
				{
					ContainerImageIdentifier: configuration.ContainerImageIdentifier{
						Name:       "test1",
						Repository: "repo1",
					},
				},
			},
		}
	}

	testCases := []struct {
		testCertStatus bool
		expectedOutput map[configuration.ContainerImageIdentifier]bool
	}{
		{
			testCertStatus: true,
			expectedOutput: map[configuration.ContainerImageIdentifier]bool{
				{
					Name:       "test1",
					Repository: "repo1",
				}: true,
				{
					Name:       "test2",
					Repository: "repo1",
				}: true,
			},
		},
		{
			testCertStatus: false,
			expectedOutput: map[configuration.ContainerImageIdentifier]bool{
				{
					Name:       "test2",
					Repository: "repo1",
				}: true,
			},
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, getContainersToQuery(generateEnv(tc.testCertStatus)))
	}
}
