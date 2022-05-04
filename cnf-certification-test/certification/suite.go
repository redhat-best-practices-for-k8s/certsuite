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
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/internal/api"
	"github.com/test-network-function/cnf-certification-test/internal/registry"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/certification/certtool"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

const (
	// timeout for eventually call
	CertifiedOperator = "certified-operators"
	Online            = "online"
)

var _ = ginkgo.Describe(common.AffiliatedCertTestKey, func() {
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
		registry.LoadCatalogs()
	})
	ginkgo.ReportAfterEach(results.RecordResult)

	testContainerCertificationStatus(&env)
	testAllOperatorCertified(&env)
	testHelmCertified(&env)
})

func testContainerCertification(c configuration.ContainerImageIdentifier) bool {
	tag := c.Tag
	digest := c.Digest
	registryName := c.Repository
	name := c.Name
	ans := registry.IsCertified(registryName, name, tag, digest)
	if !ans {
		tnf.ClaimFilePrintf("%s/%s:%s is not listed in certified containers", registryName, name, tag)
	}
	return ans
}

func testContainerCertificationStatus(env *provider.TestEnvironment) {
	// Query API for certification status of listed containers
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestContainerIsCertifiedIdentifier)
	ginkgo.It(testID, ginkgo.Label(Online, testID), func() {
		containersToQuery := certtool.GetContainersToQuery(env)
		if len(containersToQuery) == 0 {
			ginkgo.Skip("No containers to check configured in tnf_config.yml")
		}
		ginkgo.By(fmt.Sprintf("Getting certification status. Number of containers to check: %d", len(containersToQuery)))
		failedContainers := []configuration.ContainerImageIdentifier{}
		allContainersToQueryEmpty := true
		for c := range containersToQuery {
			if c.Name == "" || c.Repository == "" {
				tnf.ClaimFilePrintf("Container name = \"%s\" or repository = \"%s\" is missing, skipping this container to query", c.Name, c.Repository)
				continue
			}
			allContainersToQueryEmpty = false
			if !testContainerCertification(c) {
				failedContainers = append(failedContainers, c)
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
	ginkgo.It(testID, ginkgo.Label(Online, testID), func() {
		operatorsToQuery := env.Subscriptions
		testhelper.SkipIfEmptyAny(ginkgo.Skip, operatorsToQuery)
		ginkgo.By(fmt.Sprintf("Verify operator as certified. Number of operators to check: %d", len(operatorsToQuery)))
		testFailed := false
		ocpMinorVersion := ""
		if env.OpenshiftVersion != "" {
			// Converts	major.minor.patch version format to major.minor
			const majorMinorPatchCount = 3
			splitVersion := strings.SplitN(env.OpenshiftVersion, ".", majorMinorPatchCount)
			ocpMinorVersion = splitVersion[0] + "." + splitVersion[1]
		}
		for _, op := range operatorsToQuery {
			pack := op.Status.InstalledCSV
			isCertified := registry.IsOperatorCertified(pack, ocpMinorVersion)
			if !isCertified {
				testFailed = true
				logrus.Info(fmt.Sprintf("Operator %s not certified for OpenShift %s .", pack, ocpMinorVersion))
				tnf.ClaimFilePrintf("Operator %s  failed to be certified for OpenShift %s", pack, ocpMinorVersion)
			} else {
				logrus.Info(fmt.Sprintf("Operator %s certified OK.", pack))
			}
		}
		if testFailed {
			ginkgo.Fail("At least one operator was not certified to run on this version of OpenShift. Check Claim.json file for details.")
		}
	})
}

func testHelmCertified(env *provider.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestHelmIsCertifiedIdentifier)
	ginkgo.It(testID, ginkgo.Label(Online, testID), func() {
		certtool.CertAPIClient = api.NewHTTPClient()
		helmchartsReleases := env.HelmChartReleases
		testhelper.SkipIfEmptyAny(ginkgo.Skip, helmchartsReleases)
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
