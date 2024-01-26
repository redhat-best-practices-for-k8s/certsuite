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

package junit_test

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/junit"
)

const (
	testJunitXMLFilename = "success.junit.xml"
	testKey              = "cnf-certification-tests_junit"
)

func TestExtractTestSuiteResults(t *testing.T) {
	junitResults, err := junit.ExportJUnitAsMap(path.Join("testdata", testJunitXMLFilename))
	claim := make(map[string]interface{})
	claim[testKey] = junitResults
	assert.Nil(t, err)
	assert.NotNil(t, junitResults)
	results, err := junit.ExtractTestSuiteResults(claim, testKey)
	assert.Nil(t, err)
	// positive test
	assert.True(t, results["[It] operator Runs test on operators operator-install-status-CSV_INSTALLED"].Passed)
	// negative test
	assert.False(t, results["[It] platform-alteration platform-alteration-boot-params"].Passed)
}
