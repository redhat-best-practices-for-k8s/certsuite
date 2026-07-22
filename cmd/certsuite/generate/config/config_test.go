package config

import (
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
	"github.com/stretchr/testify/assert"
)

func TestLoadCRDfilters(t *testing.T) {
	testCases := []struct {
		name            string
		answers         []string
		expectedFilters []configuration.CrdFilter
	}{
		{
			name:    "valid scalable true",
			answers: []string{"myname/true"},
			expectedFilters: []configuration.CrdFilter{
				{NameSuffix: "myname", Scalable: true},
			},
		},
		{
			name:    "valid scalable false",
			answers: []string{"myname/false"},
			expectedFilters: []configuration.CrdFilter{
				{NameSuffix: "myname", Scalable: false},
			},
		},
		{
			name:    "multiple entries",
			answers: []string{"name1/true", "name2/false"},
			expectedFilters: []configuration.CrdFilter{
				{NameSuffix: "name1", Scalable: true},
				{NameSuffix: "name2", Scalable: false},
			},
		},
		{
			name:            "invalid bool value",
			answers:         []string{"myname/notabool"},
			expectedFilters: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			certsuiteConfig = configuration.TestConfiguration{}
			loadCRDfilters(tc.answers)
			assert.Equal(t, tc.expectedFilters, certsuiteConfig.CrdFilters)
		})
	}
}

func TestLoadNonScalableDeployments(t *testing.T) {
	testCases := []struct {
		name       string
		answers    []string
		expectedDS []configuration.SkipScalingTestDeploymentsInfo
	}{
		{
			name:    "valid name/namespace",
			answers: []string{"myname/mynamespace"},
			expectedDS: []configuration.SkipScalingTestDeploymentsInfo{
				{Name: "myname", Namespace: "mynamespace"},
			},
		},
		{
			name:    "multiple entries",
			answers: []string{"deploy1/ns1", "deploy2/ns2"},
			expectedDS: []configuration.SkipScalingTestDeploymentsInfo{
				{Name: "deploy1", Namespace: "ns1"},
				{Name: "deploy2", Namespace: "ns2"},
			},
		},
		{
			name:       "missing separator",
			answers:    []string{"nameonly"},
			expectedDS: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			certsuiteConfig = configuration.TestConfiguration{}
			loadNonScalableDeployments(tc.answers)
			assert.Equal(t, tc.expectedDS, certsuiteConfig.SkipScalingTestDeployments)
		})
	}
}

func TestLoadNonScalableStatefulSets(t *testing.T) {
	testCases := []struct {
		name       string
		answers    []string
		expectedSS []configuration.SkipScalingTestStatefulSetsInfo
	}{
		{
			name:    "valid name/namespace",
			answers: []string{"mysts/mynamespace"},
			expectedSS: []configuration.SkipScalingTestStatefulSetsInfo{
				{Name: "mysts", Namespace: "mynamespace"},
			},
		},
		{
			name:    "multiple entries",
			answers: []string{"sts1/ns1", "sts2/ns2"},
			expectedSS: []configuration.SkipScalingTestStatefulSetsInfo{
				{Name: "sts1", Namespace: "ns1"},
				{Name: "sts2", Namespace: "ns2"},
			},
		},
		{
			name:       "missing separator",
			answers:    []string{"nameonly"},
			expectedSS: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			certsuiteConfig = configuration.TestConfiguration{}
			loadNonScalableStatefulSets(tc.answers)
			assert.Equal(t, tc.expectedSS, certsuiteConfig.SkipScalingTestStatefulSets)
		})
	}
}

func TestLoadNamespaces(t *testing.T) {
	testCases := []struct {
		name               string
		answers            []string
		expectedNamespaces []configuration.Namespace
	}{
		{
			name:               "empty slice",
			answers:            []string{},
			expectedNamespaces: nil,
		},
		{
			name:    "single entry",
			answers: []string{"ns1"},
			expectedNamespaces: []configuration.Namespace{
				{Name: "ns1"},
			},
		},
		{
			name:    "multiple entries",
			answers: []string{"ns1", "ns2", "ns3"},
			expectedNamespaces: []configuration.Namespace{
				{Name: "ns1"},
				{Name: "ns2"},
				{Name: "ns3"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			certsuiteConfig = configuration.TestConfiguration{}
			loadNamespaces(tc.answers)
			assert.Equal(t, tc.expectedNamespaces, certsuiteConfig.TargetNameSpaces)
		})
	}
}

func TestLoadManagedDeployments(t *testing.T) {
	testCases := []struct {
		name     string
		answers  []string
		expected []configuration.ManagedDeploymentsStatefulsets
	}{
		{
			name:     "empty slice",
			answers:  []string{},
			expected: nil,
		},
		{
			name:    "multiple entries",
			answers: []string{"deploy1", "deploy2"},
			expected: []configuration.ManagedDeploymentsStatefulsets{
				{Name: "deploy1"},
				{Name: "deploy2"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			certsuiteConfig = configuration.TestConfiguration{}
			loadManagedDeployments(tc.answers)
			assert.Equal(t, tc.expected, certsuiteConfig.ManagedDeployments)
		})
	}
}

func TestLoadManagedStatefulSets(t *testing.T) {
	testCases := []struct {
		name     string
		answers  []string
		expected []configuration.ManagedDeploymentsStatefulsets
	}{
		{
			name:     "empty slice",
			answers:  []string{},
			expected: nil,
		},
		{
			name:    "multiple entries",
			answers: []string{"sts1", "sts2"},
			expected: []configuration.ManagedDeploymentsStatefulsets{
				{Name: "sts1"},
				{Name: "sts2"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			certsuiteConfig = configuration.TestConfiguration{}
			loadManagedStatefulSets(tc.answers)
			assert.Equal(t, tc.expected, certsuiteConfig.ManagedStatefulsets)
		})
	}
}
