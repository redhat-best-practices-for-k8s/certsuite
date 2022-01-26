// Copyright (C) 2020-2021 Red Hat, Inc.
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

package gradetool

import (
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testDataPath = "testdata/"
	goodClaim    = testDataPath + "claim-good.json"
	badClaim     = testDataPath + "claim-missing-field.json"
	goodPolicy   = testDataPath + "policy-good.json"
	badPolicy    = testDataPath + "policy-duplicate-grade.json"
	outPath      = testDataPath + "out.json"
)

var (
	testOutPath = path.Join("..", "..", "test-out.json")
)

func TestMain(m *testing.M) {
	policySchemaPath = path.Join("..", "..", policySchemaPath)
	os.Exit(m.Run())
}

func TestGenerateGrade_Success(t *testing.T) {
	err := GenerateGrade(goodClaim, goodPolicy, testOutPath)
	assert.Nil(t, err)
	assertFilesMatch(t, outPath, testOutPath)
}

func TestGenerateGrade_ErrorInput(t *testing.T) {
	err := GenerateGrade(goodClaim, badPolicy, testOutPath)
	assert.NotNil(t, err)
	err = GenerateGrade(badClaim, goodPolicy, testOutPath)
	assert.NotNil(t, err)
}

func assertFilesMatch(t *testing.T, pathA, pathB string) {
	command := exec.Command("cmp", "-s", pathA, pathB)
	err := command.Run()
	assert.Nil(t, err)
}
