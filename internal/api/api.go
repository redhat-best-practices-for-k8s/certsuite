package api

import (
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/api/offlinecheck"
	"github.com/test-network-function/cnf-certification-test/internal/api/onlinecheck"
	"helm.sh/helm/v3/pkg/release"
)

//go:generate moq -out api_moq.go . CertAPIClientFuncs
type CertificationValidator interface {
	IsContainerCertified(registry, repository, tag, digest string) bool
	IsOperatorCertified(csvName, ocpVersion, channel string) bool
	IsReleaseCertified(helm *release.Release, ourKubeVersion string) bool
	IsServiceReachable() bool
}

var onlineClient CertificationValidator
var offlineClient CertificationValidator

var offlineDBLoaded bool

func init() {
	logrus.Info("init certification client")
	onlineClient = onlinecheck.NewOnlineValidator()
	offlineClient = offlinecheck.OfflineChecker{}
}
func LoadCatalog() error {
	err := offlinecheck.LoadCatalogs()
	if err != nil {
		return err
	}
	offlineDBLoaded = true

	return nil
}

func IsContainerCertified(registry, repository, tag, digest string) bool {
	switch {
	case onlineClient.IsServiceReachable():
		return onlineClient.IsContainerCertified(registry, repository, tag, digest)
	case offlineDBLoaded:
		logrus.Warnf("Online Catalog not available. Testing with offline db.")
		return offlineClient.IsContainerCertified(registry, repository, tag, digest)
	default:
		logrus.Errorf("Neither the online catalog nor the offline DB are available. Cannot verify the container certification status.")
		return false
	}
}
func IsOperatorCertified(operatorName, ocpVersion, channel string) bool {
	switch {
	case onlineClient.IsServiceReachable():
		return onlineClient.IsOperatorCertified(operatorName, ocpVersion, channel)
	case offlineDBLoaded:
		logrus.Warnf("Online Catalog not available. Testing with offline db.")
		return offlineClient.IsOperatorCertified(operatorName, ocpVersion, channel)
	default:
		logrus.Errorf("Neither the online catalog nor the offline DB are available. Cannot verify the operator certification status.")
		return false
	}
}
func IsReleaseCertified(helm *release.Release, ourKubeVersion string) bool {
	switch {
	case onlineClient.IsServiceReachable():
		return onlineClient.IsReleaseCertified(helm, ourKubeVersion)
	case offlineDBLoaded:
		logrus.Warnf("Online Catalog not available. Testing with offline db.")
		return offlineClient.IsReleaseCertified(helm, ourKubeVersion)
	default:
		logrus.Errorf("Neither the online catalog nor the offline DB are available. Cannot verify the Helm chart certification status.")
		return false
	}
}
