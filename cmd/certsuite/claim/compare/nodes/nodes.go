package nodes

import (
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/diff"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/pkg/claim"
)

// DiffReport Summarizes differences between two node claims
//
// It aggregates comparison results for nodes, CNI networks, CSI drivers, and
// hardware information into separate diff objects. Each field holds a report of
// changes or missing entries between the two provided claim files. The struct
// provides a consolidated view that can be rendered as a human‑readable
// string.
type DiffReport struct {
	Nodes    *diff.Diffs `json:"nodes"`
	CNI      *diff.Diffs `json:"CNI"`
	CSI      *diff.Diffs `json:"CSI"`
	Hardware *diff.Diffs `json:"hardware"`
}

// DiffReport.String Formats node differences into a readable table
//
// It builds a string starting with a header and separator, then appends the
// string representations of any non‑nil subreports for Nodes, CNI, CSI, and
// Hardware, each followed by a newline. The resulting text lists discrepancies
// found in cluster nodes across two claim files.
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

// GetDiffReport Creates a report of differences between two node claim sets
//
// This function takes pointers to two node claim structures and returns a
// DiffReport containing four diff objects: Nodes, CNIs, CSIs, and Hardware.
// Each field is produced by calling the Compare helper with appropriate data
// slices and optional filters for labels and annotations. The resulting report
// aggregates all differences for downstream display or analysis.
func GetDiffReport(claim1Nodes, claim2Nodes *claim.Nodes) *DiffReport {
	return &DiffReport{
		Nodes:    diff.Compare("Nodes", claim1Nodes.NodesSummary, claim2Nodes.NodesSummary, []string{"labels", "annotations"}),
		CNI:      diff.Compare("CNIs", claim1Nodes.CniNetworks, claim2Nodes.CniNetworks, nil),
		CSI:      diff.Compare("CSIs", claim1Nodes.CsiDriver, claim2Nodes.CsiDriver, nil),
		Hardware: diff.Compare("Hardware", claim1Nodes.NodesHwInfo, claim2Nodes.NodesHwInfo, nil),
	}
}
