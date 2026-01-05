// Copyright (C) 2025-2026 Red Hat, Inc.
package autodiscover

import (
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	fakeClientConfigv1 "github.com/openshift/client-go/config/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFindClusterOperators(t *testing.T) {
	generateClusterOperator := func(name string, availableStatus configv1.ConditionStatus, degradedStatus configv1.ConditionStatus, progressingStatus configv1.ConditionStatus) *configv1.ClusterOperator {
		return &configv1.ClusterOperator{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Status: configv1.ClusterOperatorStatus{
				Conditions: []configv1.ClusterOperatorStatusCondition{
					{
						Type:   configv1.OperatorAvailable,
						Status: availableStatus,
					},
					{
						Type:   configv1.OperatorDegraded,
						Status: degradedStatus,
					},
					{
						Type:   configv1.OperatorProgressing,
						Status: progressingStatus,
					},
				},
			},
		}
	}

	// Generate a test object, store it to the fake client, and then retrieve it.
	testObject := generateClusterOperator("test-cluster-operator", configv1.ConditionTrue, configv1.ConditionFalse, configv1.ConditionFalse)

	client := fakeClientConfigv1.NewClientset(testObject)
	clusterOperators, err := findClusterOperators(client.ConfigV1().ClusterOperators())

	// Assert that the test object was retrieved successfully.
	assert.Nil(t, err)
	assert.Len(t, clusterOperators, 1)
	assert.Equal(t, testObject.Name, clusterOperators[0].Name)
	assert.Equal(t, testObject.Status.Conditions, clusterOperators[0].Status.Conditions)
}
