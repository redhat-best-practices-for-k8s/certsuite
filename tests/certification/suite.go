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

// LoadChecks loads and returns the certification checks for the suite.
//
// It constructs a group of checks, registers them with the test environment,
// and provides hooks to run before each check. The returned function
// performs any necessary cleanup or finalization when called.
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

// getContainersToQuery creates a lookup map of container image identifiers that should be queried in the current test environment.
//
// It inspects the provided TestEnvironment and constructs a map where each key is a
// ContainerImageIdentifier representing an image to check, and the value indicates
// whether that image should be included (true). The returned map can be used by
// other functions to determine which container images require certification checks.
func getContainersToQuery(env *provider.TestEnvironment) map[provider.ContainerImageIdentifier]bool {
	containersToQuery := make(map[provider.ContainerImageIdentifier]bool)
	for _, cut := range env.Containers {
		containersToQuery[cut.ContainerImageIdentifier] = true
	}
	return containersToQuery
}

// testContainerCertification checks whether a container image meets the certification criteria defined by a validator.
//
// It takes a ContainerImageIdentifier and a CertificationStatusValidator, invokes
// IsContainerCertified to determine if the image satisfies the required
// certification status, and returns true when it does, otherwise false.
func testContainerCertification(c provider.ContainerImageIdentifier, validator certdb.CertificationStatusValidator) bool {
	return validator.IsContainerCertified(c.Registry, c.Repository, c.Tag, c.Digest)
}

// testAllOperatorCertified checks that all operators in the cluster are certified.
//
// It iterates over each operator found by the test environment, verifies
// whether the cluster is an OCP cluster, splits operator names to extract
// namespace and name components, and calls IsOperatorCertified for each.
// The function logs progress and errors, builds a report object per operator,
// and sets the overall result using the provided CertificationStatusValidator.
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

// testHelmCertified performs the Helm chart certification check for a given operator.
//
// It accepts a Check object containing the operator's metadata, a TestEnvironment providing access to the cluster and Helm releases, and a CertificationStatusValidator used to validate the certification status. The function returns a closure that runs the actual test: it logs progress, verifies whether the operator's Helm chart is certified, records any errors, and updates the report object with the result.
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

// testContainerCertificationStatusByDigest tests the certification status of a container image identified by its digest and records the result in a report object.
//
// It receives a database check record, a test environment, and a certification status validator.
// The function logs progress, constructs report objects for each operator tested,
// calls testContainerCertification to perform the actual verification,
// and appends any errors to the report. Finally it sets the overall result of the check.
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

// testHelmVersion checks Helm chart version compliance for a certification check.
//
// It retrieves the Kubernetes client set from the environment, lists all Helm releases in the cluster,
// extracts the associated chart names, and compares them against known supported chart versions.
// For each release it creates a report object recording the result. If any release has an unsupported
// or missing chart version, the overall check is marked as failed. The function logs detailed information
// about each release, handles errors from client calls, and records the final status in the supplied Check
// object using SetResult.
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
