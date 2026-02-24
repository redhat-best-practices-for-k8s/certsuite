// Copyright (C) 2020-2026 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package certification

import (
	"strings"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/stretchr/testify/assert"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	corev1 "k8s.io/api/core/v1"
)

func TestGetContainersToQuery(t *testing.T) {
	var testEnv = provider.TestEnvironment{
		Containers: []*provider.Container{
			{
				ContainerImageIdentifier: provider.ContainerImageIdentifier{
					Repository: "test1",
					Registry:   "repo1",
				},
			},
			{
				ContainerImageIdentifier: provider.ContainerImageIdentifier{
					Repository: "test2",
					Registry:   "repo2",
				},
			},
		},
	}

	testCases := []struct {
		expectedOutput map[provider.ContainerImageIdentifier]bool
	}{
		{
			expectedOutput: map[provider.ContainerImageIdentifier]bool{
				{
					Repository: "test1",
					Registry:   "repo1",
				}: true,
				{
					Repository: "test2",
					Registry:   "repo2",
				}: true,
			},
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, getContainersToQuery(&testEnv))
	}
}

// mockValidator implements certdb.CertificationStatusValidator for testing
type mockValidator struct {
	containerCertified map[string]bool
	operatorCertified  map[string]bool
	helmCertified      map[string]bool
}

func (m *mockValidator) IsContainerCertified(registry, repository, tag, digest string) bool {
	key := registry + "/" + repository + ":" + tag + "@" + digest
	return m.containerCertified[key]
}

func (m *mockValidator) IsOperatorCertified(csvName, ocpVersion string) bool {
	key := csvName + "@" + ocpVersion
	return m.operatorCertified[key]
}

func (m *mockValidator) IsHelmChartCertified(helm *release.Release, kubeVersion string) bool {
	return m.helmCertified[helm.Name]
}

func setupCertCheck() *checksdb.Check {
	var logArchive strings.Builder
	log.SetupLogger(&logArchive, "INFO")
	return checksdb.NewCheck("test-cert-id", nil)
}

// ---- TestContainerCertificationStatusByDigest ----

func TestContainerCertificationStatusByDigest_EmptyDigest(t *testing.T) {
	check := setupCertCheck()
	testEnv := &provider.TestEnvironment{
		Containers: []*provider.Container{
			{
				Container: &corev1.Container{Name: "c1"},
				ContainerImageIdentifier: provider.ContainerImageIdentifier{
					Registry:   "registry.io",
					Repository: "myimage",
					Tag:        "latest",
					Digest:     "",
				},
				Namespace: "ns1",
				Podname:   "pod1",
			},
		},
	}

	validator := &mockValidator{containerCertified: map[string]bool{}}

	testContainerCertificationStatusByDigest(check, testEnv, validator)
	assert.Equal(t, "failed", string(check.Result))
}

func TestContainerCertificationStatusByDigest_Certified(t *testing.T) {
	check := setupCertCheck()
	testEnv := &provider.TestEnvironment{
		Containers: []*provider.Container{
			{
				Container: &corev1.Container{Name: "c1"},
				ContainerImageIdentifier: provider.ContainerImageIdentifier{
					Registry:   "registry.io",
					Repository: "myimage",
					Tag:        "latest",
					Digest:     "sha256:abc123",
				},
				Namespace: "ns1",
				Podname:   "pod1",
			},
		},
	}

	validator := &mockValidator{
		containerCertified: map[string]bool{
			"registry.io/myimage:latest@sha256:abc123": true,
		},
	}

	testContainerCertificationStatusByDigest(check, testEnv, validator)
	assert.Equal(t, "passed", string(check.Result))
}

func TestContainerCertificationStatusByDigest_NotCertified(t *testing.T) {
	check := setupCertCheck()
	testEnv := &provider.TestEnvironment{
		Containers: []*provider.Container{
			{
				Container: &corev1.Container{Name: "c1"},
				ContainerImageIdentifier: provider.ContainerImageIdentifier{
					Registry:   "registry.io",
					Repository: "myimage",
					Tag:        "latest",
					Digest:     "sha256:abc123",
				},
				Namespace: "ns1",
				Podname:   "pod1",
			},
		},
	}

	validator := &mockValidator{containerCertified: map[string]bool{}}

	testContainerCertificationStatusByDigest(check, testEnv, validator)
	assert.Equal(t, "failed", string(check.Result))
}

// ---- TestAllOperatorCertified ----

func TestAllOperatorCertified_AllCertified(t *testing.T) {
	check := setupCertCheck()
	testEnv := &provider.TestEnvironment{
		OpenshiftVersion: "4.13.0",
		Operators: []*provider.Operator{
			{Name: "operator1", Namespace: "ns1", Channel: "stable"},
			{Name: "operator2", Namespace: "ns2", Channel: "stable"},
		},
	}

	validator := &mockValidator{
		operatorCertified: map[string]bool{
			"operator1@4.13": true,
			"operator2@4.13": true,
		},
	}

	testAllOperatorCertified(check, testEnv, validator)
	assert.Equal(t, "passed", string(check.Result))
}

func TestAllOperatorCertified_MixedCertification(t *testing.T) {
	check := setupCertCheck()
	testEnv := &provider.TestEnvironment{
		OpenshiftVersion: "4.13.0",
		Operators: []*provider.Operator{
			{Name: "operator1", Namespace: "ns1", Channel: "stable"},
			{Name: "operator2", Namespace: "ns2", Channel: "stable"},
		},
	}

	validator := &mockValidator{
		operatorCertified: map[string]bool{
			"operator1@4.13": true,
			"operator2@4.13": false,
		},
	}

	testAllOperatorCertified(check, testEnv, validator)
	assert.Equal(t, "failed", string(check.Result))
}

func TestAllOperatorCertified_NoneCertified(t *testing.T) {
	check := setupCertCheck()
	testEnv := &provider.TestEnvironment{
		OpenshiftVersion: "4.13.0",
		Operators: []*provider.Operator{
			{Name: "operator1", Namespace: "ns1", Channel: "stable"},
		},
	}

	validator := &mockValidator{
		operatorCertified: map[string]bool{},
	}

	testAllOperatorCertified(check, testEnv, validator)
	assert.Equal(t, "failed", string(check.Result))
}

// ---- TestHelmCertified ----

func TestHelmCertified_AllCertified(t *testing.T) {
	check := setupCertCheck()
	testEnv := &provider.TestEnvironment{
		HelmChartReleases: []*release.Release{
			{
				Name:      "mychart",
				Namespace: "ns1",
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{Version: "1.0.0"},
				},
			},
		},
	}

	validator := &mockValidator{
		helmCertified: map[string]bool{
			"mychart": true,
		},
	}

	testHelmCertified(check, testEnv, validator)
	assert.Equal(t, "passed", string(check.Result))
}

func TestHelmCertified_NotCertified(t *testing.T) {
	check := setupCertCheck()
	testEnv := &provider.TestEnvironment{
		HelmChartReleases: []*release.Release{
			{
				Name:      "mychart",
				Namespace: "ns1",
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{Version: "1.0.0"},
				},
			},
		},
	}

	validator := &mockValidator{
		helmCertified: map[string]bool{},
	}

	testHelmCertified(check, testEnv, validator)
	assert.Equal(t, "failed", string(check.Result))
}

// ---- TestContainerCertification helper ----

func TestContainerCertification_Certified(t *testing.T) {
	validator := &mockValidator{
		containerCertified: map[string]bool{
			"registry.io/myimage:latest@sha256:abc123": true,
		},
	}
	cii := provider.ContainerImageIdentifier{
		Registry:   "registry.io",
		Repository: "myimage",
		Tag:        "latest",
		Digest:     "sha256:abc123",
	}
	assert.True(t, testContainerCertification(cii, validator))
}

func TestContainerCertification_NotCertified(t *testing.T) {
	validator := &mockValidator{
		containerCertified: map[string]bool{},
	}
	cii := provider.ContainerImageIdentifier{
		Registry:   "registry.io",
		Repository: "myimage",
		Tag:        "latest",
		Digest:     "sha256:abc123",
	}
	assert.False(t, testContainerCertification(cii, validator))
}
