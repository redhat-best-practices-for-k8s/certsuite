package versions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/cmd/tnf/claim/compare/diff"
	officialClaimScheme "github.com/test-network-function/test-network-function-claim/pkg/claim"
)

func TestCompare(t *testing.T) {
	type args struct {
		claim1Versions *officialClaimScheme.Versions
		claim2Versions *officialClaimScheme.Versions
	}
	tests := []struct {
		name string
		args args
		want *DiffReport
	}{
		{
			name: "empty versions",
			args: args{
				claim1Versions: &officialClaimScheme.Versions{},
				claim2Versions: &officialClaimScheme.Versions{},
			},
			want: &DiffReport{Diffs: &diff.Diffs{Name: "VERSIONS"}},
		},
		{
			name: "matching versions",
			args: args{
				claim1Versions: &officialClaimScheme.Versions{
					ClaimFormat:  "1",
					K8s:          "22",
					OcClient:     "333",
					Ocp:          "4444",
					Tnf:          "55555",
					TnfGitCommit: "666666",
				},
				claim2Versions: &officialClaimScheme.Versions{
					ClaimFormat:  "1",
					K8s:          "22",
					OcClient:     "333",
					Ocp:          "4444",
					Tnf:          "55555",
					TnfGitCommit: "666666",
				},
			},
			want: &DiffReport{Diffs: &diff.Diffs{Name: "VERSIONS"}},
		},
		{
			name: "non matching versions 1",
			args: args{
				claim1Versions: &officialClaimScheme.Versions{
					ClaimFormat:  "1",
					K8s:          "22AAA",
					OcClient:     "333",
					Ocp:          "4444",
					Tnf:          "55555",
					TnfGitCommit: "666666",
				},
				claim2Versions: &officialClaimScheme.Versions{
					ClaimFormat:  "1",
					K8s:          "22",
					OcClient:     "333",
					Ocp:          "4444",
					Tnf:          "55555",
					TnfGitCommit: "666666",
				},
			},
			want: &DiffReport{Diffs: &diff.Diffs{
				Name: "VERSIONS",
				Fields: []diff.FieldDiff{
					{
						FieldPath:   "/k8s",
						Claim1Value: "22AAA",
						Claim2Value: "22",
					},
				},
			}},
		},
		{
			name: "non matching versions 2",
			args: args{
				claim1Versions: &officialClaimScheme.Versions{
					ClaimFormat:  "1",
					K8s:          "22AAA",
					OcClient:     "333",
					Ocp:          "4444",
					Tnf:          "55555",
					TnfGitCommit: "666666",
				},
				claim2Versions: &officialClaimScheme.Versions{
					ClaimFormat:  "1",
					K8s:          "22",
					OcClient:     "333",
					Ocp:          "4444",
					Tnf:          "55555BBBB",
					TnfGitCommit: "666666",
				},
			},
			want: &DiffReport{Diffs: &diff.Diffs{
				Name: "VERSIONS",
				Fields: []diff.FieldDiff{
					{
						FieldPath:   "/k8s",
						Claim1Value: "22AAA",
						Claim2Value: "22",
					},
					{
						FieldPath:   "/tnf",
						Claim1Value: "55555",
						Claim2Value: "55555BBBB",
					},
				},
			}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Compare(tc.args.claim1Versions, tc.args.claim2Versions)
			assert.Equal(t, tc.want, got)
		})
	}
}
