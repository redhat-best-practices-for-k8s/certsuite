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

/*
Package operator provides CNFCERT tests used to validate operator CNF facets.
*/

package operator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitCsv(t *testing.T) {
	tests := []struct {
		input       string
		expectedCsv string
		expectedNs  string
	}{
		{
			input:       "hazelcast-platform-operator.v5.12.0, ns=tnf",
			expectedCsv: "hazelcast-platform-operator.v5.12.0",
			expectedNs:  "tnf",
		},
		{
			input:       "example-operator.v1.0.0, ns=example-ns",
			expectedCsv: "example-operator.v1.0.0",
			expectedNs:  "example-ns",
		},
		{
			input:       "another-operator.v2.3.1, ns=another-ns",
			expectedCsv: "another-operator.v2.3.1",
			expectedNs:  "another-ns",
		},
		{
			input:       "no-namespace",
			expectedCsv: "no-namespace",
			expectedNs:  "",
		},
		{
			input:       "ns=onlynamespace",
			expectedCsv: "",
			expectedNs:  "onlynamespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SplitCsv(tt.input)
			if result.NameCsv != tt.expectedCsv {
				t.Errorf("splitCsv(%q) got namecsv %q, want %q", tt.input, result.NameCsv, tt.expectedCsv)
			}
			if result.Namespace != tt.expectedNs {
				t.Errorf("splitCsv(%q) got namespace %q, want %q", tt.input, result.Namespace, tt.expectedNs)
			}
		})
	}
}

func TestGetCsvVersion(t *testing.T) {
	testCases := []struct {
		input       string
		expectedCsv CsvNameVersion
	}{
		{
			input:       "example-operator.v1.0.0",
			expectedCsv: CsvNameVersion{Name: "example-operator", Version: "1.0.0"},
		},
		{
			input:       "another-operator.v2.3.1",
			expectedCsv: CsvNameVersion{Name: "another-operator", Version: "2.3.1"},
		},
		{
			input:       "no-version",
			expectedCsv: CsvNameVersion{Name: "no-version", Version: ""},
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.expectedCsv, GetCsvVersion(testCase.input))
	}
}
