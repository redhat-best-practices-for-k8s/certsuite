package releasehelper

import (
	"github.com/test-network-function/cnf-certification-test/internal/registry"
	"github.com/test-network-function/cnf-certification-test/pkg/stringhelper"
	"helm.sh/helm/v3/pkg/release"
)

func IsReleaseCertified(helm *release.Release, ourKubeVersion string, chartsdb map[string][]registry.ChartEntry) bool {
	for _, entryList := range chartsdb {
		for _, entry := range entryList {
			if entry.Name == helm.Chart.Metadata.Name && entry.Version == helm.Chart.Metadata.Version {
				if entry.KubeVersion != "" {
					if stringhelper.CompareVersion(ourKubeVersion, entry.KubeVersion) {
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
