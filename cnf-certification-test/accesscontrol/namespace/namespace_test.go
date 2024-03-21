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

package namespace

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/internal/log"
)

func TestGetInvalidCRsNum(t *testing.T) {
	testCases := []struct {
		invalidCrs         map[string]map[string][]string
		expectedInvalidCRs int
	}{
		{
			invalidCrs: map[string]map[string][]string{
				"cr1": {
					"ns1": {
						"testCRDs",
					},
				},
			},
			expectedInvalidCRs: 1,
		},
		{
			invalidCrs:         map[string]map[string][]string{},
			expectedInvalidCRs: 0,
		},
	}

	for _, tc := range testCases {
		log.SetupLogger(os.Stdout, "INFO")
		result := GetInvalidCRsNum(tc.invalidCrs, log.GetLogger())
		assert.Equal(t, tc.expectedInvalidCRs, result)
	}
}
