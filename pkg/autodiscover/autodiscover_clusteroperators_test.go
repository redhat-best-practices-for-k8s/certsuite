package autodiscover

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

func TestFindClusterOperators(t *testing.T) {
	// Test ClusterOperator name
	testName := "test-cluster-operator"

	scheme := runtime.NewScheme()
	gvr := schema.GroupVersionResource{Group: "config.openshift.io", Version: "v1", Resource: "clusteroperators"}

	// Create client with custom List kind mapping to avoid panic on List
	listKinds := map[schema.GroupVersionResource]string{
		gvr: "ClusterOperatorList",
	}
	dynamicClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds)

	// Create unstructured ClusterOperator object
	u := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "config.openshift.io/v1",
		"kind":       "ClusterOperator",
		"metadata": map[string]interface{}{
			"name": testName,
		},
		"status": map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"type":   "Available",
					"status": "True",
				},
				map[string]interface{}{
					"type":   "Degraded",
					"status": "False",
				},
				map[string]interface{}{
					"type":   "Progressing",
					"status": "False",
				},
			},
		},
	}}
	_, _ = dynamicClient.Resource(gvr).Create(context.TODO(), u, metav1.CreateOptions{})

	clusterOperators, err := findClusterOperators(dynamicClient)

	// Assert that the test object was retrieved successfully.
	assert.Nil(t, err)
	assert.Len(t, clusterOperators, 1)
	assert.Equal(t, testName, clusterOperators[0].Name)
	// We do a minimal assertion on conditions presence
	assert.NotEmpty(t, clusterOperators[0].Status.Conditions)
}
