// Copyright (C) 2020-2024 Red Hat, Inc.
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

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func TestGetContainersToQuery(t *testing.T) {
	var testEnv = provider.TestEnvironment{
		Containers: []*provider.Container{
			{
				ContainerImageIdentifier: provider.ContainerImageIdentifier{
					Repository: "test1",
					Registry:   "repo1",
				},
			},
			{
				ContainerImageIdentifier: provider.ContainerImageIdentifier{
					Repository: "test2",
					Registry:   "repo2",
				},
			},
		},
	}

	testCases := []struct {
		expectedOutput map[provider.ContainerImageIdentifier]bool
	}{
		{
			expectedOutput: map[provider.ContainerImageIdentifier]bool{
				{
					Repository: "test1",
					Registry:   "repo1",
				}: true,
				{
					Repository: "test2",
					Registry:   "repo2",
				}: true,
			},
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, getContainersToQuery(&testEnv))
	}
}
