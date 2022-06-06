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
	IsOperatorCertified(operatorName, ocpVersion, channel string) bool
	IsReleaseCertified(helm *release.Release, ourKubeVersion string) bool
	IsServiceReachable() bool
}

var onlineClient CertificationValidator
var offlineClient CertificationValidator

func init() {
	logrus.Info("init certification client")
	onlineClient = onlinecheck.NewOnlineValidator()
	offlineClient = offlinecheck.OfflineChecker{}
}
func LoadCatalog() {
	offlinecheck.LoadCatalogs()
}

func IsContainerCertified(registry, repository, tag, digest string) bool {
	if onlineClient.IsServiceReachable() {
		return onlineClient.IsContainerCertified(registry, repository, tag, digest)
	}
	return offlineClient.IsContainerCertified(registry, repository, tag, digest)
}
func IsOperatorCertified(operatorName, ocpVersion, channel string) bool {
	if onlineClient.IsServiceReachable() {
		return onlineClient.IsOperatorCertified(operatorName, ocpVersion, channel)
	}
	return offlineClient.IsOperatorCertified(operatorName, ocpVersion, channel)
}
func IsReleaseCertified(helm *release.Release, ourKubeVersion string) bool {
	if onlineClient.IsServiceReachable() {
		return onlineClient.IsReleaseCertified(helm, ourKubeVersion)
	}
	return offlineClient.IsReleaseCertified(helm, ourKubeVersion)
}
