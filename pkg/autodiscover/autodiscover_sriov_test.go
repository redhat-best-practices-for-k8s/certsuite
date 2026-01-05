// Copyright (C) 2023-2026 Red Hat, Inc.
package autodiscover

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/stretchr/testify/assert"
)

func TestGetSriovNetworks_NilClient(t *testing.T) {
	// Test with nil client - should not panic
	result, err := getSriovNetworks(nil, []string{"default"})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetSriovNetworkNodePolicies_NilClient(t *testing.T) {
	// Test with nil client - should not panic
	result, err := getSriovNetworkNodePolicies(nil, []string{"default"})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetSriovNetworks_NilDynamicClient(t *testing.T) {
	// Test with client that has nil DynamicClient - should not panic
	client := &clientsholder.ClientsHolder{
		DynamicClient: nil,
	}
	result, err := getSriovNetworks(client, []string{"default"})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetSriovNetworkNodePolicies_NilDynamicClient(t *testing.T) {
	// Test with client that has nil DynamicClient - should not panic
	client := &clientsholder.ClientsHolder{
		DynamicClient: nil,
	}
	result, err := getSriovNetworkNodePolicies(client, []string{"default"})
	assert.NoError(t, err)
	assert.Empty(t, result)
}
