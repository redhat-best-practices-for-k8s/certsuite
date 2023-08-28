package nodes

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/test-network-function/cnf-certification-test/cmd/tnf/claim/compare/nodes/cnis"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/pkg/claim"
	v1 "k8s.io/api/core/v1"
)

const (
	nodeNotFoundIn = "not found in "

	// List of possible differences for nodes.
	differentCNIs = "CNIs"
	// differentHardware = "hardware"
	// differentCSIs     = "CSIs"
)

var (
	masterNodeLabels = map[string]bool{"node-role.kubernetes.io/master": true, "node-role.kubernetes.io/control-plane": true}
	workerLabels     = map[string]bool{"node-role.kubernetes.io/worker": true}
)

type RolesSummary struct {
	MasterNodes       int `json:"masterNodes"`
	WorkerNodes       int `json:"workerNodes"`
	MasterWorkerNodes int `json:"masterAndWorkerNodes"`
}

// Structure to hold the differences found in a node.
// The slice "Differences" holds an entry for each section (CNI, CSI, hardware) that
// has differences in configuration.
// Each section has its own slice with a differences report per node.
type NodeDiffReport struct {
	NodeName    string   `json:"nodeName"`
	Differences []string `json:"differences"`

	// CNINetworksDiffReport is a slice and every entry has a report of CNI networks differences
	// for each node.
	CNINetworksDiffReport cnis.CNINetworksDiffReports `json:"cniNetworksDiffReport,omitempty"`
}

// Structure that holds a summary of nodes roles and a slice of NodeDiffReports,
// one per node found in both claim files. In case one node only exists in one
// claim file, it will be marked as "not found in claim[1|2]".
type DiffReport struct {
	Summary struct {
		Claim1 RolesSummary `json:"claim1"`
		Claim2 RolesSummary `json:"claim2"`
	} `json:"nodesRolesSummary"`

	NodesDiffReports []NodeDiffReport `json:"nodesDiffReport"`
}

// Helper function to parse a string and returns true in case it's
// a "not found in claim[1|2]".
func NodeDiffIsNotFoundIn(diff string) bool {
	r := regexp.MustCompile("^" + nodeNotFoundIn + "claim[1|2]$")
	return r.MatchString(diff)
}

// Stringer method to show in a table the the differences found on each node
// appearing on both claim files. If a node only appears in one claim file, it
// will be flagged as "not found in claim[1|2]"
//
// CLUSTER NODES ROLES SUMMARY
// ---------------------------
// CLAIM     MASTERS   WORKERS   MASTER+WORKER
// claim1    0         0         3
// claim2    0         1         3

// CLUSTER NODES DIFFERENCES
// -------------------------
// NODE                                                        DIFFERENCES
// clus0-0                                                     CNIs
// ...                                                         CSIs,hardware
// clus0-7                                                     not found in claim1
//
// NODE: clus0-0
// CNI-NETWORK                   DIFFERENCES
// crio                          plugins
//
// NODE: clus0-0, CNI-NETWORK: crio
// PLUGIN                        DIFFERENCES
// bridge                        ipMasq
//
// ...
//
// The previous example shows that the crio network is not the same in both claim files
// for node clus0-0. The difference column shows a list of "fields" whose values
// are different. Node "clus0-7" was only found in claim2.
func (d DiffReport) String() string {
	const (
		rolesSummaryHeaderFmt = "%-10s%-10s%-10s%-s\n"
		rolesSummaryRowFmt    = "%-10s%-10d%-10d%-d\n"
		nodeDiffsRowFmt       = "%-60s%-s\n"
	)

	str := "CLUSTER NODES ROLES SUMMARY\n"
	str += "---------------------------\n"
	str += fmt.Sprintf(rolesSummaryHeaderFmt, "CLAIM", "MASTERS", "WORKERS", "MASTER+WORKER")
	str += fmt.Sprintf(rolesSummaryRowFmt, "claim1", d.Summary.Claim1.MasterNodes, d.Summary.Claim1.WorkerNodes, d.Summary.Claim1.MasterWorkerNodes)
	str += fmt.Sprintf(rolesSummaryRowFmt, "claim2", d.Summary.Claim2.MasterNodes, d.Summary.Claim2.WorkerNodes, d.Summary.Claim2.MasterWorkerNodes)
	str += "\n"

	str += "CLUSTER NODES DIFFERENCES\n"
	str += "-------------------------\n"
	if len(d.NodesDiffReports) == 0 {
		str += "<none>\n"
		return str
	}

	str += fmt.Sprintf(nodeDiffsRowFmt, "NODE", "DIFFERENCES")

	for _, nodeDiffReport := range d.NodesDiffReports {
		differences := ""
		for i := range nodeDiffReport.Differences {
			if i != 0 {
				differences += ","
			}
			differences += nodeDiffReport.Differences[i]
		}
		str += fmt.Sprintf(nodeDiffsRowFmt, nodeDiffReport.NodeName, differences)
	}

	// Print section differences (CNIs, CSIs, hardware...) of each node.
	for _, nodeDiffReport := range d.NodesDiffReports {
		// Do nothing in case the only diff is that the node doesn't appear in the other claim file.
		if len(nodeDiffReport.Differences) == 1 && NodeDiffIsNotFoundIn(nodeDiffReport.Differences[0]) {
			continue
		}

		str += "\nNODE: " + nodeDiffReport.NodeName + "\n"
		str += nodeDiffReport.CNINetworksDiffReport.String()

		for _, netDiffReport := range nodeDiffReport.CNINetworksDiffReport {
			// Do nothing in case the only diff is that the network doesn't appear in the other claim file.
			if len(netDiffReport.Differences) == 1 && cnis.NetworkDiffIsNotFoundIn(netDiffReport.Differences[0]) {
				continue
			}

			str += "\nNODE: " + nodeDiffReport.NodeName + ", CNI-NETWORK: " + netDiffReport.NetworkName + "\n"
			str += netDiffReport.PluginsDiffReports.String()
		}
	}

	return str
}

// Helper function that returns a sorted list of unique node names found in
// two slices claim.Nodes.
func getMergedNodeNamesList(claim1NodesSummary, claim2NodesSummary map[string]*v1.Node) []string {
	names := []string{}
	namesMap := map[string]struct{}{}

	for nodeName := range claim1NodesSummary {
		namesMap[nodeName] = struct{}{}
	}

	for nodeName := range claim2NodesSummary {
		namesMap[nodeName] = struct{}{}
	}

	for name := range namesMap {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

func isMasterNode(node *v1.Node) bool {
	for label := range node.Labels {
		if masterNodeLabels[label] {
			return true
		}
	}
	return false
}

func isWorkerNode(node *v1.Node) bool {
	for label := range node.Labels {
		if workerLabels[label] {
			return true
		}
	}
	return false
}

func getRolesSummary(nodesSummary map[string]*v1.Node) RolesSummary {
	summary := RolesSummary{}

	for _, node := range nodesSummary {
		if isMasterNode(node) {
			if isWorkerNode(node) {
				summary.MasterWorkerNodes++
			} else {
				summary.MasterNodes++
			}
		} else if isWorkerNode(node) {
			summary.WorkerNodes++
		}
	}

	return summary
}

// Generates a DiffReport from two pointers to claim.Nodes. The report consists
// of a Summary of Nodes Roles found in both files, as well as a list of
// NodeDiffReport objects, one per node.
func GetDiffReport(claim1Nodes, claim2Nodes *claim.Nodes) DiffReport {
	diffReport := DiffReport{
		Summary: struct {
			Claim1 RolesSummary `json:"claim1"`
			Claim2 RolesSummary `json:"claim2"`
		}{
			Claim1: getRolesSummary(claim1Nodes.NodesSummary),
			Claim2: getRolesSummary(claim2Nodes.NodesSummary),
		},
	}

	nodes := getMergedNodeNamesList(claim1Nodes.NodesSummary, claim2Nodes.NodesSummary)

	for _, node := range nodes {
		nodeDiffReport := NodeDiffReport{NodeName: node}

		_, found := claim1Nodes.NodesSummary[node]
		if !found {
			nodeDiffReport.Differences = append(nodeDiffReport.Differences, nodeNotFoundIn+"claim1")
			diffReport.NodesDiffReports = append(diffReport.NodesDiffReports, nodeDiffReport)
			continue
		}

		_, found = claim2Nodes.NodesSummary[node]
		if !found {
			nodeDiffReport.Differences = append(nodeDiffReport.Differences, nodeNotFoundIn+"claim2")
			diffReport.NodesDiffReports = append(diffReport.NodesDiffReports, nodeDiffReport)
			continue
		}

		nodeDiffReport.CNINetworksDiffReport = cnis.GetDiffReports(claim1Nodes.CniNetworks[node], claim2Nodes.CniNetworks[node])
		if len(nodeDiffReport.CNINetworksDiffReport) > 0 {
			nodeDiffReport.Differences = append(nodeDiffReport.Differences, differentCNIs)
		}

		if len(nodeDiffReport.Differences) > 0 {
			diffReport.NodesDiffReports = append(diffReport.NodesDiffReports, nodeDiffReport)
		}
	}

	return diffReport
}
