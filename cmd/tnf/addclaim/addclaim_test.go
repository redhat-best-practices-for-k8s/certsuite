package claim

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/test-network-function-claim/pkg/claim"
)

func TestReadClaim(t *testing.T) {
	testCases := []struct {
		testContents      string
		expectedClaimRoot *claim.Root
	}{
		{ // Test Case #1 - Happy path
			testContents: `{"claim":{"versions":{"k8s":"1.23.1","tnf":"0.3.1"},"configurations":{},"metadata":{"endTime":"1:33:00","startTime":"2:33:00"},"nodes":{},"results":{},"rawResults":{}}}`,
			expectedClaimRoot: &claim.Root{
				Claim: &claim.Claim{
					Versions: &claim.Versions{
						K8s: "1.23.1",
						Tnf: "0.3.1",
					},
					Metadata: &claim.Metadata{
						EndTime:   "1:33:00",
						StartTime: "2:33:00",
					},
					Nodes:          make(map[string]interface{}),
					Results:        make(map[string]interface{}),
					RawResults:     make(map[string]interface{}),
					Configurations: make(map[string]interface{}),
				},
			},
		},
		// Test Case 2 - Cannot test a failure to unmarshal because readClaim logs fatal
	}

	for _, tc := range testCases {
		byteContents := []byte(tc.testContents)
		assert.Equal(t, tc.expectedClaimRoot, readClaim(&byteContents))
	}
}

func TestNewCommand(t *testing.T) {
	// No parameters to test
	result := NewCommand()
	assert.NotNil(t, result)
	assert.Equal(t, "claim", result.Use)
	assert.Equal(t, "The test suite generates a \"claim\" file", result.Short)
}

//nolint:funlen
func Test_compare2TestCaseResults(t *testing.T) {
	type args struct {
		testcaseResult1 []testCase
		testcaseResult2 []testCase
	}
	tests := []struct {
		name              string
		args              args
		wantDiffresult    []testCase
		wantNotFoundtest  []string
		wantNotFoundtest2 []string
	}{
		{
			name: "test1",
			args: args{
				testcaseResult1: []testCase{
					{
						Name:   "[It] observability observability-container-logging [common, observability, observability-container-logging]",
						Status: "skipped",
					},
				},
				testcaseResult2: []testCase{
					{
						Name:   "[It] observability observability-container-logging [common, observability, observability-container-logging]",
						Status: "failed",
					},
				},
			},
			wantDiffresult: []testCase{
				{
					Name:   "[It] observability observability-container-logging [common, observability, observability-container-logging]",
					Status: "skipped",
				},
			},
			wantNotFoundtest:  []string{},
			wantNotFoundtest2: []string{},
		},
		{
			name: "test2",
			args: args{
				testcaseResult1: []testCase{
					{
						Name:   "[It] observability observability-crd-status [common, observability, observability-crd-status]",
						Status: "skipped",
					},
					{
						Name:   "[It] observability observability-container-logging [common, observability, observability-container-logging]",
						Status: "skipped",
					},
				},
				testcaseResult2: []testCase{
					{
						Name:   "[It] observability observability-container-logging [common, observability, observability-container-logging]",
						Status: "failed",
					},
				},
			},
			wantDiffresult: []testCase{
				{
					Name:   "[It] observability observability-container-logging [common, observability, observability-container-logging]",
					Status: "skipped",
				},
			},
			wantNotFoundtest:  []string{},
			wantNotFoundtest2: []string{"[It] observability observability-crd-status [common, observability, observability-crd-status]"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDiffresult, gotNotFoundtest, gotNotFoundtest2 := compare2TestCaseResults(tt.args.testcaseResult1, tt.args.testcaseResult2)
			if !reflect.DeepEqual(gotDiffresult, tt.wantDiffresult) {
				t.Errorf("compare2TestCaseResults() gotDiffresult = %v, want %v", gotDiffresult, tt.wantDiffresult)
			}
			if !reflect.DeepEqual(gotNotFoundtest, tt.wantNotFoundtest) {
				t.Errorf("compare2TestCaseResults() gotNotFoundtest = %v, want %v", gotNotFoundtest, tt.wantNotFoundtest)
			}
			if !reflect.DeepEqual(gotNotFoundtest2, tt.wantNotFoundtest2) {
				t.Errorf("compare2TestCaseResults() gotNotFoundtest2 = %v, want %v", gotNotFoundtest2, tt.wantNotFoundtest2)
			}
		})
	}
}

//nolint:funlen
func Test_compare2cnis(t *testing.T) {
	type args struct {
		cniList1 cnistruct
		cniList2 cnistruct
	}
	tests := []struct {
		name               string
		args               args
		wantDiffplugins    cnistruct
		wantNotFoundNames  []string
		wantNotFoundNames2 []string
	}{
		{
			name: "test1",
			args: args{
				cniList1: cnistruct{
					{
						Name:    "podman",
						Plugins: nil,
					},
					{
						Name:    "crio",
						Plugins: nil,
					},
				},
				cniList2: cnistruct{
					{
						Name:    "podman",
						Plugins: nil,
					},
				},
			},
			wantDiffplugins:    nil,
			wantNotFoundNames:  []string{"crio"},
			wantNotFoundNames2: []string{},
		},
		{
			name: "test2",
			args: args{
				cniList1: cnistruct{
					{
						Name:    "podman",
						Plugins: nil,
					},
				},
				cniList2: cnistruct{
					{
						Name:    "podman",
						Plugins: nil,
					},
					{
						Name:    "crio",
						Plugins: nil,
					},
				},
			},
			wantDiffplugins:    nil,
			wantNotFoundNames:  []string{},
			wantNotFoundNames2: []string{"crio"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDiffplugins, gotNotFoundNames, gotNotFoundNames2 := compare2cniHelper(tt.args.cniList1, tt.args.cniList2)
			if !reflect.DeepEqual(gotDiffplugins, tt.wantDiffplugins) {
				t.Errorf("compare2cnis() gotDiffplugins = %v, want %v", gotDiffplugins, tt.wantDiffplugins)
			}
			if !reflect.DeepEqual(gotNotFoundNames, tt.wantNotFoundNames) {
				t.Errorf("compare2cnis() gotNotFoundNames = %v, want %v", gotNotFoundNames, tt.wantNotFoundNames)
			}
			if !reflect.DeepEqual(gotNotFoundNames2, tt.wantNotFoundNames2) {
				t.Errorf("compare2cnis() gotNotFoundNames2 = %v, want %v", gotNotFoundNames2, tt.wantNotFoundNames2)
			}
		})
	}
}

func Test_claimCompareFilesfunc(t *testing.T) {
	type args struct {
		claim1 string
		claim2 string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{{

		name: "test1",
		args: args{
			claim1: "claim_test1.json",
			claim2: "claim_test2.json",
		},
		wantErr: false,
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := claimCompareFilesfunc(tt.args.claim1, tt.args.claim2); (err != nil) != tt.wantErr {
				t.Errorf("claimCompareFilesfunc() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
