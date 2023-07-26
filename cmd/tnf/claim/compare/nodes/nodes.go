package nodes

import (
	"fmt"
	"sort"
	"strings"

	"github.com/test-network-function/cnf-certification-test/cmd/tnf/claim/compare/nodes/cnis"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/pkg/claim"
)

const (
	nodeNotFoundIn = "not found in "

	differentCNIs = "CNIs"
	// differentHardware = "hardware"
	// differentCSIs     = "CSIs"
)

type Summary struct {
	MasterNodes       int `json:"masterNodes"`
	WorkerNodes       int `json:"workerNodes"`
	MasterWorkerNodes int `json:"masterAndWorkerNodes"`
}

type NodeDiffReport struct {
	NodeName              string                      `json:"nodeName"`
	Differences           []string                    `json:"differences"`
	CNINetworksDiffReport cnis.CNINetworksDiffReports `json:"cniNetworksDiffReport,omitempty"`
}

type DiffReport struct {
	Summary struct {
		Claim1Summary Summary `json:"claim1"`
		Claim2Summary Summary `json:"claim2"`
	} `json:"summary"`

	NodesDiffReports []NodeDiffReport `json:"nodesDiffReport"`
}

func NodeDiffIsNotFoundIn(diff string) bool {
	return strings.Contains(diff, nodeNotFoundIn)
}

func (d DiffReport) String() string {
	const diffRowFmt = "%-60s%-60s\n"

	if len(d.NodesDiffReports) == 0 {
		return ""
	}

	str := fmt.Sprintf(diffRowFmt, "NODE", "DIFFERENCES")

	for _, nodeDiffReport := range d.NodesDiffReports {
		if len(nodeDiffReport.Differences) == 0 {
			continue
		}

		differences := ""
		for i := range nodeDiffReport.Differences {
			if i != 0 {
				differences += ","
			}
			differences += nodeDiffReport.Differences[i]
		}
		str += fmt.Sprintf(diffRowFmt, nodeDiffReport.NodeName, differences)
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

func getMergedNodeNamesList(claim1Nodes, claim2Nodes *claim.Nodes) []string {
	names := []string{}
	namesMap := map[string]struct{}{}

	for nodeName := range claim1Nodes.NodesSummary {
		namesMap[nodeName] = struct{}{}
	}

	for nodeName := range claim2Nodes.NodesSummary {
		namesMap[nodeName] = struct{}{}
	}

	for name := range namesMap {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

func GetDiffReport(claim1Nodes, claim2Nodes *claim.Nodes) DiffReport {
	diffReport := DiffReport{}

	nodes := getMergedNodeNamesList(claim1Nodes, claim2Nodes)

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
