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
	"golang.org/x/mod/semver"

	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/internal/certdb"
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
	logrus.Debugf("Entering %s suite", common.AffiliatedCertTestKey)
	var env provider.TestEnvironment
	var validator certdb.CertificationStatusValidator
	ginkgo.BeforeEach(func() {
		var err error
		env = provider.GetTestEnvironment()
		validator, err = certdb.GetValidator(env.GetOfflineDBPath())
		if err != nil {
			errMsg := fmt.Sprintf("Cannot access the certification DB, err: %v", err)
			ginkgo.Fail(errMsg)
		}
	})
	ginkgo.ReportAfterEach(results.RecordResult)
	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestHelmVersionIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testHelmVersion(&env)
	})
	// Query API for certification status of listed containers
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestContainerIsCertifiedIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testContainerCertificationStatus(&env, validator)
	})

	// Query API for certification status of listed operators
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestOperatorIsCertifiedIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testAllOperatorCertified(&env, validator)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestHelmIsCertifiedIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testHelmCertified(&env, validator)
	})
	// Query API for certification status by digest of listed containers
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestContainerIsCertifiedDigestIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testContainerCertificationStatusByDigest(&env, validator)
	})
})

func getContainersToQuery(env *provider.TestEnvironment) map[configuration.ContainerImageIdentifier]bool {
	containersToQuery := make(map[configuration.ContainerImageIdentifier]bool)
	for _, c := range env.Config.CertifiedContainerInfo {
		containersToQuery[c] = true
	}
	if env.Config.CheckDiscoveredContainerCertificationStatus {
		for _, cut := range env.Containers {
			containersToQuery[cut.ContainerImageIdentifier] = true
		}
	}
	return containersToQuery
}

func testContainerCertification(c configuration.ContainerImageIdentifier, validator certdb.CertificationStatusValidator) bool {
	tag := c.Tag
	digest := c.Digest
	registryName := c.Repository
	name := c.Name
	ans := validator.IsContainerCertified(registryName, name, tag, digest)
	if !ans {
		tnf.ClaimFilePrintf("%s/%s:%s is not listed in certified containers", registryName, name, tag)
	}
	return ans
}

func testContainerCertificationStatus(env *provider.TestEnvironment, validator certdb.CertificationStatusValidator) {
	containersToQuery := getContainersToQuery(env)
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
		if !testContainerCertification(c, validator) {
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

func testAllOperatorCertified(env *provider.TestEnvironment, validator certdb.CertificationStatusValidator) {
	operatorsUnderTest := env.Operators
	testhelper.SkipIfEmptyAny(ginkgo.Skip, operatorsUnderTest)
	ginkgo.By(fmt.Sprintf("Verify operator as certified. Number of operators to check: %d", len(operatorsUnderTest)))
	testFailed := false
	ocpMinorVersion := ""
	if provider.IsOCPCluster() {
		// Converts	major.minor.patch version format to major.minor
		const majorMinorPatchCount = 3
		splitVersion := strings.SplitN(env.OpenshiftVersion, ".", majorMinorPatchCount)
		ocpMinorVersion = splitVersion[0] + "." + splitVersion[1]
	}
	for i := range operatorsUnderTest {
		name := operatorsUnderTest[i].Name
		channel := operatorsUnderTest[i].Channel
		isCertified := validator.IsOperatorCertified(name, ocpMinorVersion, channel)
		if !isCertified {
			testFailed = true
			logrus.Infof(
				"Operator %s (channel %s) not certified for OpenShift %s.",
				name,
				channel,
				ocpMinorVersion,
			)
			tnf.ClaimFilePrintf("Operator %s (channel %s) failed to be certified for OpenShift %s", name, channel, ocpMinorVersion)
		} else {
			logrus.Infof("Operator %s (channel %s) certified OK.", name, channel)
		}
	}
	if testFailed {
		ginkgo.Fail("At least one operator was not certified to run on this version of OpenShift. Check Claim.json file for details.")
	}
}

func testHelmCertified(env *provider.TestEnvironment, validator certdb.CertificationStatusValidator) {
	helmchartsReleases := env.HelmChartReleases
	testhelper.SkipIfEmptyAny(ginkgo.Skip, helmchartsReleases)
	// Collect all of the failed helm charts
	failedHelmCharts := [][]string{}
	for _, helm := range helmchartsReleases {
		if !validator.IsHelmChartCertified(helm, env.K8sVersion) {
			failedHelmCharts = append(failedHelmCharts, []string{helm.Chart.Metadata.Version, helm.Name})
			tnf.ClaimFilePrintf(
				"Helm Chart %s version %s is not certified.",
				helm.Name,
				helm.Chart.Metadata.Version,
			)
		} else {
			logrus.Infof(
				"Helm Chart %s version %s is certified.",
				helm.Name,
				helm.Chart.Metadata.Version,
			)
		}
	}
	testhelper.AddTestResultLog("Non-compliant", failedHelmCharts, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testContainerCertificationStatusByDigest(env *provider.TestEnvironment, validator certdb.CertificationStatusValidator) {
	failedContainers := []configuration.ContainerImageIdentifier{}
	for _, c := range env.Containers {
		if c.ContainerImageIdentifier.Name == "" || c.ContainerImageIdentifier.Repository == "" {
			tnf.ClaimFilePrintf("Container name = %q or repository = %q is missing, skipping this container to query", c.ContainerImageIdentifier.Name, c.ContainerImageIdentifier.Repository)
			continue
		}
		if c.ContainerImageIdentifier.Digest == "" {
			tnf.ClaimFilePrintf("%s is missing digest field, failing validation (repo=%s image=%s)", c, c.ContainerImageIdentifier.Repository, c.ContainerImageIdentifier.Name)
			failedContainers = append(failedContainers, c.ContainerImageIdentifier)
		} else if !testContainerCertification(c.ContainerImageIdentifier, validator) {
			tnf.ClaimFilePrintf("%s digest not found in database, failing validation (repo=%s image=%s)", c, c.ContainerImageIdentifier.Repository, c.ContainerImageIdentifier.Name)
			failedContainers = append(failedContainers, c.ContainerImageIdentifier)
		}
	}
	testhelper.AddTestResultLog("Non-compliant", failedContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testHelmVersion(env *provider.TestEnvironment) {
	helmchartsReleases := env.HelmChartReleases
	// Collect all of the failed helm charts
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, helm := range helmchartsReleases {
		helmChartVersion := helm.Chart.Metadata.APIVersion
		if !semver.IsValid(helmChartVersion) {
			logrus.Errorf("Failed to parse helm %s version %s, but its major should be v3", helm.Name, helm.Chart.Metadata.Version)
			reportObject := testhelper.NewReportObject("Failed to parse helm", testhelper.HelmChart, false).AddField(testhelper.ChartName, helm.Name)
			reportObject = reportObject.AddField(testhelper.ChartVersion, helm.Chart.Metadata.Version)
			nonCompliantObjects = append(nonCompliantObjects, reportObject)
		}
		charAPIVersionMajor := semver.Major(helmChartVersion)
		if charAPIVersionMajor != "v3" {
			reportObject := testhelper.NewReportObject(fmt.Sprintf("This Helm Chart is v%v but needs to be v3 due to the security risks associated with Tiller", helmChartVersion), testhelper.HelmChart, false).AddField(testhelper.ChartName, helm.Name)
			reportObject = reportObject.AddField(testhelper.ChartVersion, helm.Chart.Metadata.Version)
			nonCompliantObjects = append(nonCompliantObjects, reportObject)
		} else {
			reportObject := testhelper.NewReportObject("Helm Chart version is v3", testhelper.HelmChart, true).AddField(testhelper.ChartName, helm.Name)
			reportObject = reportObject.AddField(testhelper.ChartVersion, helm.Chart.Metadata.Version)
			compliantObjects = append(compliantObjects, reportObject)
		}
	}
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}
