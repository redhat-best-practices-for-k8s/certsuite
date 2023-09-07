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

package autodiscover

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
)

func TestNamespacesListToStringList(t *testing.T) {
	testCases := []struct {
		testList       []configuration.Namespace
		expectedOutput []string
	}{
		{
			testList: []configuration.Namespace{
				{
					Name: "ns1",
				},
				{
					Name: "ns2",
				},
			},
			expectedOutput: []string{"ns1", "ns2"},
		},
		{
			testList:       []configuration.Namespace{},
			expectedOutput: nil,
		},
		{
			testList: []configuration.Namespace{
				{Name: "name1"},
				{Name: "name1"},
			},
			expectedOutput: []string{"name1", "name1"},
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, namespacesListToStringList(tc.testList))
	}
}
