// Copyright (C) 2020-2023 Red Hat, Inc.
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
package offlinecheck

import (
	"fmt"
	"io"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/release"
)

const (
	helmRelativePath = "%s/data/helm/helm.db"
)

type ChartEntry struct {
	Name                  string `yaml:"name"`
	ChartVersion          string `yaml:"version"`
	KubeVersionConstraint string `yaml:"kubeVersion"`
}
type ChartStruct struct {
	Entries map[string][]ChartEntry `yaml:"entries"`
}

var chartsdb = make(map[string][]ChartEntry)
var loaded = false

func loadHelmCatalog(offlineDBPath string) error {
	if loaded {
		return nil
	}
	loaded = true
	filePath := fmt.Sprintf(helmRelativePath, offlineDBPath)
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("cannot process file %s, err: %v", filePath, err)
	}
	defer f.Close()
	bytes, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("cannot process file %s, err: %v", filePath, err)
	}
	var charts ChartStruct
	if err = yaml.Unmarshal(bytes, &charts); err != nil {
		return fmt.Errorf("cannot parse the yaml file of the helm certification list, err: %v", err)
	}
	chartsdb = charts.Entries

	return nil
}

func LoadHelmCharts(charts ChartStruct) {
	chartsdb = map[string][]ChartEntry{}
	chartsdb = charts.Entries
}

// CompareVersion compare between versions
func CompareVersion(version, constraint string) bool {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		logrus.Errorf("cannot parse semver constraint string=%s, err=%s", constraint, err)
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		// Handle version not being parsable.
		logrus.Errorf("cannot parse semver version, string=%s err=%s", version, err)
	}
	// Check if the version meets the constraints. The a variable will be true.
	return c.Check(v)
}

func (validator OfflineValidator) IsHelmChartCertified(helm *release.Release, ourKubeVersion string) bool {
	for _, entryList := range chartsdb {
		for _, entry := range entryList {
			if entry.Name == helm.Chart.Metadata.Name && entry.ChartVersion == helm.Chart.Metadata.Version {
				if entry.KubeVersionConstraint != "" {
					if CompareVersion(ourKubeVersion, entry.KubeVersionConstraint) {
						return true
					}
				} else {
					return true
				}
			}
		}
	}
	return false
}
