package nodes

import (
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/diff"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/pkg/claim"
)

// DiffReport holds a summary of node roles and detailed differences per node.
//
// It contains diff.Diffs objects for each role category (CNI, CSI, Hardware) as well as
// an aggregate set of node differences. When a node appears only in one claim file,
// it is marked with a “not found in claim[1|2]” indicator within the diffs. The struct
// serves as the return value for GetDiffReport and implements Stringer to provide a
// table‑formatted representation of all recorded differences.
type DiffReport struct {
	Nodes    *diff.Diffs `json:"nodes"`
	CNI      *diff.Diffs `json:"CNI"`
	CSI      *diff.Diffs `json:"CSI"`
	Hardware *diff.Diffs `json:"hardware"`
}

// String returns a table of differences for each node between two claim files.
//
// It formats the DiffReport as a string suitable for display in a table,
// showing which nodes differ or are missing from either claim file.
// The method implements fmt.Stringer and is used to present
// comparison results to the user.
func (d DiffReport) String() string {
	str := "CLUSTER NODES DIFFERENCES\n"
	str += "-------------------------\n\n"

	if d.Nodes != nil {
		str += d.Nodes.String() + "\n"
	}

	if d.CNI != nil {
		str += d.CNI.String() + "\n"
	}

	if d.CSI != nil {
		str += d.CSI.String() + "\n"
	}

	if d.Hardware != nil {
		str += d.Hardware.String() + "\n"
	}

	return str
}

// GetDiffReport generates a DiffReport from two claim.Nodes.
//
// It compares the CNIs, CSIs, and Hardware sections of the provided node sets
// and aggregates the differences into a DiffReport structure.
// The function returns a pointer to the resulting DiffReport.
func GetDiffReport(claim1Nodes, claim2Nodes *claim.Nodes) *DiffReport {
	return &DiffReport{
		Nodes:    diff.Compare("Nodes", claim1Nodes.NodesSummary, claim2Nodes.NodesSummary, []string{"labels", "annotations"}),
		CNI:      diff.Compare("CNIs", claim1Nodes.CniNetworks, claim2Nodes.CniNetworks, nil),
		CSI:      diff.Compare("CSIs", claim1Nodes.CsiDriver, claim2Nodes.CsiDriver, nil),
		Hardware: diff.Compare("Hardware", claim1Nodes.NodesHwInfo, claim2Nodes.NodesHwInfo, nil),
	}
}
