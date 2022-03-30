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
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/internal/api"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/certification/certtool"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

const (
	// timeout for eventually call
	apiRequestTimeout = 40 * time.Second
	CertifiedOperator = "certified-operators"
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
		env = provider.GetTestEnvironment()
	})
	ginkgo.ReportAfterEach(results.RecordResult)

	testContainerCertificationStatus(&env)
	testAllOperatorCertified(&env)
	testHelmCertified(&env)
})

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
		certtool.CertAPIClient = api.NewHTTPClient()
		failedContainers := []configuration.ContainerImageIdentifier{}
		allContainersToQueryEmpty := true
		for c := range containersToQuery {
			if c.Name == "" || c.Repository == "" {
				tnf.ClaimFilePrintf("Container name = \"%s\" or repository = \"%s\" is missing, skipping this container to query", c.Name, c.Repository)
				continue
			}
			allContainersToQueryEmpty = false
			ginkgo.By(fmt.Sprintf("Container %s/%s should eventually be verified as certified", c.Repository, c.Name))
			entry := certtool.WaitForCertificationRequestToSuccess(certtool.GetContainerCertificationRequestFunction(c), apiRequestTimeout).(*api.ContainerCatalogEntry)
			if entry == nil {
				tnf.ClaimFilePrintf("Container %s (repository %s) is not found in the certified container catalog.", c.Name, c.Repository)
				failedContainers = append(failedContainers, c)
			} else {
				if entry.GetBestFreshnessGrade() > "C" {
					tnf.ClaimFilePrintf("Container %s (repository %s) is found in the certified container catalog but with low health index '%s'.", c.Name, c.Repository, entry.GetBestFreshnessGrade())
					failedContainers = append(failedContainers, c)
				}
				logrus.Info(fmt.Sprintf("Container %s (repository %s) is certified.", c.Name, c.Repository))
			}
		}
		if allContainersToQueryEmpty {
			ginkgo.Skip("No containers to check because either container name or repository is empty for all containers in tnf_config.yml")
		}

		if n := len(failedContainers); n > 0 {
			logrus.Warnf("Containers that are not certified: %+v", failedContainers)
			ginkgo.Fail(fmt.Sprintf("%d container images are not certified.", n))
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
		certtool.CertAPIClient = api.NewHTTPClient()
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
				isCertified := certtool.WaitForCertificationRequestToSuccess(certtool.GetOperatorCertificationRequestFunction(org, pack, ocpversion), apiRequestTimeout).(bool)
				if !isCertified {
					testFailed = true
					logrus.Info(fmt.Sprintf("Operator %s (organization %s) not certified for Openshift %s .", pack, org, ocpversion))
					tnf.ClaimFilePrintf("Operator %s (organization %s) failed to be certified for Openshift %s", pack, org, ocpversion)
				} else {
					logrus.Info(fmt.Sprintf("Operator %s (organization %s) certified OK.", pack, org))
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
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestHelmIsCertifiedIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		certtool.CertAPIClient = api.NewHTTPClient()
		helmchartsReleases := env.HelmChartReleases
		if len(helmchartsReleases) == 0 {
			ginkgo.Skip("No helm charts to check")
		}
		out, err := certtool.CertAPIClient.GetYamlFile()
		if err != nil {
			ginkgo.Fail(fmt.Sprintf("error while reading the helm yaml file from the api %s", err))
		}
		if out.Entries == nil {
			ginkgo.Skip("No helm charts from the api")
		}

		// Collect all of the failed helm charts
		failedHelmCharts := [][]string{}
		for _, helm := range helmchartsReleases {
			if !certtool.IsReleaseCertified(helm, env.K8sVersion, out) {
				failedHelmCharts = append(failedHelmCharts, []string{helm.Chart.Metadata.Version, helm.Name})
			} else {
				logrus.Info(fmt.Sprintf("Helm %s with version %s is certified", helm.Name, helm.Chart.Metadata.Version))
			}
		}
		if len(failedHelmCharts) > 0 {
			logrus.Errorf("Helms that are not certified: %+v", failedHelmCharts)
			tnf.ClaimFilePrintf("Helms that are not certified: %+v", failedHelmCharts)
			ginkgo.Fail(fmt.Sprintf("%d helms chart are not certified.", len(failedHelmCharts)))
		}
	})
}
