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
package offlinecheck

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/release"
)

const (
	helmRelativePath = "%s/../cmd/tnf/fetch/data/helm/helm.db"
)

type ChartEntry struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	KubeVersion string `yaml:"kubeVersion"`
}
type ChartStruct struct {
	Entries map[string][]ChartEntry `yaml:"entries"`
}

var chartsdb = make(map[string][]ChartEntry)
var loaded = false

func loadHelmCatalog(pathToRoot string) {
	if loaded {
		return
	}
	loaded = true
	filePath := fmt.Sprintf(helmRelativePath, pathToRoot)
	f, err := os.Open(filePath)
	if err != nil {
		log.Error("can't process file", f.Name(), err, " trying to proceed")
		return
	}
	defer f.Close()
	bytes, err := io.ReadAll(f)
	if err != nil {
		log.Error("can't process file", f.Name(), err, " trying to proceed")
	}
	var charts ChartStruct
	if err = yaml.Unmarshal(bytes, &charts); err != nil {
		log.Error("error while parsing the yaml file of the helm certification list ", err)
	}
	chartsdb = charts.Entries
}

func LoadHelmCharts(charts ChartStruct) {
	chartsdb = map[string][]ChartEntry{}
	chartsdb = charts.Entries
}

// CompareVersion compare between versions
func CompareVersion(ver1, ver2 string) bool {
	ourKubeVersion, _ := version.NewVersion(ver1)
	kubeVersion := strings.ReplaceAll(ver2, " ", "")[2:]
	if strings.Contains(kubeVersion, "<") {
		kubever := strings.Split(kubeVersion, "<")
		minVersion, _ := version.NewVersion(kubever[0])
		maxVersion, _ := version.NewVersion(kubever[1])
		if ourKubeVersion.GreaterThanOrEqual(minVersion) && ourKubeVersion.LessThan(maxVersion) {
			return true
		}
	} else {
		kubever := strings.Split(kubeVersion, "-")
		minVersion, _ := version.NewVersion(kubever[0])
		if ourKubeVersion.GreaterThanOrEqual(minVersion) {
			return true
		}
	}
	return false
}

func (checker OfflineChecker) IsReleaseCertified(helm *release.Release, ourKubeVersion string) bool {
	for _, entryList := range chartsdb {
		for _, entry := range entryList {
			if entry.Name == helm.Chart.Metadata.Name && entry.Version == helm.Chart.Metadata.Version {
				if entry.KubeVersion != "" {
					if CompareVersion(ourKubeVersion, entry.KubeVersion) {
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
