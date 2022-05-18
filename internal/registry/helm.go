package registry

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
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

func IsReleaseCertified(helm *release.Release, ourKubeVersion string) bool {
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
