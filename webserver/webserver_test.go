// Copyright (C) 2023-2026 Red Hat, Inc.
package webserver

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite-claim/pkg/claim"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v3"
)

func TestUpdateTnfManagedDeploymentsAndStatefulsets(t *testing.T) {
	testCases := []struct {
		name                        string
		managedDeployments          []string
		managedStatefulsets         []string
		expectedManagedDeployments  []configuration.ManagedDeploymentsStatefulsets
		expectedManagedStatefulsets []configuration.ManagedDeploymentsStatefulsets
	}{
		{
			name:                "different values are stored independently",
			managedDeployments:  []string{"my-dep-1", "my-dep-2"},
			managedStatefulsets: []string{"my-sts-1", "my-sts-2"},
			expectedManagedDeployments: []configuration.ManagedDeploymentsStatefulsets{
				{Name: "my-dep-1"},
				{Name: "my-dep-2"},
			},
			expectedManagedStatefulsets: []configuration.ManagedDeploymentsStatefulsets{
				{Name: "my-sts-1"},
				{Name: "my-sts-2"},
			},
		},
		{
			name:                        "empty inputs produce nil slices",
			managedDeployments:          []string{},
			managedStatefulsets:         []string{},
			expectedManagedDeployments:  nil,
			expectedManagedStatefulsets: nil,
		},
		{
			name:                "only deployments populated",
			managedDeployments:  []string{"dep-only"},
			managedStatefulsets: []string{},
			expectedManagedDeployments: []configuration.ManagedDeploymentsStatefulsets{
				{Name: "dep-only"},
			},
			expectedManagedStatefulsets: nil,
		},
		{
			name:                       "only statefulsets populated",
			managedDeployments:         []string{},
			managedStatefulsets:        []string{"sts-only"},
			expectedManagedDeployments: nil,
			expectedManagedStatefulsets: []configuration.ManagedDeploymentsStatefulsets{
				{Name: "sts-only"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseConfig := configuration.TestConfiguration{}
			tnfConfig, err := yaml.Marshal(&baseConfig)
			assert.NoError(t, err)

			data := &RequestedData{
				ManagedDeployments:  tc.managedDeployments,
				ManagedStatefulsets: tc.managedStatefulsets,
			}

			resultYAML := updateTnf(tnfConfig, data)

			var result configuration.TestConfiguration
			err = yaml.Unmarshal(resultYAML, &result)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedManagedDeployments, result.ManagedDeployments)
			assert.Equal(t, tc.expectedManagedStatefulsets, result.ManagedStatefulsets)
		})
	}
}

func TestToJSONString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    map[string]string
		expected string
	}{
		{
			name: "valid map",
			input: map[string]string{
				"key1": "value1",
			},
			expected: "{\n  \"key1\": \"value1\"\n}",
		},
		{
			name:     "empty map",
			input:    map[string]string{},
			expected: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := toJSONString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSuitesFromIdentifiers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []claim.Identifier
		expected []string
	}{
		{
			name: "multiple identifiers with duplicate suites",
			input: []claim.Identifier{
				{Id: "test-1", Suite: "networking"},
				{Id: "test-2", Suite: "networking"},
				{Id: "test-3", Suite: "lifecycle"},
			},
			expected: []string{"networking", "lifecycle"},
		},
		{
			name: "single identifier",
			input: []claim.Identifier{
				{Id: "test-1", Suite: "operator"},
			},
			expected: []string{"operator"},
		},
		{
			name:     "empty input",
			input:    []claim.Identifier{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := GetSuitesFromIdentifiers(tt.input)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestCreatePrintableCatalogFromIdentifiers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          []claim.Identifier
		expectedSuites []string
		expectedCounts map[string]int
	}{
		{
			name: "multiple suites",
			input: []claim.Identifier{
				{Id: "net-test-1", Suite: "networking"},
				{Id: "net-test-2", Suite: "networking"},
				{Id: "life-test-1", Suite: "lifecycle"},
			},
			expectedSuites: []string{"networking", "lifecycle"},
			expectedCounts: map[string]int{
				"networking": 2,
				"lifecycle":  1,
			},
		},
		{
			name: "single suite",
			input: []claim.Identifier{
				{Id: "op-test-1", Suite: "operator"},
			},
			expectedSuites: []string{"operator"},
			expectedCounts: map[string]int{
				"operator": 1,
			},
		},
		{
			name:           "empty input",
			input:          []claim.Identifier{},
			expectedSuites: nil,
			expectedCounts: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := CreatePrintableCatalogFromIdentifiers(tt.input)

			require.NotNil(t, result)

			for _, suite := range tt.expectedSuites {
				entries, ok := result[suite]
				assert.True(t, ok)
				assert.Len(t, entries, tt.expectedCounts[suite])
			}

			for suite, count := range tt.expectedCounts {
				assert.Len(t, result[suite], count)
			}

			for _, id := range tt.input {
				entries := result[id.Suite]
				found := false
				for _, e := range entries {
					if e.testName == id.Id && e.identifier == id {
						found = true
						break
					}
				}
				assert.True(t, found)
			}
		})
	}
}
