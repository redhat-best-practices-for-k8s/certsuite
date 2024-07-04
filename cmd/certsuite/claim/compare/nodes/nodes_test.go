package nodes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/cmd/certsuite/claim/compare/diff"
)

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
			name: "CNI Differences",
			diffReport: DiffReport{
				CNI: &diff.Diffs{
					Name: "CNI",
					Fields: []diff.FieldDiff{
						{
							FieldPath:   "/path/to/field1",
							Claim1Value: "value1",
							Claim2Value: "value2",
						},
						{
							FieldPath:   "/path/to/another/different/field2",
							Claim1Value: 5,
							Claim2Value: 6,
						},
					},
				},
			},
			expectedString: `CLUSTER NODES DIFFERENCES
-------------------------

CNI: Differences
FIELD                                 CLAIM 1     CLAIM 2
/path/to/field1                       value1      value2
/path/to/another/different/field2     5           6

CNI: Only in CLAIM 1
<none>

CNI: Only in CLAIM 2
<none>

`,
		},
		{
			name: "CNI Differences, CNI Differences & Hardware fields missing in both files.",
			diffReport: DiffReport{
				CNI: &diff.Diffs{
					Name: "CNI",
					Fields: []diff.FieldDiff{
						{
							FieldPath:   "/path/to/cni/field1",
							Claim1Value: "value1",
							Claim2Value: "value2",
						},
					},
				},
				CSI: &diff.Diffs{
					Name: "CSI",
					Fields: []diff.FieldDiff{
						{
							FieldPath:   "/path/to/csi/field1",
							Claim1Value: "value2",
							Claim2Value: "value3",
						},
					},
				},
				Hardware: &diff.Diffs{
					Name:   "Hardware",
					Fields: []diff.FieldDiff{},
					FieldsInClaim1Only: []string{
						"/path/to/field/not/found/in/claim2_1=value1",
						"/path/to/another/field/not/found/in/claim2=value2",
					},
					FieldsInClaim2Only: []string{"/path/to/weird/field/not/found/in/claim1=value3"},
				},
			},
			expectedString: `CLUSTER NODES DIFFERENCES
-------------------------

CNI: Differences
FIELD                   CLAIM 1     CLAIM 2
/path/to/cni/field1     value1      value2

CNI: Only in CLAIM 1
<none>

CNI: Only in CLAIM 2
<none>

CSI: Differences
FIELD                   CLAIM 1     CLAIM 2
/path/to/csi/field1     value2      value3

CSI: Only in CLAIM 1
<none>

CSI: Only in CLAIM 2
<none>

Hardware: Differences
FIELD     CLAIM 1     CLAIM 2
<none>

Hardware: Only in CLAIM 1
/path/to/field/not/found/in/claim2_1=value1
/path/to/another/field/not/found/in/claim2=value2

Hardware: Only in CLAIM 2
/path/to/weird/field/not/found/in/claim1=value3

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
