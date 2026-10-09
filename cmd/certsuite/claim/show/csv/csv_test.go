package csv

import (
	"testing"

	claimschema "github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/pkg/claim"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCatalogByID(t *testing.T) {
	catalogMap := buildCatalogByID()
	assert.NotNil(t, catalogMap)

	t.Run("catalog map is populated from identifiers.Catalog", func(t *testing.T) {
		assert.GreaterOrEqual(t, len(catalogMap), 1)
	})

	t.Run("entries are keyed by ID string", func(t *testing.T) {
		for id, desc := range catalogMap {
			assert.NotEmpty(t, id)
			assert.NotEmpty(t, desc.Identifier.Id)
			assert.Equal(t, id, desc.Identifier.Id)
		}
	})
}

func TestBuildCSVReasonColumn(t *testing.T) {
	const reasonColumn = 8

	tests := []struct {
		name           string
		result         claim.TestCaseResult
		expectedReason string
	}{
		{
			name:           "skipped result uses skip reason",
			result:         claim.TestCaseResult{State: claim.TestCaseResultSkipped, SkipReason: "no matching labels"},
			expectedReason: "no matching labels",
		},
		{
			name:           "errored result falls back to error reason",
			result:         claim.TestCaseResult{State: "error", ErrorType: "check-panic", ErrorReason: "panic in check"},
			expectedReason: "panic in check",
		},
		{
			name:           "legacy errored result keeps skip reason",
			result:         claim.TestCaseResult{State: "error", SkipReason: "probe exec failure"},
			expectedReason: "probe exec failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := &claim.Schema{}
			schema.Claim.Results = claim.TestSuiteResults{"test-case": tt.result}

			records := buildCSV(schema, "", map[string]claimschema.TestCaseDescription{})
			require.Len(t, records, 1)
			assert.Equal(t, tt.expectedReason, records[0][reasonColumn])
		})
	}
}
