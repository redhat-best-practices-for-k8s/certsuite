package checksadapter

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/checks"
	"github.com/stretchr/testify/assert"
)

func TestCheckInfoToClaimIdentifier(t *testing.T) {
	t.Parallel()

	info := &checks.CheckInfo{
		Name:     "check-x",
		Category: "networking",
		Tags:     []string{"common", "telco"},
	}
	id := CheckInfoToClaimIdentifier(info)
	assert.Equal(t, "check-x", id.Id)
	assert.Equal(t, "networking", id.Suite)
	assert.Equal(t, "common,telco", id.Tags)
}

func TestCheckInfoToClaimIdentifier_SingleTag(t *testing.T) {
	t.Parallel()

	info := &checks.CheckInfo{
		Name:     "simple-check",
		Category: "lifecycle",
		Tags:     []string{"common"},
	}
	id := CheckInfoToClaimIdentifier(info)
	assert.Equal(t, "common", id.Tags)
}

func TestCheckInfoToTestCaseDescription(t *testing.T) {
	t.Parallel()

	info := &checks.CheckInfo{
		Name:                  "test-check",
		Category:              "accesscontrol",
		Description:           "Tests something important",
		Remediation:           "Fix the thing",
		BestPracticeReference: "https://example.com/docs",
		ExceptionProcess:      "No exceptions",
		Tags:                  []string{"common", "telco"},
		Qe:                    true,
		CategoryClassification: map[string]string{
			"FarEdge": "Mandatory",
			"Telco":   "Mandatory",
		},
	}

	desc := CheckInfoToTestCaseDescription(info)
	assert.Equal(t, "test-check", desc.Identifier.Id)
	assert.Equal(t, "accesscontrol", desc.Identifier.Suite)
	assert.Equal(t, "Tests something important", desc.Description)
	assert.Equal(t, "Fix the thing", desc.Remediation)
	assert.Equal(t, "https://example.com/docs", desc.BestPracticeReference)
	assert.Equal(t, "No exceptions", desc.ExceptionProcess)
	assert.Equal(t, "common,telco", desc.Tags)
	assert.True(t, desc.Qe)
	assert.Equal(t, "Mandatory", desc.CategoryClassification["Telco"])
}

func TestGetCheckIDAndLabels_Found(t *testing.T) {
	t.Parallel()

	allChecks := checks.All()
	if len(allChecks) == 0 {
		t.Skip("no checks registered in checks library")
	}

	firstCheck := allChecks[0]

	id, tags := GetCheckIDAndLabels(firstCheck.Name)
	assert.Equal(t, firstCheck.Name, id)
	assert.Contains(t, tags, firstCheck.Name)
	assert.Contains(t, tags, firstCheck.Category)
	for _, tag := range firstCheck.Tags {
		assert.Contains(t, tags, tag)
	}
}

func TestGetCheckIDAndLabels_NotFound(t *testing.T) {
	t.Parallel()

	id, tags := GetCheckIDAndLabels("nonexistent-check-name")
	assert.Equal(t, "nonexistent-check-name", id)
	assert.Equal(t, []string{"nonexistent-check-name"}, tags)
}
