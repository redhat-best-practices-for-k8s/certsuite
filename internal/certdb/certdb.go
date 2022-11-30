package certdb

import (
	"fmt"

	"github.com/test-network-function/cnf-certification-test/internal/certdb/offlinecheck"
	"github.com/test-network-function/cnf-certification-test/internal/certdb/onlinecheck"
	"helm.sh/helm/v3/pkg/release"
)

//go:generate moq -out api_moq.go . CertAPIClientFuncs
type ResourceCertificator interface {
	IsContainerCertified(registry, repository, tag, digest string) bool
	IsOperatorCertified(csvName, ocpVersion, channel string) bool
	IsHelmChartCertified(helm *release.Release, ourKubeVersion string) bool
}

func GetCertificator(offlineDBPath string) (ResourceCertificator, error) {
	// use the online certificator by default
	onlineValidator := onlinecheck.NewOnlineValidator()
	if onlineValidator.IsServiceReachable() {
		return onlineValidator, nil
	}

	// use the offline DB for disconnected environments
	err := offlinecheck.LoadCatalogs(offlineDBPath)
	if err != nil {
		return nil, fmt.Errorf("offline DB not available, err: %v", err)
	}

	return offlinecheck.OfflineValidator{}, nil
}
