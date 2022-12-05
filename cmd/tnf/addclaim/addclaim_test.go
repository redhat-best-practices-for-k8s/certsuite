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

func Test_compare2TestCaseResults(t *testing.T) {
	type args struct {
		testcaseResult1 testcase
		testcaseResult2 testcase
	}
	tests := []struct {
		name             string
		args             args
		wantDiffresult   testcase
		wantNotFoundtest []string
	}{
		{
			name: "test1",
			args: args{
				testcaseResult1: testcase{
					{
						Name:   "[It] observability observability-container-logging [common, observability, observability-container-logging]",
						Status: "skipped",
					},
				},
				testcaseResult2: testcase{
					{
						Name:   "[It] observability observability-container-logging [common, observability, observability-container-logging]",
						Status: "failed",
					},
				},
			},
			wantDiffresult: testcase{
				{
					Name:   "[It] observability observability-container-logging [common, observability, observability-container-logging]",
					Status: "skipped",
				},
			},
			wantNotFoundtest: nil,
		},
		{
			name: "test2",
			args: args{
				testcaseResult1: testcase{
					{
						Name:   "[It] observability observability-crd-status [common, observability, observability-crd-status]",
						Status: "skipped",
					},
					{
						Name:   "[It] observability observability-container-logging [common, observability, observability-container-logging]",
						Status: "skipped",
					},
				},
				testcaseResult2: testcase{
					{
						Name:   "[It] observability observability-container-logging [common, observability, observability-container-logging]",
						Status: "failed",
					},
				},
			},
			wantDiffresult: testcase{
				{
					Name:   "[It] observability observability-container-logging [common, observability, observability-container-logging]",
					Status: "skipped",
				},
			},
			wantNotFoundtest: []string{"[It] observability observability-crd-status [common, observability, observability-crd-status]"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDiffresult, gotNotFoundtest := compare2TestCaseResults(tt.args.testcaseResult1, tt.args.testcaseResult2)
			if !reflect.DeepEqual(gotDiffresult, tt.wantDiffresult) {
				t.Errorf("compare2TestCaseResults() gotDiffresult = %v, want %v", gotDiffresult, tt.wantDiffresult)
			}
			if !reflect.DeepEqual(gotNotFoundtest, tt.wantNotFoundtest) {
				t.Errorf("compare2TestCaseResults() gotNotFoundtest = %v, want %v", gotNotFoundtest, tt.wantNotFoundtest)
			}
		})
	}
}

func Test_compare2cnis(t *testing.T) {
	type args struct {
		cniList1 cnistruct
		cniList2 cnistruct
	}
	tests := []struct {
		name              string
		args              args
		wantDiffplugins   cnistruct
		wantNotFoundNames []string
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
			wantDiffplugins:   nil,
			wantNotFoundNames: []string{"crio"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDiffplugins, gotNotFoundNames := compare2cnis(tt.args.cniList1, tt.args.cniList2)
			if !reflect.DeepEqual(gotDiffplugins, tt.wantDiffplugins) {
				t.Errorf("compare2cnis() gotDiffplugins = %v, want %v", gotDiffplugins, tt.wantDiffplugins)
			}
			if !reflect.DeepEqual(gotNotFoundNames, tt.wantNotFoundNames) {
				t.Errorf("compare2cnis() gotNotFoundNames = %v, want %v", gotNotFoundNames, tt.wantNotFoundNames)
			}
		})
	}
}
