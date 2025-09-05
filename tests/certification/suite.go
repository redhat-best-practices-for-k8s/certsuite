// Copyright (C) 2020-2024 Red Hat, Inc.
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

	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/oct/pkg/certdb"
)

const (
	// timeout for eventually call
	CertifiedOperator = "certified-operators"
	Online            = "online"
)

var (
	env       provider.TestEnvironment
	validator certdb.CertificationStatusValidator

	beforeEachFn = func(check *checksdb.Check) error {
		env = provider.GetTestEnvironment()

		var err error
		validator, err = certdb.GetValidator(env.GetOfflineDBPath())
		if err != nil {
			return fmt.Errorf("cannot access the certification DB, err: %v", err)
		}

		return nil
	}

	skipIfNoOperatorsFn = func() (bool, string) {
		if len(env.Operators) == 0 {
			return true, "There are no operators to check. Please check under test labels."
		}

		return false, ""
	}

	skipIfNoHelmChartReleasesFn = func() (bool, string) {
		if len(env.HelmChartReleases) == 0 {
			return true, "There are no helm chart releases to check."
		}

		return false, ""
	}
)

// LoadChecks Loads the suite of certification checks
//
// This function registers several checks for Helm, operators, Helm charts, and
// container certifications by creating a group in the checks database. It
// attaches skip functions that prevent tests from running when necessary data
// is missing and assigns each check a callback to perform its validation logic.
// The function logs its activity at debug level and relies on global
// environment and validator objects for execution.
func LoadChecks() {
	log.Debug("Loading %s suite checks", common.AffiliatedCertTestKey)

	checksGroup := checksdb.NewChecksGroup(common.AffiliatedCertTestKey).
		WithBeforeEachFn(beforeEachFn)

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestHelmVersionIdentifier)).
		WithSkipCheckFn(skipIfNoHelmChartReleasesFn).
		WithCheckFn(func(check *checksdb.Check) error {
			testHelmVersion(check)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestOperatorIsCertifiedIdentifier)).
		WithSkipCheckFn(skipIfNoOperatorsFn).
		WithCheckFn(func(c *checksdb.Check) error {
			testAllOperatorCertified(c, &env, validator)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestHelmIsCertifiedIdentifier)).
		WithSkipCheckFn(skipIfNoHelmChartReleasesFn).
		WithCheckFn(func(c *checksdb.Check) error {
			testHelmCertified(c, &env, validator)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestContainerIsCertifiedDigestIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainerCertificationStatusByDigest(c, &env, validator)
			return nil
		}))
}

// getContainersToQuery Creates a set of container image identifiers for querying
//
// The function iterates over the containers defined in the test environment,
// adding each container's image identifier to a map with a true value. This map
// represents the collection of images that should be queried during testing. It
// returns the constructed map.
func getContainersToQuery(env *provider.TestEnvironment) map[provider.ContainerImageIdentifier]bool {
	containersToQuery := make(map[provider.ContainerImageIdentifier]bool)
	for _, cut := range env.Containers {
		containersToQuery[cut.ContainerImageIdentifier] = true
	}
	return containersToQuery
}

// testContainerCertification Checks if a container image is certified in the database
//
// This function accepts an image identifier and a validator, delegating to the
// validator's method to determine certification status. It returns true when
// the image's registry, repository, tag, and digest match a certified record,
// otherwise false. The result informs test logic that verifies container
// compliance.
func testContainerCertification(c provider.ContainerImageIdentifier, validator certdb.CertificationStatusValidator) bool {
	return validator.IsContainerCertified(c.Registry, c.Repository, c.Tag, c.Digest)
}

// testAllOperatorCertified Verifies that all operators are certified for the current OpenShift version
//
// The function iterates over every operator listed in the test environment,
// determining whether each is certified for the detected OpenShift minor
// release. It logs successes or failures and builds separate lists of compliant
// and non‑compliant report objects. Finally, it records these results in the
// check’s outcome.
func testAllOperatorCertified(check *checksdb.Check, env *provider.TestEnvironment, validator certdb.CertificationStatusValidator) {
	operatorsUnderTest := env.Operators
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	ocpMinorVersion := ""
	if provider.IsOCPCluster() {
		// Converts	major.minor.patch version format to major.minor
		const majorMinorPatchCount = 3
		splitVersion := strings.SplitN(env.OpenshiftVersion, ".", majorMinorPatchCount)
		ocpMinorVersion = splitVersion[0] + "." + splitVersion[1]
	}
	for _, operator := range operatorsUnderTest {
		check.LogInfo("Testing Operator %q", operator)
		isCertified := validator.IsOperatorCertified(operator.Name, ocpMinorVersion)
		if !isCertified {
			check.LogError("Operator %q (channel %q) failed to be certified for OpenShift %s", operator.Name, operator.Channel, ocpMinorVersion)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewOperatorReportObject(operator.Namespace, operator.Name, "Operator failed to be certified for OpenShift", false).
				AddField(testhelper.OCPVersion, ocpMinorVersion).
				AddField(testhelper.OCPChannel, operator.Channel))
		} else {
			check.LogInfo("Operator %q (channel %q) is certified for OpenShift %s", operator.Name, operator.Channel, ocpMinorVersion)
			compliantObjects = append(compliantObjects, testhelper.NewOperatorReportObject(operator.Namespace, operator.Name, "Operator certified OK", true).
				AddField(testhelper.OCPVersion, ocpMinorVersion).
				AddField(testhelper.OCPChannel, operator.Channel))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testHelmCertified Verifies each Helm chart release against certification rules
//
// The function iterates over all Helm chart releases in the test environment,
// logging status for each. It uses a validator to determine if a chart is
// certified for the current Kubernetes version and records compliant or
// non‑compliant reports accordingly. Finally, it sets the overall test result
// with lists of both compliant and non‑compliant objects.
func testHelmCertified(check *checksdb.Check, env *provider.TestEnvironment, validator certdb.CertificationStatusValidator) {
	helmchartsReleases := env.HelmChartReleases

	// Collect all of the failed helm charts
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, helm := range helmchartsReleases {
		check.LogInfo("Testing Helm Chart Release %q", helm.Name)
		if !validator.IsHelmChartCertified(helm, env.K8sVersion) {
			check.LogError("Helm Chart %q version %q is not certified.", helm.Name, helm.Chart.Metadata.Version)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewHelmChartReportObject(helm.Namespace, helm.Name, "helm chart is not certified", false).
				SetType(testhelper.HelmVersionType).
				AddField(testhelper.Version, helm.Chart.Metadata.Version))
		} else {
			check.LogInfo("Helm Chart %q version %q is certified.", helm.Name, helm.Chart.Metadata.Version)
			compliantObjects = append(compliantObjects, testhelper.NewHelmChartReportObject(helm.Namespace, helm.Name, "helm chart is certified", true).
				SetType(testhelper.HelmVersionType).
				AddField(testhelper.Version, helm.Chart.Metadata.Version))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testContainerCertificationStatusByDigest Validates container digests against certification database
//
// The function iterates over containers in the test environment, checking that
// each has a digest and that the digest exists in the certification database.
// Containers missing a digest or with an unregistered digest are logged as
// errors and added to non‑compliant results; compliant ones are recorded
// accordingly. Finally, it sets the check result with lists of compliant and
// non‑compliant containers.
func testContainerCertificationStatusByDigest(check *checksdb.Check, env *provider.TestEnvironment, validator certdb.CertificationStatusValidator) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, c := range env.Containers {
		check.LogInfo("Testing Container %q", c)
		switch {
		case c.ContainerImageIdentifier.Digest == "":
			check.LogError("Container %q is missing digest field, failing validation (repo=%q image=%q)", c, c.ContainerImageIdentifier.Registry, c.ContainerImageIdentifier.Repository)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(c.Namespace, c.Podname, c.Name, "Missing digest field", false).
				AddField(testhelper.Repository, c.ContainerImageIdentifier.Registry).
				AddField(testhelper.ImageName, c.ContainerImageIdentifier.Repository).
				AddField(testhelper.ImageDigest, c.ContainerImageIdentifier.Digest))
		case !testContainerCertification(c.ContainerImageIdentifier, validator):
			check.LogError("Container %q digest not found in database, failing validation (repo=%q image=%q tag=%q digest=%q)", c,
				c.ContainerImageIdentifier.Registry, c.ContainerImageIdentifier.Repository,
				c.ContainerImageIdentifier.Tag, c.ContainerImageIdentifier.Digest)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(c.Namespace, c.Podname, c.Name, "Digest not found in database", false).
				AddField(testhelper.Repository, c.ContainerImageIdentifier.Registry).
				AddField(testhelper.ImageName, c.ContainerImageIdentifier.Repository).
				AddField(testhelper.ImageDigest, c.ContainerImageIdentifier.Digest))
		default:
			check.LogInfo("Container %q digest found in database, image certified (repo=%q image=%q tag=%q digest=%q)", c,
				c.ContainerImageIdentifier.Registry, c.ContainerImageIdentifier.Repository,
				c.ContainerImageIdentifier.Tag, c.ContainerImageIdentifier.Digest)
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(c.Namespace, c.Podname, c.Name, "Container is certified", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testHelmVersion Checks Helm release version compatibility
//
// This routine inspects the cluster for a Tiller pod to determine whether Helm
// v2 or v3 is in use. If no Tiller pod exists, it records all installed charts
// as compliant with Helm v3 and logs that v3 is being used. When a Tiller pod
// is found, it flags each such pod as non‑compliant because the required
// version is v3, then sets the test result accordingly.
func testHelmVersion(check *checksdb.Check) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	clients := clientsholder.GetClientsHolder()
	// Get the Tiller pod in the specified namespace
	podList, err := clients.K8sClient.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=helm,name=tiller",
	})
	if err != nil {
		check.LogError("Could not get Tiller pod, err=%v", err)
	}

	if len(podList.Items) == 0 {
		check.LogInfo("Tiller pod not found in any namespaces. Helm version is v3.")
		for _, helm := range env.HelmChartReleases {
			compliantObjects = append(compliantObjects, testhelper.NewHelmChartReportObject(helm.Namespace, helm.Name, "helm chart was installed with helm v3", true))
		}
	}

	check.LogError("Tiller pod found, Helm version is v2 but v3 required")
	for i := range podList.Items {
		nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(podList.Items[i].Namespace, podList.Items[i].Name,
			"This pod is a Tiller pod. Helm Chart version is v2 but needs to be v3 due to the security risks associated with Tiller", false))
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}
