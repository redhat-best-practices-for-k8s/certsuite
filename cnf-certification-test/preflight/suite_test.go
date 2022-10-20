// Copyright (C) 2022 Red Hat, Inc.
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

package preflight

import (
	"encoding/json"
	"os"
	"testing"

	plibRuntime "github.com/sebrandon1/openshift-preflight/certification/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
)

func TestGatherTestNamesFromContainerResults(t *testing.T) {
	c := &provider.Container{}
	testData := "testdata/container1.json"

	output, err := os.ReadFile(testData)
	assert.Nil(t, err)

	var tempPreflightResults plibRuntime.Results
	err = json.Unmarshal(output, &tempPreflightResults)
	assert.Nil(t, err)
	c.PreflightResults = tempPreflightResults

	containerList := []*provider.Container{c}
	results := gatherTestNamesFromContainerResults(containerList)

	// Assert a "Passed" result
	assert.Equal(t, results["LayerCountAcceptable"].Metadata().Description, "Checking if container has less than 40 layers.  Too many layers within the container images can degrade container performance.")

	// Assert a "Failed" result
	assert.Equal(t, results["HasLicense"].Metadata().Description, "Checking if terms and conditions applicable to the software including open source licensing information are present. The license must be at /licenses")
	assert.Equal(t, results["HasLicense"].Help().Suggestion, "Create a directory named /licenses and include all relevant licensing and/or terms and conditions as text file(s) in that directory.")
	assert.Equal(t, results["HasLicense"].Metadata().KnowledgeBaseURL, "https://access.redhat.com/documentation/en-us/red_hat_software_certification/8.45/html/red_hat_openshift_software_certification_policy_guide/assembly-requirements-for-container-images_openshift-sw-cert-policy-introduction")
}

func TestGatherTestNamesFromOperatorResults(t *testing.T) {
	op := &provider.Operator{}
	testData := "testdata/operator1.json"

	output, err := os.ReadFile(testData)
	assert.Nil(t, err)

	var tempPreflightResults plibRuntime.Results
	err = json.Unmarshal(output, &tempPreflightResults)
	assert.Nil(t, err)
	op.PreflightResults = tempPreflightResults

	operatorList := []*provider.Operator{op}
	results := gatherTestNamesFromOperatorResults(operatorList)

	// Assert a "Passed" result
	assert.Equal(t, results["ValidateOperatorBundle"].Metadata().Description, "Validating Bundle image that checks if it can validate the content and format of the operator bundle")

	// Assert a "Error" result
	assert.Equal(t, results["ScorecardBasicSpecCheck"].Metadata().Description, "Check to make sure that all CRs have a spec block.")
}
