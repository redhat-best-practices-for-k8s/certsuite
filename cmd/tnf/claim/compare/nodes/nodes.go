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

type NodeDiffReport struct {
	NodeName              string                      `json:"nodeName"`
	Differences           []string                    `json:"differences"`
	CNINetworksDiffReport cnis.CNINetworksDiffReports `json:"cniNetworksDiffReport,omitempty"`
}

type DiffReport struct {
	Summary struct {
		Claim1 RolesSummary `json:"claim1"`
		Claim2 RolesSummary `json:"claim2"`
	} `json:"nodesRolesSummary"`

	NodesDiffReports []NodeDiffReport `json:"nodesDiffReport"`
}

func NodeDiffIsNotFoundIn(diff string) bool {
	r := regexp.MustCompile("^" + nodeNotFoundIn + "claim[1|2]$")
	return r.MatchString(diff)
}

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
