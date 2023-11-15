// Copyright (C) 2020-2023 Red Hat, Inc.
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
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"

	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
	"github.com/test-network-function/oct/pkg/certdb"
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

	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestHelmVersionIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testHelmVersion()
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

func getContainersToQuery(env *provider.TestEnvironment) map[provider.ContainerImageIdentifier]bool {
	containersToQuery := make(map[provider.ContainerImageIdentifier]bool)
	for _, cut := range env.Containers {
		containersToQuery[cut.ContainerImageIdentifier] = true
	}
	return containersToQuery
}

func testContainerCertification(c provider.ContainerImageIdentifier, validator certdb.CertificationStatusValidator) bool {
	ans := validator.IsContainerCertified(c.Registry, c.Repository, c.Tag, c.Digest)
	if !ans {
		tnf.ClaimFilePrintf("%s/%s:%s is not listed in certified containers", c.Registry, c.Repository, c.Tag)
	}
	return ans
}

func testAllOperatorCertified(env *provider.TestEnvironment, validator certdb.CertificationStatusValidator) {
	operatorsUnderTest := env.Operators
	testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(operatorsUnderTest, "operatorsUnderTest"))
	ginkgo.By(fmt.Sprintf("Verify operator as certified. Number of operators to check: %d", len(operatorsUnderTest)))
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
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
			tnf.Logf(logrus.InfoLevel, "Operator %s (channel %s) failed to be certified for OpenShift %s", name, channel, ocpMinorVersion)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(operatorsUnderTest[i].Namespace, operatorsUnderTest[i].Name, "Operator failed to be certified for OpenShift", false).
				AddField(testhelper.OCPVersion, ocpMinorVersion).
				AddField(testhelper.OCPChannel, channel))
		} else {
			logrus.Infof("Operator %s (channel %s) certified OK.", name, channel)
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(operatorsUnderTest[i].Namespace, operatorsUnderTest[i].Name, "Operator certified OK", true).
				AddField(testhelper.OCPVersion, ocpMinorVersion).
				AddField(testhelper.OCPChannel, channel))
		}
	}
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testHelmCertified(env *provider.TestEnvironment, validator certdb.CertificationStatusValidator) {
	helmchartsReleases := env.HelmChartReleases
	testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(helmchartsReleases, "helmchartsReleases"))
	// Collect all of the failed helm charts
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, helm := range helmchartsReleases {
		if !validator.IsHelmChartCertified(helm, env.K8sVersion) {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewHelmChartReportObject(helm.Namespace, helm.Name, "helm chart is not certified", false).
				SetType(testhelper.HelmVersionType).
				AddField(testhelper.Version, helm.Chart.Metadata.Version))
			tnf.ClaimFilePrintf("Helm Chart %s version %s is not certified.", helm.Name, helm.Chart.Metadata.Version)
		} else {
			logrus.Infof("Helm Chart %s version %s is certified.", helm.Name, helm.Chart.Metadata.Version)
			compliantObjects = append(compliantObjects, testhelper.NewHelmChartReportObject(helm.Namespace, helm.Name, "helm chart is certified", true).
				SetType(testhelper.HelmVersionType).
				AddField(testhelper.Version, helm.Chart.Metadata.Version))
		}
	}
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testContainerCertificationStatusByDigest(env *provider.TestEnvironment, validator certdb.CertificationStatusValidator) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, c := range env.Containers {
		switch {
		case c.ContainerImageIdentifier.Digest == "":
			tnf.ClaimFilePrintf("%s is missing digest field, failing validation (repo=%s image=%s digest=%s)", c, c.ContainerImageIdentifier.Registry, c.ContainerImageIdentifier.Repository, c.ContainerImageIdentifier.Digest)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(c.Namespace, c.Podname, c.Name, "Missing digest field", false).
				AddField(testhelper.Repository, c.ContainerImageIdentifier.Registry).
				AddField(testhelper.ImageName, c.ContainerImageIdentifier.Repository).
				AddField(testhelper.ImageDigest, c.ContainerImageIdentifier.Digest))
		case !testContainerCertification(c.ContainerImageIdentifier, validator):
			tnf.ClaimFilePrintf("%s digest not found in database, failing validation (repo=%s image=%s digest=%s)", c, c.ContainerImageIdentifier.Registry, c.ContainerImageIdentifier.Repository, c.ContainerImageIdentifier.Digest)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(c.Namespace, c.Podname, c.Name, "Digest not found in database", false).
				AddField(testhelper.Repository, c.ContainerImageIdentifier.Registry).
				AddField(testhelper.ImageName, c.ContainerImageIdentifier.Repository).
				AddField(testhelper.ImageDigest, c.ContainerImageIdentifier.Digest))
		default:
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(c.Namespace, c.Podname, c.Name, "Container is certified", true))
		}
	}
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testHelmVersion() {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	clients := clientsholder.GetClientsHolder()
	// Get the Tiller pod in the specified namespace
	podList, err := clients.K8sClient.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=helm,name=tiller",
	})
	if err != nil {
		ginkgo.Fail(fmt.Sprintf("Error getting Tiller pod: %v\n", err))
	}
	if len(podList.Items) == 0 {
		tnf.ClaimFilePrintf("Tiller pod is not found in all namespaces helm version is v3\n")
	} else {
		tnf.ClaimFilePrintf("Tiller pod found, helm version is v2")
		for i := range podList.Items {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(podList.Items[i].Namespace, podList.Items[i].Name,
				"This pod is a Tiller pod. Helm Chart version is v2 but needs to be v3 due to the security risks associated with Tiller", false))
		}
	}
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}
