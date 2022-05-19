package registry

import (
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
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
