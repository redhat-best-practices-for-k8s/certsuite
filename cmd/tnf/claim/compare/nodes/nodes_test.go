package nodes

import (
	"testing"

	"github.com/test-network-function/cnf-certification-test/cmd/tnf/pkg/claim"
	"gotest.tools/v3/assert"
	v1 "k8s.io/api/core/v1"
)

func TestGetMergedNodeNamesList(t *testing.T) {
	testCases := []struct {
		name          string
		claim1Nodes   *claim.Nodes
		claim2Nodes   *claim.Nodes
		expectedNames []string
	}{
		{
			name:          "empty structs",
			claim1Nodes:   &claim.Nodes{},
			claim2Nodes:   &claim.Nodes{},
			expectedNames: []string{},
		},
		{
			name: "nodes in claim1 only",
			claim1Nodes: &claim.Nodes{
				NodesSummary: map[string]*v1.Node{
					"node1": {},
					"node2": {}},
			},
			claim2Nodes:   &claim.Nodes{},
			expectedNames: []string{"node1", "node2"},
		},
		{
			name:        "nodes in claim2 only",
			claim1Nodes: &claim.Nodes{},
			claim2Nodes: &claim.Nodes{
				NodesSummary: map[string]*v1.Node{
					"node1": {},
					"node2": {}},
			},
			expectedNames: []string{"node1", "node2"},
		},
		{
			name: "same nodes in both claim files",
			claim1Nodes: &claim.Nodes{
				NodesSummary: map[string]*v1.Node{
					"node1": {},
					"node2": {}},
			},
			claim2Nodes: &claim.Nodes{
				NodesSummary: map[string]*v1.Node{
					"node1": {},
					"node2": {}},
			},
			expectedNames: []string{"node1", "node2"},
		},
		{
			name: "shared nodes in both files but they have an extra different node",
			claim1Nodes: &claim.Nodes{
				NodesSummary: map[string]*v1.Node{
					"node1": {},
					"node2": {},
					"node3": {},
				},
			},
			claim2Nodes: &claim.Nodes{
				NodesSummary: map[string]*v1.Node{
					"node1": {},
					"node2": {},
					"node4": {},
				},
			},
			expectedNames: []string{"node1", "node2", "node3", "node4"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			names := getMergedNodeNamesList(tc.claim1Nodes, tc.claim2Nodes)
			assert.DeepEqual(t, tc.expectedNames, names)
		})
	}
}

func TestNodeDiffIsNotFoundIn(t *testing.T) {
	testCases := []struct {
		name               string
		differences        string
		expectedIsNotFound bool
	}{
		{
			name:               "empty differences string",
			differences:        "",
			expectedIsNotFound: false,
		},
		{
			name:               "single random diff",
			differences:        "diff1",
			expectedIsNotFound: false,
		},
		{
			name:               "multiple random diffs",
			differences:        "diff1,diff2",
			expectedIsNotFound: false,
		},
		{
			name:               "not found with wrong claim file number",
			differences:        "not found in claim0",
			expectedIsNotFound: false,
		},
		{
			name:               "not found with wrong format",
			differences:        "not found in claim1,ipam",
			expectedIsNotFound: false,
		},
		{
			name:               "not found with claim file 1",
			differences:        "not found in claim1",
			expectedIsNotFound: true,
		},
		{
			name:               "not found with claim file 2",
			differences:        "not found in claim2",
			expectedIsNotFound: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedIsNotFound, NodeDiffIsNotFoundIn(tc.differences))
		})
	}
}

// The UT checks the node role summary and the node differences table only.
// The table for CNI/CSI differences on each node will be empty as its stringer method
// is tested in UTs in their corresponding packages.
func TestDiffReportString(t *testing.T) {
	testCases := []struct {
		name           string
		diffReport     DiffReport
		expectedString string
	}{
		{
			name:       "empty diff report",
			diffReport: DiffReport{},
			expectedString: `CLUSTER NODES ROLES SUMMARY
---------------------------
CLAIM     MASTERS   WORKERS   MASTER+WORKER
claim1    0         0         0
claim2    0         0         0

CLUSTER NODES DIFFERENCES
-------------------------
<none>
`,
		},
		{
			name: "one worker+master node with differences in the CNIs",
			diffReport: DiffReport{
				Summary: struct {
					Claim1 RolesSummary `json:"claim1"`
					Claim2 RolesSummary `json:"claim2"`
				}{
					Claim1: RolesSummary{0, 0, 1},
					Claim2: RolesSummary{0, 0, 1},
				},
				NodesDiffReports: []NodeDiffReport{
					{
						NodeName:    "node1",
						Differences: []string{"CNIs"},
					},
				},
			},
			expectedString: `CLUSTER NODES ROLES SUMMARY
---------------------------
CLAIM     MASTERS   WORKERS   MASTER+WORKER
claim1    0         0         1
claim2    0         0         1

CLUSTER NODES DIFFERENCES
-------------------------
NODE                                                        DIFFERENCES
node1                                                       CNIs

NODE: node1
CNI-NETWORK                   DIFFERENCES
`,
		},
		{
			name: "one master and one worker, worker diff in CNIs and CSIs",
			diffReport: DiffReport{
				Summary: struct {
					Claim1 RolesSummary `json:"claim1"`
					Claim2 RolesSummary `json:"claim2"`
				}{
					Claim1: RolesSummary{1, 0, 0},
					Claim2: RolesSummary{0, 1, 0},
				},
				NodesDiffReports: []NodeDiffReport{
					{
						NodeName:    "worker1",
						Differences: []string{"CNIs", "CSIs"},
					},
				},
			},
			expectedString: `CLUSTER NODES ROLES SUMMARY
---------------------------
CLAIM     MASTERS   WORKERS   MASTER+WORKER
claim1    1         0         0
claim2    0         1         0

CLUSTER NODES DIFFERENCES
-------------------------
NODE                                                        DIFFERENCES
worker1                                                     CNIs,CSIs

NODE: worker1
CNI-NETWORK                   DIFFERENCES
`,
		},
		{
			name: "three workers, tree masters, worker1 differs in CNIs, worker2 differs in CSIs",
			diffReport: DiffReport{
				Summary: struct {
					Claim1 RolesSummary `json:"claim1"`
					Claim2 RolesSummary `json:"claim2"`
				}{
					Claim1: RolesSummary{3, 0, 0},
					Claim2: RolesSummary{0, 3, 0},
				},
				NodesDiffReports: []NodeDiffReport{
					{
						NodeName:    "worker1",
						Differences: []string{"CNIs"},
					},
					{
						NodeName:    "worker2",
						Differences: []string{"CSIs"},
					},
				},
			},
			expectedString: `CLUSTER NODES ROLES SUMMARY
---------------------------
CLAIM     MASTERS   WORKERS   MASTER+WORKER
claim1    3         0         0
claim2    0         3         0

CLUSTER NODES DIFFERENCES
-------------------------
NODE                                                        DIFFERENCES
worker1                                                     CNIs
worker2                                                     CSIs

NODE: worker1
CNI-NETWORK                   DIFFERENCES

NODE: worker2
CNI-NETWORK                   DIFFERENCES
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			str := tc.diffReport.String()
			t.Logf("Expected:\n%s\n", tc.expectedString)
			t.Logf("Actual  :\n%s\n", str)
			assert.Equal(t, tc.expectedString, str)
		})
	}
}
