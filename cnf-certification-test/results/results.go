// Copyright (C) 2021-2022 Red Hat, Inc.
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

package results

import (
	"fmt"

	"github.com/onsi/ginkgo/v2/types"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"

	"github.com/test-network-function/test-network-function-claim/pkg/claim"
)

// results is the results map
var results = map[string][]claim.Result{}

// RecordResult is a hook provided to save aspects of the ginkgo.GinkgoTestDescription for a given claim.Identifier.
// Multiple results for a given identifier are aggregated as an array under the same key.
func RecordResult(report types.SpecReport) { //nolint:gocritic // From Ginkgo
	if claimID, ok := identifiers.TestIDToClaimID[report.LeafNodeText]; ok {
		testText := identifiers.Catalog[claimID].Description
		results[report.LeafNodeText] = append(results[report.LeafNodeText], claim.Result{
			Duration:           int(report.RunTime.Nanoseconds()),
			FailureLocation:    report.FailureLocation().String(),
			FailureLineContent: report.FailureLocation().ContentsOfLine(),
			TestText:           testText,
			FailureReason:      report.FailureMessage(),
			State:              report.State.String(),
			StartTime:          report.StartTime.String(),
			EndTime:            report.EndTime.String(),
			CapturedTestOutput: report.CapturedGinkgoWriterOutput,
			TestID:             &claimID,
		})
	} else {
		panic(fmt.Sprintf("TestID %s has no corresponding Claim ID", report.LeafNodeText))
	}
}

// GetReconciledResults is a function added to aggregate a Claim's results.  Due to the limitations of
// test-network-function-claim's Go Client, results are generalized to map[string]interface{}.  This method is needed
// to take the results gleaned from JUnit output, and to combine them with the contexts built up by subsequent calls to
// RecordResult.  The combination of the two forms a Claim's results.
func GetReconciledResults() map[string]interface{} {
	resultMap := make(map[string]interface{})
	for key, vals := range results {
		// initializes the result map, if necessary
		if _, ok := resultMap[key]; !ok {
			resultMap[key] = make([]claim.Result, 0)
		}
		for _, val := range vals { //nolint:gocritic // Only done once at the end
			resultMap[key] = append(resultMap[key].([]claim.Result), val)
		}
	}
	return resultMap
}
