package checksadapter

import (
	"context"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/arrayhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/checks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	release "helm.sh/helm/v4/pkg/release/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

func TestDerefSlice(t *testing.T) {
	t.Parallel()

	a, b := 1, 2
	result := arrayhelper.DerefSlice([]*int{&a, nil, &b})
	assert.Equal(t, []int{1, 2}, result)
}

func TestDerefSlice_Empty(t *testing.T) {
	t.Parallel()

	result := arrayhelper.DerefSlice([]*int{})
	assert.Empty(t, result)
}

func TestDerefSlice_AllNil(t *testing.T) {
	t.Parallel()

	result := arrayhelper.DerefSlice([]*string{nil, nil})
	assert.Empty(t, result)
}


func TestConvertNodes(t *testing.T) {
	t.Parallel()

	nodeMap := map[string]provider.Node{
		"worker-1": {
			Data: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "worker-1"},
			},
		},
		"worker-2": {
			Data: nil,
		},
		"worker-3": {
			Data: &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "worker-3"},
			},
		},
	}

	result := convertNodes(nodeMap)
	assert.Len(t, result, 2)

	names := make(map[string]bool)
	for _, n := range result {
		names[n.Name] = true
	}
	assert.True(t, names["worker-1"])
	assert.True(t, names["worker-3"])
}

func TestConvertNodes_Empty(t *testing.T) {
	t.Parallel()

	result := convertNodes(map[string]provider.Node{})
	assert.Empty(t, result)
}

func TestConvertHelmReleases(t *testing.T) {
	t.Parallel()

	env := &provider.TestEnvironment{
		HelmChartReleases: []*release.Release{
			{
				Name:      "my-chart",
				Namespace: "default",
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{Version: "1.2.3"},
				},
			},
			nil,
			{
				Name:      "bad-chart",
				Namespace: "ns2",
				Chart:     nil,
			},
			{
				Name:      "no-metadata",
				Namespace: "ns3",
				Chart:     &chart.Chart{Metadata: nil},
			},
		},
	}

	result := convertHelmReleases(env)
	require.Len(t, result, 1)
	assert.Equal(t, "my-chart", result[0].Name)
	assert.Equal(t, "default", result[0].Namespace)
	assert.Equal(t, "1.2.3", result[0].Version)
}

func TestConvertHelmReleases_Empty(t *testing.T) {
	t.Parallel()

	env := &provider.TestEnvironment{}
	result := convertHelmReleases(env)
	assert.Empty(t, result)
}

func TestBuildPodMultusNetworks(t *testing.T) {
	t.Parallel()

	pods := []*provider.Pod{
		{
			Pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"},
			},
			MultusNetworkInterfaces: map[string]provider.CniNetworkInterface{
				"net-a": {Interface: "eth1", IPs: []string{"10.0.0.1"}},
			},
		},
		{
			Pod: nil,
		},
		{
			Pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: "ns1"},
			},
			MultusNetworkInterfaces: map[string]provider.CniNetworkInterface{},
		},
	}

	result := buildPodMultusNetworks(pods)
	require.Contains(t, result, "ns1/pod1")
	assert.Len(t, result["ns1/pod1"], 1)
	assert.Equal(t, "net-a", result["ns1/pod1"][0].Name)
	assert.Equal(t, "eth1", result["ns1/pod1"][0].InterfaceName)
	assert.Equal(t, []string{"10.0.0.1"}, result["ns1/pod1"][0].IPs)

	assert.NotContains(t, result, "ns1/pod2")
}

func TestBuildCRInstances(t *testing.T) {
	// Setup fake dynamic client
	scheme := runtime.NewScheme()
	
	crd := apiextv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: "test-crds.example.com"},
		Spec: apiextv1.CustomResourceDefinitionSpec{
			Group: "example.com",
			Names: apiextv1.CustomResourceDefinitionNames{Plural: "test-crds"},
			Versions: []apiextv1.CustomResourceDefinitionVersion{
				{Name: "v1", Served: true},
			},
		},
	}

	res := &checks.DiscoveredResources{
		CRDs: []apiextv1.CustomResourceDefinition{crd},
	}

	// buildCRInstances is private, but we can test it through ConvertToDiscoveredResources
	// or just test the logic if we were able to call it.
	// Since we are in the same package, we can call it directly.
	
	// Create a CR instance
	gvr := schema.GroupVersionResource{Group: "example.com", Version: "v1", Resource: "test-crds"}
	fakeDynamicClient := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, map[schema.GroupVersionResource]string{
		gvr: "TestCRDList",
	})
	clientsholder.SetTestK8sDynamicClientsHolder(fakeDynamicClient)

	cr := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "example.com/v1",
			"kind":       "TestCRD",
			"metadata": map[string]interface{}{
				"name":      "cr1",
				"namespace": "ns1",
			},
		},
	}
	_, err := fakeDynamicClient.Resource(gvr).Namespace("ns1").Create(context.TODO(), cr, metav1.CreateOptions{})
	require.NoError(t, err)

	instances := buildCRInstances(res)
	assert.NotNil(t, instances)
	assert.Contains(t, instances, "test-crds.example.com")
	assert.Contains(t, instances["test-crds.example.com"], "ns1")
	assert.Equal(t, []string{"cr1"}, instances["test-crds.example.com"]["ns1"])
}

func TestConvertToDiscoveredResources(t *testing.T) {
	clientsholder.GetTestClientsHolder(nil)
	env := &provider.TestEnvironment{
		Namespaces: []string{"ns1"},
		Pods: []*provider.Pod{
			{Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "ns1"}}},
		},
		Deployments: []*provider.Deployment{
			{Deployment: &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "dep1", Namespace: "ns1"}}},
		},
	}

	res := ConvertToDiscoveredResources(env)
	assert.NotNil(t, res)
	assert.Equal(t, env.Namespaces, res.Namespaces)
	assert.Len(t, res.Pods, 1)
	assert.Equal(t, "pod1", res.Pods[0].Name)
	assert.Len(t, res.Deployments, 1)
	assert.Equal(t, "dep1", res.Deployments[0].Name)
	assert.NotNil(t, res.ProbeExecutor)
}

func TestResetCache(t *testing.T) {
	clientsholder.GetTestClientsHolder(nil)
	env := &provider.TestEnvironment{Namespaces: []string{"ns1"}}
	
	res1 := ConvertToDiscoveredResources(env)
	assert.NotNil(t, res1)
	
	ResetCache()
	
	res2 := ConvertToDiscoveredResources(env)
	assert.NotNil(t, res2)
	assert.NotSame(t, res1, res2)
}


func TestBuildPodMultusNetworks_Empty(t *testing.T) {
	t.Parallel()

	result := buildPodMultusNetworks(nil)
	assert.Empty(t, result)
}
