package checksadapter

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDerefSlice(t *testing.T) {
	t.Parallel()

	a, b := 1, 2
	result := derefSlice([]*int{&a, nil, &b})
	assert.Equal(t, []int{1, 2}, result)
}

func TestDerefSlice_Empty(t *testing.T) {
	t.Parallel()

	result := derefSlice([]*int{})
	assert.Empty(t, result)
}

func TestDerefSlice_AllNil(t *testing.T) {
	t.Parallel()

	result := derefSlice([]*string{nil, nil})
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

func TestBuildPodMultusNetworks_Empty(t *testing.T) {
	t.Parallel()

	result := buildPodMultusNetworks(nil)
	assert.Empty(t, result)
}
