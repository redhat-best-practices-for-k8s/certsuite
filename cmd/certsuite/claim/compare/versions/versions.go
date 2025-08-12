package versions

import (
	"encoding/json"
	"log"

	officialClaimScheme "github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/diff"
)

// DiffReport represents a report of differences between two versions.
//
// It contains the computed diffs and provides a String method to format them as a readable string.
type DiffReport struct {
	Diffs *diff.Diffs `json:"differences"`
}

// String formats the diff report into a human‑readable string.
//
// It renders the stored differences between two version sets in a
// readable form, suitable for printing or logging. The method takes no
// parameters and returns a single string containing the formatted
// report.
func (d *DiffReport) String() string {
	if d.Diffs == nil {
		return (&diff.Diffs{}).String()
	}

	return d.Diffs.String()
}

// Compare compares two official claim scheme versions and returns a report of differences.
//
// It accepts pointers to two Versions structs, serializes them for comparison,
// performs a field‑by‑field diff, and constructs a DiffReport that summarizes any
// discrepancies found between the input versions. The function logs fatal errors if
// marshalling or unmarshalling fails during the process.
func Compare(claim1Versions, claim2Versions *officialClaimScheme.Versions) *DiffReport {
	// Convert the versions struct type to agnostic map[string]interface{} objects so
	// it can be compared using the diff.Compare func.

	bytes1, err := json.Marshal(claim1Versions)
	if err != nil {
		log.Fatalf("Failed to marshal versions from claim 1: %v\nq", err)
	}

	bytes2, err := json.Marshal(claim2Versions)
	if err != nil {
		log.Fatalf("Failed to marshal versions from claim 2: %v\n", err)
	}

	// Now let's unmarshal them into interface{} vars
	var v1, v2 interface{}
	err = json.Unmarshal(bytes1, &v1)
	if err != nil {
		log.Fatalf("Failed to unmarshal versions from claim 1: %v\n", err)
	}

	err = json.Unmarshal(bytes2, &v2)
	if err != nil {
		log.Fatalf("Failed to unmarshal versions from claim 2: %v\n", err)
	}

	return &DiffReport{
		Diffs: diff.Compare("VERSIONS", v1, v2, nil),
	}
}
