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

package jsonschema

import (
	"os"

	"github.com/xeipuuv/gojsonschema"
)

// ValidateJSONFileAgainstSchema validates a given file against the supplied JSON schema.
func ValidateJSONFileAgainstSchema(filename, schemaPath string) (*gojsonschema.Result, error) {
	inputBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ValidateJSONAgainstSchema(inputBytes, schemaPath)
}

// ValidateJSONAgainstSchema validates a given byte array against the supplied JSON schema.
func ValidateJSONAgainstSchema(inputBytes []byte, schemaPath string) (*gojsonschema.Result, error) {
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}

	schemaLoader := gojsonschema.NewStringLoader(string(schemaBytes))
	if err != nil {
		return nil, err
	}

	inputLoader := gojsonschema.NewStringLoader(string(inputBytes))
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, err
	}
	return schema.Validate(inputLoader)
}
