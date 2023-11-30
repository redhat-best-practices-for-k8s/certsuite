// Copyright (C) 2021-2023 Red Hat, Inc.
// Copyright (C) 2021-2023 Red Hat, Inc.
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

package identifiers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/stringhelper"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
)

func TestGetGinkgoTestIDAndLabels(t *testing.T) {
	testCases := []struct {
		testIdentifier     claim.Identifier
		expectedIDOutput   string
		expectedTagsOutput []string
	}{
		{
			testIdentifier: claim.Identifier{
				Id:    "test-id-1",
				Suite: "test-suite",
				Tags:  "tag1",
			},
			expectedIDOutput:   "test-id-1",
			expectedTagsOutput: []string{"tag1", "test-suite"},
		},
		{
			testIdentifier: claim.Identifier{
				Id:    "test-id-2",
				Suite: "test-suite2",
				Tags:  "tag1,tag2,tag3",
			},
			expectedIDOutput:   "test-id-2",
			expectedTagsOutput: []string{"tag1", "tag2", "tag3", "test-suite2"},
		},
	}

	for _, tc := range testCases {
		resultID, resultTags := GetGinkgoTestIDAndLabels(tc.testIdentifier)
		assert.Equal(t, tc.expectedIDOutput, resultID)
		for _, e := range tc.expectedTagsOutput {
			assert.True(t, stringhelper.StringInSlice(resultTags, e, false))
		}
	}
}
