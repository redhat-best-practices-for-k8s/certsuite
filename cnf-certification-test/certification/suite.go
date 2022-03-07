// Copyright (C) 2020-2021 Red Hat, Inc.
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

package certification

import (
	//"fmt"
	"fmt"
	"strings"
	"time"

	version "github.com/hashicorp/go-version"
	"github.com/onsi/ginkgo/v2"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/internal/api"

	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

const (
	// timeout for eventually call
	apiRequestTimeout           = 40 * time.Second
	expectersVerboseModeEnabled = false
	CertifiedOperator           = "certified-operators"
	outMinikubeVersion          = "null"
)

var (
	certAPIClient api.CertAPIClient
)

type ChartStruct struct {
	Entries map[string][]struct {
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		KubeVersion string `yaml:"kubeVersion"`
	} `yaml:"entries"`
}

var _ = ginkgo.Describe(common.AffiliatedCertTestKey, func() {
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		provider.BuildTestEnvironment()
		env = provider.GetTestEnvironment()
	})
	//testContainerCertificationStatus(&env)
	//testAllOperatorCertified(&env)
	testHelmCertified(&env)
})

// getContainerCertificationRequestFunction returns function that will try to get the certification status (CCP) for a container.
func getContainerCertificationRequestFunction(id configuration.ContainerImageIdentifier) func() (interface{}, error) {
	return func() (interface{}, error) {
		return certAPIClient.GetContainerCatalogEntry(id)
	}
}

// getOperatorCertificationRequestFunction returns function that will try to get the certification status (OCP) for an operator.
func getOperatorCertificationRequestFunction(organization, operatorName, ocpversion string) func() (interface{}, error) {
	return func() (interface{}, error) {
		return certAPIClient.IsOperatorCertified(organization, operatorName, ocpversion)
	}
}

// waitForCertificationRequestToSuccess calls to certificationRequestFunc until it returns true.
func waitForCertificationRequestToSuccess(certificationRequestFunc func() (interface{}, error), timeout time.Duration) interface{} {
	const pollingPeriod = 1 * time.Second
	var elapsed time.Duration
	var err error
	var result interface{}

	for elapsed < timeout {
		result, err = certificationRequestFunc()

		if err == nil {
			break
		}
		time.Sleep(pollingPeriod)
		elapsed += pollingPeriod
	}
	return result
}
func testContainerCertificationStatus(env *provider.TestEnvironment) {
	// Query API for certification status of listed containers
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestContainerIsCertifiedIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		containersToQuery := make(map[configuration.ContainerImageIdentifier]bool)

		for _, c := range env.Config.CertifiedContainerInfo {
			containersToQuery[c] = true
		}
		if env.Config.CheckDiscoveredContainerCertificationStatus {
			for _, cut := range env.Containers {
				containersToQuery[cut.ContainerImageIdentifier] = true
			}
		}
		if len(containersToQuery) == 0 {
			ginkgo.Skip("No containers to check configured in tnf_config.yml")
		}
		ginkgo.By(fmt.Sprintf("Getting certification status. Number of containers to check: %d", len(containersToQuery)))
		if len(containersToQuery) > 0 {
			certAPIClient = api.NewHTTPClient()
			failedContainers := []configuration.ContainerImageIdentifier{}
			allContainersToQueryEmpty := true
			for c := range containersToQuery {
				if c.Name == "" || c.Repository == "" {
					tnf.ClaimFilePrintf("Container name = \"%s\" or repository = \"%s\" is missing, skipping this container to query", c.Name, c.Repository)
					continue
				}
				allContainersToQueryEmpty = false
				ginkgo.By(fmt.Sprintf("Container %s/%s should eventually be verified as certified", c.Repository, c.Name))
				entry := waitForCertificationRequestToSuccess(getContainerCertificationRequestFunction(c), apiRequestTimeout).(*api.ContainerCatalogEntry)
				if entry == nil {
					tnf.ClaimFilePrintf("Container %s (repository %s) is not found in the certified container catalog.", c.Name, c.Repository)
					failedContainers = append(failedContainers, c)
				} else {
					if entry.GetBestFreshnessGrade() > "C" {
						tnf.ClaimFilePrintf("Container %s (repository %s) is found in the certified container catalog but with low health index '%s'.", c.Name, c.Repository, entry.GetBestFreshnessGrade())
						failedContainers = append(failedContainers, c)
					}
					log.Info(fmt.Sprintf("Container %s (repository %s) is certified.", c.Name, c.Repository))
				}
			}
			if allContainersToQueryEmpty {
				ginkgo.Skip("No containers to check because either container name or repository is empty for all containers in tnf_config.yml")
			}

			if n := len(failedContainers); n > 0 {
				log.Warnf("Containers that are not certified: %+v", failedContainers)
				ginkgo.Fail(fmt.Sprintf("%d container images are not certified.", n))
			}
		}
	})
}

func testAllOperatorCertified(env *provider.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestOperatorIsCertifiedIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		operatorsToQuery := env.Subscriptions

		if len(operatorsToQuery) == 0 {
			ginkgo.Skip("No operators to check configured ")
		}
		certAPIClient = api.NewHTTPClient()
		ginkgo.By(fmt.Sprintf("Verify operator as certified. Number of operators to check: %d", len(operatorsToQuery)))
		testFailed := false
		for _, op := range operatorsToQuery {
			ocpversion := ""
			if env.OpenshiftVersion != "" {
				ocpversion = env.OpenshiftVersion
			}
			pack := op.Status.InstalledCSV
			org := op.Spec.CatalogSource
			if org == CertifiedOperator {
				isCertified := waitForCertificationRequestToSuccess(getOperatorCertificationRequestFunction(org, pack, ocpversion), apiRequestTimeout).(bool)
				if !isCertified {
					testFailed = true
					log.Info(fmt.Sprintf("Operator %s (organization %s) not certified for Openshift %s .", pack, org, ocpversion))
					tnf.ClaimFilePrintf("Operator %s (organization %s) failed to be certified for Openshift %s", pack, org, ocpversion)
				} else {
					log.Info(fmt.Sprintf("Operator %s (organization %s) certified OK.", pack, org))
				}
			} else {
				testFailed = true
				tnf.ClaimFilePrintf("Operator %s is not certified (needs to be part of the operator-certified organization in the catalog)", pack)
			}
		}
		if testFailed {
			ginkgo.Skip("At least one  operator was not certified to run on this version of openshift. Check Claim.json file for details.")
		}
	})
}
func testHelmCertified(env *provider.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestOperatorIsCertifiedIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		certAPIClient = api.NewHTTPClient()
		helmcharts := env.HelmList
		if len(helmcharts) == 0 {
			ginkgo.Skip("No helm charts to check")
		}
		out, err := certAPIClient.GetYamlFile()
		if err != nil {
			ginkgo.Fail(fmt.Sprintf("error while reading the helm yaml file from the api %s", err))
		}
		if out.Entries == nil {
			ginkgo.Skip("No helm charts from the api")
		}
		ourKubeVersion := env.K8sVersion
		failedHelmCharts := [][]string{}
		for _, helm := range helmcharts {
			certified := false
			for _, entryList := range out.Entries {
				for _, entry := range entryList {
					if entry.Name == helm.Chart.Metadata.Name && entry.Version == helm.Chart.Metadata.Version {
						if entry.KubeVersion != "" {
							if CompareVersion(ourKubeVersion, entry.KubeVersion) {
								certified = true
								break
							}
						} else {
							certified = true
							break
						}
					}
				}
				if certified {
					log.Info(fmt.Sprintf("Helm %s with version %s is certified", helm.Name, helm.Chart.Metadata.Version))
					break
				}
			}
			if !certified {
				fail := []string{helm.Chart.Metadata.Version, helm.Name}
				failedHelmCharts = append(failedHelmCharts, fail)
			}
		}
		if len(failedHelmCharts) > 0 {
			log.Errorf("Helms that are not certified: %+v", failedHelmCharts)
			tnf.ClaimFilePrintf("Helms that are not certified: %+v", failedHelmCharts)
			ginkgo.Fail(fmt.Sprintf("%d helms chart are not certified.", len(failedHelmCharts)))
		}
	})
}
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
