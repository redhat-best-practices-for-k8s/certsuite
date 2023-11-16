package add

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/claim"
)

func TestReadClaim(t *testing.T) {
	testCases := []struct {
		testContents      string
		expectedClaimRoot *claim.Root
	}{
		{ // Test Case #1 - Happy path
			testContents: `{"claim":{"versions":{"k8s":"1.23.1","tnf":"0.3.1", "claimFormat":"v0.0.1"},"configurations":{},"metadata":{"endTime":"1:33:00","startTime":"2:33:00"},"nodes":{},"results":{},"rawResults":{}}}`,
			expectedClaimRoot: &claim.Root{
				Claim: &claim.Claim{
					Versions: &claim.Versions{
						K8s:         "1.23.1",
						Tnf:         "0.3.1",
						ClaimFormat: "v0.0.1",
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
	assert.Equal(t, "add", result.Use)
	assert.Equal(t, "Add results from xml junit files to an existing claim file.", result.Short)
}
