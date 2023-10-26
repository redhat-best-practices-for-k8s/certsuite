package versions

import (
	"encoding/json"
	"log"

	"github.com/test-network-function/cnf-certification-test/cmd/tnf/claim/compare/diff"
	officialClaimScheme "github.com/test-network-function/test-network-function-claim/pkg/claim"
)

type DiffReport struct {
	Diffs *diff.Diffs `json:"differences"`
}

func (d *DiffReport) String() string {
	if d.Diffs == nil {
		return (&diff.Diffs{}).String()
	}

	return d.Diffs.String()
}

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
