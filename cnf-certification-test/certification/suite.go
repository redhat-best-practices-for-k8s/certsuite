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

package certification

import (
	"fmt"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/certification/certtool"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	api "github.com/test-network-function/cnf-certification-test/internal/api"
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
		api.LoadCatalog()
	})
	ginkgo.ReportAfterEach(results.RecordResult)

	// Query API for certification status of listed containers
	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestContainerIsCertifiedIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		//testContainerCertificationStatus(&env)
	})

	// Query API for certification status of listed operators
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestOperatorIsCertifiedIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		//testAllOperatorCertified(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestHelmIsCertifiedIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		//	testHelmCertified(&env)
	})
	// Query API for certification status by digest of listed containers
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestContainerIsCertifiedDigestIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testContainerCertificationStatusByDigest(&env)
	})
})

func testContainerCertification(c configuration.ContainerImageIdentifier) bool {
	tag := c.Tag
	digest := c.Digest
	registryName := c.Repository
	name := c.Name
	ans := api.IsContainerCertified(registryName, name, tag, digest)
	if !ans {
		tnf.ClaimFilePrintf("%s/%s:%s is not listed in certified containers", registryName, name, tag)
	}
	return ans
}

func testContainerCertificationStatus(env *provider.TestEnvironment) {
	containersToQuery := certtool.GetContainersToQuery(env)
	testhelper.SkipIfEmptyAny(ginkgo.Skip, containersToQuery)
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
}

func testAllOperatorCertified(env *provider.TestEnvironment) {
	operatorsUnderTest := env.Operators
	testhelper.SkipIfEmptyAny(ginkgo.Skip, operatorsUnderTest)
	ginkgo.By(fmt.Sprintf("Verify operator as certified. Number of operators to check: %d", len(operatorsUnderTest)))
	testFailed := false
	ocpMinorVersion := ""
	if env.OpenshiftVersion != "" {
		// Converts	major.minor.patch version format to major.minor
		const majorMinorPatchCount = 3
		splitVersion := strings.SplitN(env.OpenshiftVersion, ".", majorMinorPatchCount)
		ocpMinorVersion = splitVersion[0] + "." + splitVersion[1]
	}
	for i := range operatorsUnderTest {
		name := operatorsUnderTest[i].Name
		channel := operatorsUnderTest[i].Channel
		isCertified := api.IsOperatorCertified(name, ocpMinorVersion, channel)
		if !isCertified {
			testFailed = true
			logrus.Info(fmt.Sprintf("Operator %s (channel %s) not certified for OpenShift %s .", name, channel, ocpMinorVersion))
			tnf.ClaimFilePrintf("Operator %s (channel %s) failed to be certified for OpenShift %s", name, channel, ocpMinorVersion)
		} else {
			logrus.Info(fmt.Sprintf("Operator %s (channel %s) certified OK.", name, channel))
		}
	}
	if testFailed {
		ginkgo.Fail("At least one operator was not certified to run on this version of OpenShift. Check Claim.json file for details.")
	}
}

func testHelmCertified(env *provider.TestEnvironment) {
	helmchartsReleases := env.HelmChartReleases
	testhelper.SkipIfEmptyAny(ginkgo.Skip, helmchartsReleases)
	// Collect all of the failed helm charts
	failedHelmCharts := [][]string{}
	for _, helm := range helmchartsReleases {
		if !api.IsReleaseCertified(helm, env.K8sVersion) {
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
}

func testContainerCertificationStatusByDigest(env *provider.TestEnvironment) {
	failedContainers := []configuration.ContainerImageIdentifier{}
	for _, c := range env.Containers {
		if c.ContainerImageIdentifier.Name == "" || c.ContainerImageIdentifier.Repository == "" {
			tnf.ClaimFilePrintf("Container name = %q or repository = %q is missing, skipping this container to query", c.ContainerImageIdentifier.Name, c.ContainerImageIdentifier.Repository)
			continue
		}
		if c.ContainerImageIdentifier.Digest == "" {
			tnf.ClaimFilePrintf("%s is missing digest field, failing validation (repo=%s image=%s)", c.ContainerImageIdentifier.Repository, c.ContainerImageIdentifier.Name)
			failedContainers = append(failedContainers, c.ContainerImageIdentifier)
		} else if !testContainerCertification(c.ContainerImageIdentifier) {
			tnf.ClaimFilePrintf("%s digest not found in database, failing validation (repo=%s image=%s)", c.ContainerImageIdentifier.Repository, c.ContainerImageIdentifier.Name)
			failedContainers = append(failedContainers, c.ContainerImageIdentifier)
		} else {
			logrus.Info(fmt.Sprintf("container %v is certified by digest.", c.ContainerImageIdentifier))
		}
	}
	testhelper.AddTestResultLog("Non-compliant", failedContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
}
