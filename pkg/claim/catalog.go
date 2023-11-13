// Copyright (C) 2023 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public
// License as published by the Free Software Foundation; either version 2 of the License, or (at your option) any later
// version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied
// warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along with this program; if not, write to the Free
// Software Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package claim

import (
	"strings"
)

// TestCaseDescription describes a JUnit test case.
type TestCaseDescription struct {
	// Identifier is the unique test identifier.
	Identifier Identifier `json:"identifier" yaml:"identifier"`

	// Description is a helpful description of the purpose of the test case.
	Description string `json:"description" yaml:"description"`

	// Remediation is an optional suggested remediation for passing the test.
	Remediation string `json:"remediation,omitempty" yaml:"remediation,omitempty"`

	// BestPracticeReference is a helpful best practice references of the test case.
	BestPracticeReference string `json:"BestPracticeReference" yaml:"BestPracticeReference"`

	// ExceptionProcess will show any possible exception processes documented for partners to follow.
	ExceptionProcess string `json:"exceptionProcess,omitempty" yaml:"exceptionProcess,omitempty"`

	// Tags will show all of the ginkgo tags that the test case applies to
	Tags string `json:"tags,omitempty" yaml:"tags,omitempty"`

	// Whether or not automated tests exist for the test case. Not to be rendered.
	Qe bool `json:"qe" yaml:"qe"`

	// classification for each test case
	CategoryClassification map[string]string `json:"categoryclassification" yaml:"categoryclassification"`
	/* an example to how it CategoryClassification would be
	   {
	   	"ForTelco": "Mandatory",
	   	"FarEdge" : "Optional",
	   	"ForNonTelco": "Optional",
	   	"ForVZ": "Mandatory"
	      }*/
}

func formTestTags(tags ...string) string {
	return strings.Join(tags, ",")
}

//nolint:lll
func BuildTestCaseDescription(testID, suiteName, description, remediation, exception, reference string, qe bool, categoryclassification map[string]string, tags ...string) (TestCaseDescription, Identifier) {
	aID := Identifier{
		Tags:  formTestTags(tags...),
		Id:    suiteName + "-" + testID,
		Suite: suiteName,
	}
	aTCDescription := TestCaseDescription{}
	aTCDescription.Identifier = aID
	aTCDescription.Description = description
	aTCDescription.Remediation = remediation
	aTCDescription.ExceptionProcess = exception
	aTCDescription.BestPracticeReference = reference
	aTCDescription.Tags = strings.Join(tags, ",")
	aTCDescription.Qe = qe
	aTCDescription.CategoryClassification = categoryclassification
	return aTCDescription, aID
}
