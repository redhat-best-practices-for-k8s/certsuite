// Copyright (C) 2023-2026 Red Hat, Inc.
package webserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v3"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
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
