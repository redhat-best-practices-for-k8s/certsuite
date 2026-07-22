package info

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
	"github.com/stretchr/testify/assert"
)

func TestGetTestDescriptionsFromTestIDs(t *testing.T) {
	t.Parallel()

	var knownID string
	for id := range identifiers.Catalog {
		knownID = id.Id
		break
	}

	testCases := []struct {
		name     string
		testIDs  []string
		validate func(t *testing.T, results []claim.TestCaseDescription)
	}{
		{
			name:    "empty input",
			testIDs: []string{},
			validate: func(t *testing.T, results []claim.TestCaseDescription) {
				assert.Nil(t, results)
			},
		},
		{
			name:    "IDs not in catalog",
			testIDs: []string{"nonexistent-test-id-12345", "another-fake-id"},
			validate: func(t *testing.T, results []claim.TestCaseDescription) {
				assert.Nil(t, results)
			},
		},
		{
			name:    "valid ID from catalog",
			testIDs: []string{knownID},
			validate: func(t *testing.T, results []claim.TestCaseDescription) {
				assert.Len(t, results, 1)
				assert.Equal(t, knownID, results[0].Identifier.Id)
			},
		},
		{
			name:    "mix of valid and invalid IDs",
			testIDs: []string{knownID, "nonexistent-test-id"},
			validate: func(t *testing.T, results []claim.TestCaseDescription) {
				assert.Len(t, results, 1)
				assert.Equal(t, knownID, results[0].Identifier.Id)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			results := getTestDescriptionsFromTestIDs(tc.testIDs)
			tc.validate(t, results)
		})
	}
}
