package versions

import (
	"encoding/json"
	"log"

	officialClaimScheme "github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/diff"
)

// DiffReport Represents the differences between two claim versions
//
// This struct holds a pointer to a diff.Diffs object that captures all detected
// changes when comparing two sets of claim versions. The String method formats
// those differences into a human-readable string, or returns an empty
// representation if no differences exist.
type DiffReport struct {
	Diffs *diff.Diffs `json:"differences"`
}

// DiffReport.String Returns a formatted string representation of the diff report
//
// When called on a DiffReport instance, this method checks if its internal
// Diffs field is nil. If it is, it creates an empty Diffs object and returns
// its string form; otherwise, it delegates to the existing Diffs object's
// String method. The resulting string summarizes the differences captured by
// the report.
func (d *DiffReport) String() string {
	if d.Diffs == nil {
		return (&diff.Diffs{}).String()
	}

	return d.Diffs.String()
}

// Compare compares two claim version structures
//
// The function serializes each versions object to JSON, then unmarshals them
// into generic interface values so they can be compared by the diff package. It
// returns a report containing differences between the two sets of versions.
// Errors during marshaling or unmarshaling cause the program to log a fatal
// message.
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
