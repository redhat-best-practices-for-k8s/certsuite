package releasehelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/internal/api"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

//nolint:funlen
func TestIsReleaseCertified(t *testing.T) {
	// Create a helm object
	generateHelm := func(name, version string) *release.Release {
		return &release.Release{
			Chart: &chart.Chart{
				Metadata: &chart.Metadata{
					Name:    name,
					Version: version,
				},
			},
		}
	}
	// Create a chart struct
	generateChartStruct := func(name, version, kubeVersion string) api.ChartStruct {
		return api.ChartStruct{
			Entries: map[string][]struct {
				Name        string "yaml:\"name\""
				Version     string "yaml:\"version\""
				KubeVersion string "yaml:\"kubeVersion\""
			}{
				"entry1": {
					{
						Name:        name,
						Version:     version,
						KubeVersion: kubeVersion,
					},
				},
			},
		}
	}

	testCases := []struct {
		testKubeVersion   string
		testRelease       *release.Release
		testChartStruct   api.ChartStruct
		expectedCertified bool
	}{
		{ // Test Case #1 - FAIL the entries mismatched helm1 vs. helm2
			testRelease:       generateHelm("helm1", "0.0.1"),
			testKubeVersion:   "1.18.1",
			testChartStruct:   generateChartStruct("helm2", "0.0.1", ">= 1.19"),
			expectedCertified: false,
		},
		{ // Test Case #2 - PASS the entries matched
			testRelease:       generateHelm("helm1", "0.0.1"),
			testKubeVersion:   "1.20.1",
			testChartStruct:   generateChartStruct("helm1", "0.0.1", ">= 1.19"),
			expectedCertified: true,
		},
		{ // Test Case #3 - FAIL the versions mismatch 0.0.1 vs 0.0.2
			testRelease:       generateHelm("helm1", "0.0.1"),
			testKubeVersion:   "1.18.1",
			testChartStruct:   generateChartStruct("helm1", "0.0.2", ">= 1.19"),
			expectedCertified: false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedCertified, IsReleaseCertified(tc.testRelease, tc.testKubeVersion, tc.testChartStruct))
	}
}
