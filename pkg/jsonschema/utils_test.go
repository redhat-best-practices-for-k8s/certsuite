// Copyright (C) 2021 Red Hat, Inc.
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

package jsonschema_test

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/jsonschema"
)

var (
	schemaFile = "../../schemas/generic-pty.schema.json"
)

var testCases = []struct {
	inputFile             string
	inputSchema           string
	expectedErr           bool
	expectedResultIsValid bool
}{
	{
		inputFile:             "zsh.json",
		inputSchema:           schemaFile,
		expectedErr:           false,
		expectedResultIsValid: true,
	},
	{
		inputFile:             "non_schema_match.json",
		inputSchema:           schemaFile,
		expectedErr:           false,
		expectedResultIsValid: false,
	},
	{
		inputFile:   "empty.json",
		inputSchema: schemaFile,
		expectedErr: true,
	},
	{
		inputFile:   "non_json.json",
		inputSchema: schemaFile,
		expectedErr: true,
	},
	{
		inputFile:   "does_not_exist",
		inputSchema: schemaFile,
		expectedErr: true,
	},
	{
		inputFile:   "zsh.json",
		inputSchema: "does_not_exist",
		expectedErr: true,
	},
}

func getTestCasePath(inputFile string) string {
	return path.Join("testdata", inputFile)
}

func TestValidateJSONAgainstSchemaFile(t *testing.T) {
	for _, testCase := range testCases {
		inputFile := getTestCasePath(testCase.inputFile)
		result, err := jsonschema.ValidateJSONFileAgainstSchema(inputFile, testCase.inputSchema)
		assert.Equal(t, testCase.expectedErr, err != nil)
		if !testCase.expectedErr {
			assert.Equal(t, testCase.expectedResultIsValid, result.Valid())
		}
	}
}
