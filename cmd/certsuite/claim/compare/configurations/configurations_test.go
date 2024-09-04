package configurations

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/diff"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/pkg/claim"
	"github.com/stretchr/testify/assert"
)

func TestGetDiffReport(t *testing.T) {
	tests := []struct {
		name            string
		configs1        *claim.Configurations
		configs2        *claim.Configurations
		expectedDiff    *DiffReport
		expectedDiffStr string
	}{
		{
			name:     "empty configs",
			configs1: &claim.Configurations{},
			configs2: &claim.Configurations{},
			expectedDiff: &DiffReport{
				Config:         &diff.Diffs{Name: "Cert Suite Configuration"},
				AbnormalEvents: AbnormalEventsCount{},
			},
			expectedDiffStr: `CONFIGURATIONS
--------------

Cert Suite Configuration: Differences
FIELD     CLAIM 1     CLAIM 2
<none>

Cert Suite Configuration: Only in CLAIM 1
<none>

Cert Suite Configuration: Only in CLAIM 2
<none>

Cluster abnormal events count
CLAIM 1     CLAIM 2
0           0
`,
		},
		{
			name: "same config with one field and two abnormal events",
			configs1: &claim.Configurations{
				Config: map[string]interface{}{
					"field1": "value1",
				},
				AbnormalEvents: []interface{}{"event1", "event2"},
			},
			configs2: &claim.Configurations{
				Config: map[string]interface{}{
					"field1": "value1",
				},
				AbnormalEvents: []interface{}{"event1", "event2"},
			},
			expectedDiff: &DiffReport{
				Config: &diff.Diffs{Name: "Cert Suite Configuration"},
				AbnormalEvents: AbnormalEventsCount{
					Claim1: 2,
					Claim2: 2,
				},
			},
			expectedDiffStr: `CONFIGURATIONS
--------------

Cert Suite Configuration: Differences
FIELD     CLAIM 1     CLAIM 2
<none>

Cert Suite Configuration: Only in CLAIM 1
<none>

Cert Suite Configuration: Only in CLAIM 2
<none>

Cluster abnormal events count
CLAIM 1     CLAIM 2
2           2
`,
		},
		{
			name: "different configs",
			configs1: &claim.Configurations{
				Config: map[string]interface{}{
					"field1": "value1",
				},
				AbnormalEvents: []interface{}{"event1"},
			},
			configs2: &claim.Configurations{
				Config: map[string]interface{}{
					"field1": "value11",
					"field2": map[string]interface{}{"subfield1": 58},
				},
				AbnormalEvents: []interface{}{"event1", "event2"},
			},
			expectedDiff: &DiffReport{
				Config: &diff.Diffs{
					Name: "Cert Suite Configuration",
					Fields: []diff.FieldDiff{{
						FieldPath:   "/field1",
						Claim1Value: "value1",
						Claim2Value: "value11",
					}},
					FieldsInClaim2Only: []string{"/field2/subfield1=58"},
				},
				AbnormalEvents: AbnormalEventsCount{
					Claim1: 1,
					Claim2: 2,
				},
			},
			expectedDiffStr: `CONFIGURATIONS
--------------

Cert Suite Configuration: Differences
FIELD       CLAIM 1     CLAIM 2
/field1     value1      value11

Cert Suite Configuration: Only in CLAIM 1
<none>

Cert Suite Configuration: Only in CLAIM 2
/field2/subfield1=58

Cluster abnormal events count
CLAIM 1     CLAIM 2
1           2
`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			diffReport := GetDiffReport(tc.configs1, tc.configs2)
			assert.Equal(t, tc.expectedDiff, diffReport)
			assert.Equal(t, tc.expectedDiffStr, diffReport.String())
		})
	}
}
