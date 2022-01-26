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
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/test-network-function/cnf-certification-test/pkg/jsonschema"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf/identifier"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
	"github.com/xeipuuv/gojsonschema"
)

const (
	outputFilePermissions = 420
)

var (
	policySchemaPath = path.Join("schemas", "gradetool-policy-schema.json")
)

// Grade is a single grade object from policy file
type Grade struct {
	GradeName            string
	RequiredPassingTests []identifier.Identifier
	NextGrade            *Grade
}

// Policy is the object in the policy file
type Policy struct {
	Grades Grade
}

// GradeResult is the grade output object
type GradeResult struct {
	Name    string
	Propose bool
	Pass    []identifier.Identifier
	Fail    []identifier.Identifier
}

// GenerateGrade outputs a grade file based on input test results and input grading policy
func GenerateGrade(resultsPath, policyPath, outputPath string) error {
	err := validatePolicySchema(policyPath)
	if err != nil {
		return err
	}

	policyObj := Policy{}
	err = unmarshalFromFile(policyPath, &policyObj)
	if err != nil {
		return err
	}

	err = validatePolicy(&policyObj)
	if err != nil {
		return err
	}

	claimObj := claim.Root{}
	err = unmarshalFromFile(resultsPath, &claimObj)
	if err != nil {
		return err
	}

	// start grading process
	gradingOutput, err := doGrading(policyObj, claimObj.Claim.Results)
	if err != nil {
		return err
	}

	// write output
	err = generateOutput(gradingOutput, outputPath)
	if err != nil {
		return err
	}

	return nil
}

// NewGradeResult creates a new object without nil properties
func NewGradeResult(gradeName string) GradeResult {
	emptySlice := []identifier.Identifier{}
	return GradeResult{gradeName, false, emptySlice, emptySlice}
}

func generateTestResultsKey(id identifier.Identifier) string {
	return fmt.Sprintf("{\"url\":%q,\"version\":%q}", id.URL, id.SemanticVersion)
}

func doGrading(policy Policy, results map[string]interface{}) (interface{}, error) {
	gradingOutput := []GradeResult{}

	grade := &policy.Grades
	previousGradePassed := true

	for grade != nil {
		gradeResult := NewGradeResult(grade.GradeName)
		for _, id := range grade.RequiredPassingTests {
			resultsKey := generateTestResultsKey(id)
			results, ok := results[resultsKey]
			if !ok {
				gradeResult.Fail = append(gradeResult.Fail, id)
				continue
			}
			testPass, err := processTestResults(results)
			if err != nil {
				return nil, err
			}
			if testPass {
				gradeResult.Pass = append(gradeResult.Pass, id)
			} else {
				gradeResult.Fail = append(gradeResult.Fail, id)
			}
		}
		if previousGradePassed && len(gradeResult.Fail) == 0 {
			gradeResult.Propose = true
		}
		gradingOutput = append(gradingOutput, gradeResult)
		grade = grade.NextGrade
	}

	return gradingOutput, nil
}

func processTestResults(results interface{}) (bool, error) {
	pass := false
	var resultsTyped []interface{}
	resultsTyped, ok := results.([]interface{})
	if !ok {
		return pass, fmt.Errorf("the test results object is not of expected type. "+
			"found: %T. expected: %T", results, resultsTyped)
	}
	for _, result := range resultsTyped {
		resultTyped, ok := result.(map[string]interface{})
		if !ok {
			return pass, fmt.Errorf("the test result object is not of expected type. "+
				"found: %T. expected: %T", result, resultTyped)
		}
		val, ok := resultTyped["passed"]
		if !ok {
			return pass, fmt.Errorf("the field 'passed' is missing in test result")
		}
		pass, ok = val.(bool)
		if !ok {
			return pass, fmt.Errorf("field 'passed' is not of type bool")
		}
	}
	return pass, nil
}

func generateOutput(outputObj interface{}, outputPath string) error {
	outputBytes, err := json.MarshalIndent(outputObj, "", "    ")
	if err != nil {
		return err
	}
	err = os.WriteFile(outputPath, outputBytes, outputFilePermissions)
	if err != nil {
		return err
	}
	return nil
}

func validatePolicy(policyObj *Policy) error {
	grade := &policyObj.Grades
	gradeNames := map[string]bool{}
	for grade != nil {
		_, ok := gradeNames[grade.GradeName]
		if ok {
			return fmt.Errorf("duplicate grade name %s in policy", grade.GradeName)
		}
		gradeNames[grade.GradeName] = true

		grade = grade.NextGrade
	}
	return nil
}

func validatePolicySchema(policyPath string) error {
	validationResult, err := jsonschema.ValidateJSONFileAgainstSchema(policyPath, policySchemaPath)
	if err != nil || !validationResult.Valid() {
		validationErrors := []gojsonschema.ResultError{}
		if validationResult != nil {
			validationErrors = validationResult.Errors()
		}
		return fmt.Errorf("invalid policy file %s. Error: %s. Parse result: %v", policyPath, err, validationErrors)
	}
	return nil
}

func unmarshalFromFile(jsonPath string, obj interface{}) error {
	jsonBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonBytes, obj)
	if err != nil {
		return err
	}
	return nil
}
