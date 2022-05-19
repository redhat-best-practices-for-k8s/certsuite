// Copyright (C) 2020-2022 Red Hat, Inc.
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
			Entries: map[string][]api.ChartEntry{
				"entry1": []api.ChartEntry{
					{Name: name, Version: version, KubeVersion: kubeVersion},
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
		assert.Equal(t, tc.expectedCertified, IsReleaseCertified(tc.testRelease, tc.testKubeVersion, tc.testChartStruct.Entries))
	}
}
