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

package lifecycle

import (
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/postmortem"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle/ownerreference"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle/podrecreation"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle/podsets"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle/scaling"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle/tolerations"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle/volumes"
	corev1 "k8s.io/api/core/v1"
)

const (
	timeout                    = 300 * time.Second
	timeoutPodRecreationPerPod = time.Minute
	timeoutPodSetReady         = 7 * time.Minute
	minWorkerNodesForLifecycle = 2
	statefulSet                = "StatefulSet"
	localStorage               = "local-storage"
	intrusiveTcSkippedReason   = "This is an intrusive test case and the flag --intrusive was set"
)

var (
	env provider.TestEnvironment

	beforeEachFn = func(check *checksdb.Check) error {
		env = provider.GetTestEnvironment()
		return nil
	}

	// podset = deployment or statefulset
	skipIfNoPodSetsetsUnderTest = func() (bool, string) {
		if len(env.Deployments) == 0 && len(env.StatefulSets) == 0 {
			return true, "no deployments nor statefulsets to check found"
		}
		return false, ""
	}
)

// LoadChecks registers lifecycle checks to the test suite.
//
// It sets up a before-each function, creates several check groups,
// and adds individual checks with optional skip functions.
// The returned function performs no additional work but allows
// the caller to defer any finalization if needed.
func LoadChecks() {
	log.Debug("Loading %s suite checks", common.LifecycleTestKey)

	checksGroup := checksdb.NewChecksGroup(common.LifecycleTestKey).
		WithBeforeEachFn(beforeEachFn)

	// Prestop test
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestContainerPrestopIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainersPreStop(c, &env)
			return nil
		}))

	// Scale CRD test
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestCrdScalingIdentifier)).
		WithSkipCheckFn(
			testhelper.GetNoCrdsUnderTestSkipFn(&env),
			testhelper.GetNotIntrusiveSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			// Note: We skip this test because 'testHighAvailability' in the lifecycle suite is already
			// testing the replicas and antiaffinity rules that should already be in place for crd.
			testScaleCrd(&env, timeout, c)
			return nil
		}))

	// Poststart test
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestContainerPostStartIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainersPostStart(c, &env)
			return nil
		}))

	// Image pull policy test
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestImagePullPolicyIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainersImagePolicy(c, &env)
			return nil
		}))

	// Readiness probe test
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestReadinessProbeIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainersReadinessProbe(c, &env)
			return nil
		}))

	// Liveness probe test
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestLivenessProbeIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainersLivenessProbe(c, &env)
			return nil
		}))

	// Startup probe test
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestStartupProbeIdentifier)).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainersStartupProbe(c, &env)
			return nil
		}))

	// Pod owner reference test
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestPodDeploymentBestPracticesIdentifier)).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodsOwnerReference(c, &env)
			return nil
		}))

	// High availability test
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestPodHighAvailabilityBestPractices)).
		WithSkipCheckFn(testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle)).
		WithSkipCheckFn(skipIfNoPodSetsetsUnderTest).
		WithCheckFn(func(c *checksdb.Check) error {
			testHighAvailability(c, &env)
			return nil
		}))

	// Selector and affinity best practices test
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestPodNodeSelectorAndAffinityBestPractices)).
		WithSkipCheckFn(
			testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle),
			testhelper.GetPodsWithoutAffinityRequiredLabelSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodNodeSelectorAndAffinityBestPractices(env.GetPodsWithoutAffinityRequiredLabel(), c)
			return nil
		}))

	// Pod recreation test
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestPodRecreationIdentifier)).
		WithSkipCheckFn(
			testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle),
			testhelper.GetNotIntrusiveSkipFn(&env)).
		WithSkipCheckFn(skipIfNoPodSetsetsUnderTest).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodsRecreation(c, &env)
			return nil
		}))

	// Deployment scaling test
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestDeploymentScalingIdentifier)).
		WithSkipCheckFn(
			testhelper.GetNotIntrusiveSkipFn(&env),
			testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle)).
		WithSkipCheckFn(skipIfNoPodSetsetsUnderTest).
		WithCheckFn(func(c *checksdb.Check) error {
			testDeploymentScaling(&env, timeout, c)
			return nil
		}))

	// Statefulset scaling test
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestStatefulSetScalingIdentifier)).
		WithSkipCheckFn(
			testhelper.GetNotIntrusiveSkipFn(&env),
			testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle)).
		WithSkipCheckFn(skipIfNoPodSetsetsUnderTest).
		WithCheckFn(func(c *checksdb.Check) error {
			testStatefulSetScaling(&env, timeout, c)
			return nil
		}))

	// Persistent volume reclaim policy test
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestPersistentVolumeReclaimPolicyIdentifier)).
		WithSkipCheckFn(
			testhelper.GetNoPersistentVolumesSkipFn(&env),
			testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodPersistentVolumeReclaimPolicy(c, &env)
			return nil
		}))

	// CPU Isolation test
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestCPUIsolationIdentifier)).
		WithSkipCheckFn(testhelper.GetNoGuaranteedPodsWithExclusiveCPUsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testCPUIsolation(c, &env)
			return nil
		}))

	// Affinity required pods test
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestAffinityRequiredPods)).
		WithSkipCheckFn(testhelper.GetNoAffinityRequiredPodsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testAffinityRequiredPods(c, &env)
			return nil
		}))

	// Pod toleration bypass test
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestPodTolerationBypassIdentifier)).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodTolerationBypass(c, &env)
			return nil
		}))

	// Storage provisioner test
	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestStorageProvisioner)).
		WithSkipCheckFn(
			testhelper.GetNoPodsUnderTestSkipFn(&env),
			testhelper.GetNoStorageClassesSkipFn(&env),
			testhelper.GetNoPersistentVolumeClaimsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testStorageProvisioner(c, &env)
			return nil
		}))
}

// testContainersPreStop checks that container pre‑stop hooks run correctly during a lifecycle test.
//
// It receives the current check context and the test environment, performs pre‑stop
// hook validation on all relevant containers, logs progress and errors, and updates
// the check result with any failures found. The function does not return a value; it
// records its outcome through the provided *checksdb.Check interface.
func testContainersPreStop(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		check.LogInfo("Testing Container %q", cut)

		if cut.Lifecycle == nil || (cut.Lifecycle != nil && cut.Lifecycle.PreStop == nil) {
			check.LogError("Container %q does not have preStop defined", cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container does not have preStop defined", false))
		} else {
			check.LogInfo("Container %q has preStop defined", cut)
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has preStop defined", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testContainersPostStart checks the PostStart lifecycle hook of containers in a test environment.
//
// It receives a check object and a test environment, then logs information about each container,
// creates report objects for them, and sets results based on the outcome of the PostStart
// execution. The function does not return any value; it records the results directly in
// the provided check object.
// ```
func testContainersPostStart(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		check.LogInfo("Testing Container %q", cut)

		if cut.Lifecycle == nil || (cut.Lifecycle != nil && cut.Lifecycle.PostStart == nil) {
			check.LogError("Container %q does not have postStart defined", cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container does not have postStart defined", false))
		} else {
			check.LogInfo("Container %q has postStart defined", cut)
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has postStart defined."+
				"Attention: There is a known upstream bug where a pod with a still-running postStart lifecycle hook that is deleted may not be terminated even after "+
				"the terminationGracePeriod k8s bug link: kubernetes/kubernetes#116032", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testContainersImagePolicy verifies that the container image policy is applied correctly to all containers in the pod set.
//
// It receives a check and the test environment, inspects each container’s image,
// logs information or errors as needed, and records results in the report object.
// The function does not return a value; it reports success or failure through the
// check's result mechanism.
func testContainersImagePolicy(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		check.LogInfo("Testing Container %q", cut)
		if cut.ImagePullPolicy != corev1.PullIfNotPresent {
			check.LogError("Container %q is using %q as ImagePullPolicy (compliant containers must use %q)", cut, cut.ImagePullPolicy, corev1.PullIfNotPresent)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container is not using IfNotPresent as ImagePullPolicy", false))
		} else {
			check.LogInfo("Container %q is using %q as ImagePullPolicy", cut, cut.ImagePullPolicy)
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container is using IfNotPresent as ImagePullPolicy", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testContainersReadinessProbe checks the readiness probe status of all containers in the test environment.
//
// It receives a Check object and a TestEnvironment, logs progress, creates container reports,
// and sets the result on the check based on whether any containers failed their readiness probes.
func testContainersReadinessProbe(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		check.LogInfo("Testing Container %q", cut)
		if cut.ReadinessProbe == nil {
			check.LogError("Container %q does not have ReadinessProbe defined", cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container does not have ReadinessProbe defined", false))
		} else {
			check.LogInfo("Container %q has ReadinessProbe defined", cut)
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has ReadinessProbe defined", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testContainersLivenessProbe checks that all containers in the test environment have a working liveness probe.
//
// It receives a Check object to record results and a TestEnvironment that contains the target pods.
// For each pod it verifies the liveness probe configuration, logs progress, and records success or failure
// in the check result. The function does not return a value; errors are logged and reported through the
// Check interface.
func testContainersLivenessProbe(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		check.LogInfo("Testing Container %q", cut)
		if cut.LivenessProbe == nil {
			check.LogError("Container %q does not have LivenessProbe defined", cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container does not have LivenessProbe defined", false))
		} else {
			check.LogInfo("Container %q has LivenessProbe defined", cut)
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has LivenessProbe defined", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testContainersStartupProbe verifies that the startup probes of all containers in a pod are functioning correctly.
//
// testContainersStartupProbe checks each container's startup probe status and records a report.
// It receives a Check object to log results and a TestEnvironment providing access to Kubernetes resources.
// The function logs progress, gathers container reports, updates the check result, and handles any errors encountered during probe verification.
func testContainersStartupProbe(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		check.LogInfo("Testing Container %q", cut)
		if cut.StartupProbe == nil {
			check.LogError("Container %q does not have StartupProbe defined", cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container does not have StartupProbe defined", false))
		} else {
			check.LogInfo("Container %q has StartupProbe defined", cut)
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has StartupProbe defined", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodsOwnerReference verifies that pods created during the lifecycle tests correctly reference their owner objects.
//
// It receives a Check and TestEnvironment, logs test progress, constructs an expected OwnerReference,
// runs the test, collects pod reports, and sets the test result accordingly. The function does not return
// any value; it records success or failure via the Check's Result field.
func testPodsOwnerReference(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		check.LogInfo("Testing Pod %q", put)
		o := ownerreference.NewOwnerReference(put.Pod)
		o.RunTest(check.GetLogger())
		if o.GetResults() != testhelper.SUCCESS {
			check.LogError("Pod %q found with non-compliant owner reference", put)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has non-compliant owner reference", false))
		} else {
			check.LogInfo("Pod %q has compliant owner reference", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has compliant owner reference", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodNodeSelectorAndAffinityBestPractices evaluates best practice compliance for pod node selectors and affinity settings.
//
// It receives a slice of pods and a check record, iterating over each pod to determine if node selector or affinity rules are present.
// For each pod it logs information, constructs a report object, and records errors when the rules are missing.
// The function updates the provided check with the final result status.
func testPodNodeSelectorAndAffinityBestPractices(testPods []*provider.Pod, check *checksdb.Check) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range testPods {
		check.LogInfo("Testing Pod %q", put)
		compliantPod := true
		if put.HasNodeSelector() {
			check.LogError("Pod %q has a node selector. Node selector: %v", put, put.Spec.NodeSelector)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has node selector", false))
			compliantPod = false
		}
		if put.Spec.Affinity != nil && put.Spec.Affinity.NodeAffinity != nil {
			check.LogError("Pod %q has a node affinity clause. Node affinity: %v", put, put.Spec.Affinity.NodeAffinity)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has node affinity", false))
			compliantPod = false
		}

		if compliantPod {
			check.LogInfo("Pod %q has no node selector or affinity", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has no node selector or affinity", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// nameInDeploymentSkipList determines if a deployment should be skipped based on its name and namespace.
//
// It takes the deployment's namespace, name, and a slice of SkipScalingTestDeploymentsInfo structs,
// checks whether the deployment matches any entry in the skip list, and returns true if it should be
// excluded from scaling tests. The function does not modify its inputs.
func nameInDeploymentSkipList(name, namespace string, list []configuration.SkipScalingTestDeploymentsInfo) bool {
	for _, l := range list {
		if name == l.Name && namespace == l.Namespace {
			return true
		}
	}
	return false
}

// nameInStatefulSetSkipList checks whether a stateful set should be skipped during scaling tests.
//
// It takes the namespace and name of a StatefulSet, along with a slice of SkipScalingTestStatefulSetsInfo
// that contains patterns or exact names to skip. The function returns true if the given StatefulSet matches
// any entry in the skip list, indicating it should not be included in the scaling test run.
// Otherwise it returns false.
func nameInStatefulSetSkipList(name, namespace string, list []configuration.SkipScalingTestStatefulSetsInfo) bool {
	for _, l := range list {
		if name == l.Name && namespace == l.Namespace {
			return true
		}
	}
	return false
}

// testDeploymentScaling tests scaling behavior of a deployment or HPA-controlled deployment.
//
// It receives a TestEnvironment, a timeout duration, and a Check record,
// performs scaling operations (directly or via an HPA), logs progress,
// collects reports on success or failure, and updates the Check with results.
// The function does not return a value but sets state through the provided
// environment and check objects.
func testDeploymentScaling(env *provider.TestEnvironment, timeout time.Duration, check *checksdb.Check) {
	defer env.SetNeedsRefresh()
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, deployment := range env.Deployments {
		check.LogInfo("Testing Deployment %q", deployment.ToString())
		if scaling.IsManaged(deployment.Name, env.Config.ManagedDeployments) {
			if !scaling.CheckOwnerReference(deployment.GetOwnerReferences(), env.Config.CrdFilters, env.Crds) {
				check.LogError("Deployment %q scaling failed due to OwnerReferences that are not scalable", deployment.ToString())
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewDeploymentReportObject(deployment.Namespace, deployment.Name, "Deployment is not scalable", false))
			} else {
				check.LogInfo("Deployment %q scaling skipped due to scalable OwnerReferences, test will run on the CR scaling", deployment.ToString())
			}
			continue
		}
		// Skip deployment if it is allowed by config
		if nameInDeploymentSkipList(deployment.Name, deployment.Namespace, env.Config.SkipScalingTestDeployments) {
			check.LogInfo("Deployment %q is being skipped due to configuration setting", deployment.ToString())
			continue
		}

		// TestDeploymentScaling test scaling of deployment
		// This is the entry point for deployment scaling tests
		ns, name := deployment.Namespace, deployment.Name
		if hpa := scaling.GetResourceHPA(env.HorizontalScaler, name, ns, "Deployment"); hpa != nil {
			// if the deployment is controller by
			// horizontal scaler, then test that scaler
			// can scale the deployment
			if !scaling.TestScaleHpaDeployment(deployment, hpa, timeout, check.GetLogger()) {
				check.LogError("Deployment %q has failed the HPA scale test", deployment.ToString())
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewDeploymentReportObject(deployment.Namespace, deployment.Name, "Deployment has failed the HPA scale test", false))
			}
			continue
		}
		// if the deployment is not controller by HPA
		// scale it directly
		if !scaling.TestScaleDeployment(deployment.Deployment, timeout, check.GetLogger()) {
			check.LogError("Deployment %q has failed the non-HPA scale test", deployment.ToString())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewDeploymentReportObject(deployment.Namespace, deployment.Name, "Deployment has failed the non-HPA scale test", false))
		} else {
			check.LogInfo("Deployment %q is scalable", deployment.ToString())
			compliantObjects = append(compliantObjects, testhelper.NewDeploymentReportObject(deployment.Namespace, deployment.Name, "Deployment is scalable", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testScaleCrd runs a scaling test for a custom resource definition.
//
// It accepts a test environment, a timeout duration, and a check record,
// and returns a function that performs the test when executed.
// The returned function logs progress, sets up necessary refresh flags,
// retrieves horizontal pod autoscaling information for the CRD,
// and then calls TestScaleHPACrd to validate scaling behavior.
// Results are recorded in a CrdReportObject and logged appropriately.
func testScaleCrd(env *provider.TestEnvironment, timeout time.Duration, check *checksdb.Check) {
	defer env.SetNeedsRefresh()
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for i := range env.ScaleCrUnderTest {
		groupResourceSchema := env.ScaleCrUnderTest[i].GroupResourceSchema
		scaleCr := env.ScaleCrUnderTest[i].Scale
		if hpa := scaling.GetResourceHPA(env.HorizontalScaler, scaleCr.Name, scaleCr.Namespace, scaleCr.Kind); hpa != nil {
			if !scaling.TestScaleHPACrd(&scaleCr, hpa, groupResourceSchema, timeout, check.GetLogger()) {
				check.LogError("CR has failed the scaling test: %s", scaleCr.GetName())
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewCrdReportObject(scaleCr.Namespace, scaleCr.Name, "cr has failed the HPA scaling test", false))
			}
			continue
		}
		if !scaling.TestScaleCrd(&scaleCr, groupResourceSchema, timeout, check.GetLogger()) {
			check.LogError("CR has failed the non-HPA scale test: %s", scaleCr.GetName())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewCrdReportObject(scaleCr.Namespace, scaleCr.Name, "CR has failed the non-HPA scale test", false))
		} else {
			check.LogInfo("CR is scalable")
			compliantObjects = append(compliantObjects, testhelper.NewCrdReportObject(scaleCr.Namespace, scaleCr.Name, "CR is scalable", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testStatefulSetScaling(env *provider.TestEnvironment, timeout time.Duration, check *checksdb.Check)() {
// testStatefulSetScaling verifies that a StatefulSet scales correctly in response to HPA or manual scaling operations.
// It performs the following steps:
//   1. Sets up a report object for the StatefulSet and logs initial state.
//   2. Checks if the StatefulSet is managed and has valid owner references; if not, it marks the test as needing refresh.
//   3. If the StatefulSet is in a skip list, it records that the scaling check was skipped.
//   4. Retrieves the HPA associated with the StatefulSet (if any) and tests scaling via HPA.
//   5. Tests manual scaling of the StatefulSet directly.
//   6. Records success or failure in the report object and updates the test result accordingly.
// The function accepts a test environment, a timeout duration for operations, and a check record to store results.
func testStatefulSetScaling(env *provider.TestEnvironment, timeout time.Duration, check *checksdb.Check) {
	defer env.SetNeedsRefresh()
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, statefulSet := range env.StatefulSets {
		check.LogInfo("Testing StatefulSet %q", statefulSet.ToString())
		if scaling.IsManaged(statefulSet.Name, env.Config.ManagedStatefulsets) {
			if !scaling.CheckOwnerReference(statefulSet.GetOwnerReferences(), env.Config.CrdFilters, env.Crds) {
				check.LogError("StatefulSet %q scaling failed due to OwnerReferences that are not scalable", statefulSet.ToString())
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewStatefulSetReportObject(statefulSet.Namespace, statefulSet.Name, "StatefulSet has OwnerReferences that are not scalable", false))
			} else {
				check.LogInfo("StatefulSet %q scaling skipped due to scalable OwnerReferences, test will run on te CR scaling", statefulSet.ToString())
			}
			continue
		}
		// Skip statefulset if it is allowed by config
		if nameInStatefulSetSkipList(statefulSet.Name, statefulSet.Namespace, env.Config.SkipScalingTestStatefulSets) {
			check.LogInfo("StatefulSet %q is being skipped due to configuration setting", statefulSet.String())
			continue
		}

		// TeststatefulsetScaling test scaling of statefulset
		// This is the entry point for statefulset scaling tests
		ns, name := statefulSet.Namespace, statefulSet.Name
		if hpa := scaling.GetResourceHPA(env.HorizontalScaler, name, ns, "StatefulSet"); hpa != nil {
			// if the statefulset is controller by
			// horizontal scaler, then test that scaler
			// can scale the statefulset
			if !scaling.TestScaleHpaStatefulSet(statefulSet.StatefulSet, hpa, timeout, check.GetLogger()) {
				check.LogError("StatefulSet has failed the scaling test: %q", statefulSet.ToString())
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewStatefulSetReportObject(statefulSet.Namespace, statefulSet.Name, "StatefulSet has failed the HPA scaling test", false))
			}
			continue
		}
		// if the statefulset is not controller by HPA
		// scale it directly
		if !scaling.TestScaleStatefulSet(statefulSet.StatefulSet, timeout, check.GetLogger()) {
			check.LogError("StatefulSet has failed the scaling test: %s", statefulSet.ToString())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewStatefulSetReportObject(statefulSet.Namespace, statefulSet.Name, "StatefulSet has failed the non-HPA scale test", false))
		} else {
			check.LogInfo("StatefulSet is scalable")
			compliantObjects = append(compliantObjects, testhelper.NewStatefulSetReportObject(statefulSet.Namespace, statefulSet.Name, "StatefulSet is scalable", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testHighAvailability verifies that the system remains available when multiple replicas are running concurrently.
//
// testHighAvailability checks that a high‑availability deployment continues to respond correctly
// while pods are being recreated or scaled. It logs status changes, collects deployment and
// stateful set reports, and records any errors encountered during the process. The function
// takes a Check object for recording results and a TestEnvironment providing access to the
// cluster state. No value is returned; the outcome is stored in the Check via SetResult.
func testHighAvailability(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, dp := range env.Deployments {
		check.LogInfo("Testing Deployment %q", dp.ToString())
		if dp.Spec.Replicas == nil || *(dp.Spec.Replicas) <= 1 {
			check.LogError("Deployment %q found without valid high availability (number of replicas must be greater than 1)", dp.ToString())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewDeploymentReportObject(dp.Namespace, dp.Name, "Deployment found without valid high availability", false))
			continue
		}

		// Skip any AffinityRequired pods
		//nolint:goconst
		if dp.Spec.Template.Labels["AffinityRequired"] == "true" {
			check.LogInfo("Skipping Deployment %q with affinity required", dp.ToString())
			continue
		}

		if dp.Spec.Template.Spec.Affinity == nil ||
			dp.Spec.Template.Spec.Affinity.PodAntiAffinity == nil {
			check.LogError("Deployment %q found without valid high availability (PodAntiAffinity must be defined)", dp.ToString())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewDeploymentReportObject(dp.Namespace, dp.Name, "Deployment found without valid high availability", false))
		} else {
			check.LogInfo("Deployment %q has valid high availability", dp.ToString())
			compliantObjects = append(compliantObjects, testhelper.NewDeploymentReportObject(dp.Namespace, dp.Name, "Deployment has valid high availability", true))
		}
	}
	for _, st := range env.StatefulSets {
		if st.Spec.Replicas == nil || *(st.Spec.Replicas) <= 1 {
			check.LogError("StatefulSet %q found without valid high availability (number of replicas must be greater than 1)", st.ToString())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewStatefulSetReportObject(st.Namespace, st.Name, "StatefulSet found without valid high availability", false))
			continue
		}

		// Skip any AffinityRequired pods
		if st.Spec.Template.Labels["AffinityRequired"] == "true" {
			check.LogInfo("Skipping StatefulSet %q with affinity required", st.ToString())
			continue
		}

		if st.Spec.Template.Spec.Affinity == nil ||
			st.Spec.Template.Spec.Affinity.PodAntiAffinity == nil {
			check.LogError("StatefulSet %q found without valid high availability (PodAntiAffinity must be defined)", st.ToString())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewStatefulSetReportObject(st.Namespace, st.Name, "StatefulSet found without valid high availability", false))
		} else {
			check.LogInfo("StatefulSet %q has valid high availability", st.ToString())
			compliantObjects = append(compliantObjects, testhelper.NewStatefulSetReportObject(st.Namespace, st.Name, "StatefulSet has valid high availability", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodsRecreation verifies that pods belonging to deployments and statefulsets are re‑created and become ready after a node loss.
//
// It accepts a checksdb.Check pointer and a TestEnvironment pointer, performs a node cordon and cleanup,
// waits for all pod sets to be ready within configured timeouts, and records the results in the check.
// The function logs debug and error information, updates the check result status, and may mark the check
// as needing a refresh if any pods fail to recover. It handles both deployments and statefulsets,
// creating appropriate report objects for each pod, node, and pod set that is affected.
func testPodsRecreation(check *checksdb.Check, env *provider.TestEnvironment) { //nolint:funlen,gocyclo
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	needsPostMortemInfo := true
	defer func() {
		if needsPostMortemInfo {
			check.LogDebug("%s", postmortem.Log())
		}
		// Since we are possible exiting early, we need to make sure we set the result at the end of the function.
		check.SetResult(compliantObjects, nonCompliantObjects)
	}()
	check.LogInfo("Testing node draining effect of deployment")
	check.LogInfo("Testing initial state for deployments")
	defer env.SetNeedsRefresh()

	// Before draining any node, wait until all podsets are ready. The timeout depends on the number of podsets to check.
	// timeout = k-mins + (1min * (num-deployments + num-statefulsets))
	allPodsetsReadyTimeout := timeoutPodSetReady + time.Minute*time.Duration(len(env.Deployments)+len(env.StatefulSets))
	notReadyDeployments, notReadyStatefulSets := podsets.WaitForAllPodSetsReady(env, allPodsetsReadyTimeout, check.GetLogger())
	if len(notReadyDeployments) > 0 || len(notReadyStatefulSets) > 0 {
		for _, dep := range notReadyDeployments {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewDeploymentReportObject(dep.Namespace, dep.Name, "Deployment was not ready before draining any node.", false))
		}
		for _, sts := range notReadyStatefulSets {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewStatefulSetReportObject(sts.Namespace, sts.Name, "Statefulset was not ready before draining any node.", false))
		}
		return
	}

	// Filter out pods with Node Assignments present and FAIL them.
	// We run into problems with this test when there are nodeSelectors assigned affecting where
	// pods are scheduled.  Also, they are not allowed in general, see the node-selector test case.
	// Skip the safeguard for any pods that are using a runtimeClassName.  This is potentially
	// because pods that are derived from a performance profile might have a built-in nodeSelector.
	var podsWithNodeAssignment []*provider.Pod
	for _, put := range env.Pods {
		if !put.IsRuntimeClassNameSpecified() && put.HasNodeSelector() {
			podsWithNodeAssignment = append(podsWithNodeAssignment, put)
			check.LogError("Pod %q has been found with node selector(s): %v", put, put.Spec.NodeSelector)
		}
	}
	if len(podsWithNodeAssignment) > 0 {
		check.LogError("Pod(s) have been found to contain a node assignment and cannot perform the pod-recreation test: %v", podsWithNodeAssignment)
		for _, pod := range podsWithNodeAssignment {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(pod.Namespace, pod.Name, "Pod has node assignment.", false))
		}

		return
	}

	for nodeName := range podsets.GetAllNodesForAllPodSets(env.Pods) {
		defer podrecreation.CordonCleanup(nodeName, check) //nolint:gocritic // The defer in loop is intentional, calling the cleanup function once per node
		err := podrecreation.CordonHelper(nodeName, podrecreation.Cordon)
		if err != nil {
			check.LogError("Error cordoning the node: %s", nodeName)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(nodeName, "Node cordoning failed", false))
			return
		}
		check.LogInfo("Draining and Cordoning node %s: ", nodeName)
		count, err := podrecreation.CountPodsWithDelete(env.Pods, nodeName, podrecreation.NoDelete)
		if err != nil {
			check.LogError("Getting pods list to drain failed, err=%v", err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(nodeName, "Getting pods list to drain failed", false))
			return
		}
		nodeTimeout := timeoutPodSetReady + timeoutPodRecreationPerPod*time.Duration(count)
		check.LogDebug("Draining node: %s with timeout: %s", nodeName, nodeTimeout)
		_, err = podrecreation.CountPodsWithDelete(env.Pods, nodeName, podrecreation.DeleteForeground)
		if err != nil {
			check.LogError("Draining node %q failed, err=%v", nodeName, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(nodeName, "Draining node failed", false))
			return
		}

		notReadyDeployments, notReadyStatefulSets := podsets.WaitForAllPodSetsReady(env, nodeTimeout, check.GetLogger())
		if len(notReadyDeployments) > 0 || len(notReadyStatefulSets) > 0 {
			for _, dep := range notReadyDeployments {
				check.LogError("Deployment %q not ready after draining node %q", dep.ToString(), nodeName)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewDeploymentReportObject(dep.Namespace, dep.Name, "Deployment not ready after draining node "+nodeName, false))
			}
			for _, sts := range notReadyStatefulSets {
				check.LogError("StatefulSet %q not ready after draining node %q", sts.ToString(), nodeName)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewStatefulSetReportObject(sts.Namespace, sts.Name, "Statefulset not ready after draining node "+nodeName, false))
			}
			return
		}

		err = podrecreation.CordonHelper(nodeName, podrecreation.Uncordon)
		if err != nil {
			check.LogFatal("Error uncordoning the node: %s", nodeName)
		}
	}

	// If everything went well for all nodes, the nonCompliantObjects should be empty. We need to
	// manually add all the deps/sts into the compliant object lists so the check is marked as skipped.
	// ToDo: Improve this.
	if len(nonCompliantObjects) == 0 {
		for _, dep := range env.Deployments {
			check.LogInfo("Deployment's pods successfully re-schedulled after node draining.")
			compliantObjects = append(compliantObjects, testhelper.NewDeploymentReportObject(dep.Namespace, dep.Name, "Deployment's pods successfully re-schedulled after node draining.", true))
		}

		for _, sts := range env.StatefulSets {
			check.LogInfo("Statefulset's pods successfully re-schedulled after node draining.")
			compliantObjects = append(compliantObjects, testhelper.NewStatefulSetReportObject(sts.Namespace, sts.Name, "Statefulset's pods successfully re-schedulled after node draining.", true))
		}
	}

	needsPostMortemInfo = false
}

// testPodPersistentVolumeReclaimPolicy verifies that a pod's persistent volume reclaim policy is Delete.
//
// It checks whether the pod’s associated PVC uses the Delete reclamation policy.
// The function logs progress, creates report objects, and records the result
// in the provided check database entry. If the policy is not Delete,
// it reports an error and marks the check as failed.
func testPodPersistentVolumeReclaimPolicy(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	// Look through all of the pods, matching their persistent volumes to the list of overall cluster PVs and checking their reclaim status.
	for _, put := range env.Pods {
		check.LogInfo("Testing Pod %q", put)
		compliantPod := true
		// Loop through all of the volumes attached to the pod.
		for pvIndex := range put.Spec.Volumes {
			// Skip any volumes that do not have a PVC.  No need to test them.
			if put.Spec.Volumes[pvIndex].PersistentVolumeClaim == nil {
				check.LogInfo("Pod %q does not have a PVC", put)
				continue
			}

			// If the Pod Volume is not tied back to a PVC and corresponding PV that has a reclaim policy of DELETE.
			if !volumes.IsPodVolumeReclaimPolicyDelete(&put.Spec.Volumes[pvIndex], env.PersistentVolumes, env.PersistentVolumeClaims) {
				check.LogError("Pod %q with volume %q has been found without a reclaim policy of DELETE.", put, put.Spec.Volumes[pvIndex].Name)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod contains volume without a reclaim policy of DELETE", false).
					AddField(testhelper.PersistentVolumeName, put.Spec.Volumes[pvIndex].Name).
					AddField(testhelper.PersistentVolumeClaimName, put.Spec.Volumes[pvIndex].PersistentVolumeClaim.ClaimName))
				compliantPod = false
				break
			}
		}

		if compliantPod {
			check.LogInfo("Pod %q complies with volume reclaim policy rules", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod complies with volume reclaim policy rules", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testCPUIsolation checks CPU isolation compliance of guaranteed pods with exclusive CPUs.
//
// It retrieves all guaranteed pods that request exclusive CPUs from the environment,
// logs information about each pod, and verifies whether they are compliant
// with CPU isolation requirements using IsCPUIsolationCompliant.
// For each pod it creates a report object indicating success or failure
// and records any errors. The overall result of the check is stored in
// the provided Check object via SetResult. This function does not return
// a value; all outcomes are reported through side effects on the Check
// and the test environment's logging facilities.
func testCPUIsolation(check *checksdb.Check, env *provider.TestEnvironment) {
	// Individual requirements we are looking for:
	//  - CPU Requests and Limits must be in the form of whole units
	// - Resource Requests and Limits must be provided and identical

	// Additional checks if the above pass
	// - 'runtimeClassName' must be specified
	// - Annotations must be provided disabling CPU and IRQ load-balancing.

	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, put := range env.GetGuaranteedPodsWithExclusiveCPUs() {
		check.LogInfo("Testing Pod %q", put)
		if !put.IsCPUIsolationCompliant() {
			check.LogError("Pod %q is not CPU isolated", put)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is not CPU isolated", false))
		} else {
			check.LogInfo("Pod %q is CPU isolated", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is CPU isolated", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testAffinityRequiredPods verifies that pods required by affinity rules are scheduled correctly.
//
// It retrieves the list of pods that must be present according to the pod's
// affinity specification, checks each against the current cluster state,
// and logs information about compliance or errors.
// For each pod it creates a report object and records the result in the check.
func testAffinityRequiredPods(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.GetAffinityRequiredPods() {
		check.LogInfo("Testing Pod %q", put)
		// Check if the pod is Affinity compliant.
		result, err := put.IsAffinityCompliant()
		if !result {
			check.LogError("Pod %q is not Affinity compliant, reason=%v", put, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is not Affinity compliant", false))
		} else {
			check.LogInfo("Pod %q is Affinity compliant", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is Affinity compliant", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodTolerationBypass checks that pods can bypass node tolerations when the
// cluster is configured to allow such behavior.
//
// It receives a check definition and a test environment, logs diagnostic
// information, and determines whether pod toleration modifications were made.
// The function creates pod report objects for both success and failure cases,
// sets the result status accordingly, and records relevant fields in the
// report. If any error occurs during validation it is logged and marked as
// an error in the pod report.
func testPodTolerationBypass(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, put := range env.Pods {
		check.LogInfo("Testing Pod %q", put)
		podIsCompliant := true
		for _, t := range put.Spec.Tolerations {
			// Check if the tolerations fall outside the 'default' and are modified versions
			// Take also into account the qosClass applied to the pod
			if tolerations.IsTolerationModified(t, put.Status.QOSClass) {
				check.LogError("Pod %q has been found with non-default toleration %s/%s which is not allowed.", put, t.Key, t.Effect)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has non-default toleration", false).
					AddField(testhelper.TolerationKey, t.Key).
					AddField(testhelper.TolerationEffect, string(t.Effect)))
				podIsCompliant = false
			}
		}

		if podIsCompliant {
			check.LogInfo("Pod %q has default toleration", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has default toleration", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testStorageProvisioner checks that storage provisioners in the cluster can create and delete persistent volumes correctly.
//
// testStorageProvisioner verifies that the storage provisioners configured in the
// Kubernetes cluster are able to create a PersistentVolumeClaim, bind it to a
// PersistentVolume, and subsequently clean up the claim. It logs progress,
// collects pod reports for any provisioning pods, and records success or
// failure in the provided checksdb.Check object. The function does not return
// a value; all results are recorded through the Check object's SetResult
// method.
func testStorageProvisioner(check *checksdb.Check, env *provider.TestEnvironment) {
	const localStorageProvisioner = "kubernetes.io/no-provisioner"
	const lvmProvisioner = "topolvm.io"
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	var StorageClasses = env.StorageClassList
	var Pvc = env.PersistentVolumeClaims
	snoSingleLocalStorageProvisionner := ""
	for _, put := range env.Pods {
		check.LogInfo("Testing Pod %q", put)
		usesPvcAndStorageClass := false
		for pvIndex := range put.Spec.Volumes {
			// Skip any nil persistentClaims.
			volume := put.Spec.Volumes[pvIndex]
			if volume.PersistentVolumeClaim == nil {
				continue
			}
			// We have the list of pods/volumes/claims.
			// Look through the storageClass list for a match.
			for i := range Pvc {
				if Pvc[i].Name == put.Spec.Volumes[pvIndex].PersistentVolumeClaim.ClaimName && Pvc[i].Namespace == put.Namespace {
					for j := range StorageClasses {
						if Pvc[i].Spec.StorageClassName != nil && StorageClasses[j].Name == *Pvc[i].Spec.StorageClassName {
							usesPvcAndStorageClass = true
							check.LogDebug("Pod %q pvc_name: %s, storageclass_name: %s, provisioner_name: %s", put, put.Spec.Volumes[pvIndex].PersistentVolumeClaim.ClaimName,
								StorageClasses[j].Name, StorageClasses[j].Provisioner)

							if env.IsSNO() {
								// For SNO, only one local storage provisionner is allowed. The first local storage provisioner for this pod is assumed to be the only local storage provisioner allowed in the cluster.
								if snoSingleLocalStorageProvisionner == "" &&
									(StorageClasses[j].Provisioner == localStorageProvisioner ||
										StorageClasses[j].Provisioner == lvmProvisioner) {
									snoSingleLocalStorageProvisionner = StorageClasses[j].Provisioner
								}
								if StorageClasses[j].Provisioner == snoSingleLocalStorageProvisionner {
									check.LogInfo("Pod %q: Local storage (no provisioner or lvms) is recommended for SNO clusters.", put)
									compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Local storage (no provisioner or lvms) is recommended for SNO clusters.", false).
										AddField(testhelper.StorageClassName, StorageClasses[j].Name).
										AddField(testhelper.StorageClassProvisioner, StorageClasses[j].Provisioner).
										AddField(testhelper.PersistentVolumeClaimName, put.Spec.Volumes[pvIndex].PersistentVolumeClaim.ClaimName))
									continue
								}
								if StorageClasses[j].Provisioner == localStorageProvisioner || StorageClasses[j].Provisioner == lvmProvisioner {
									check.LogError("Pod %q: A single type of local storage cluster is recommended for single node clusters. Use lvms or kubernetes noprovisioner, but not both.", put)
									nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name,
										"A single type of local storage cluster is recommended for single node clusters. Use lvms or kubernetes noprovisioner, but not both.", false).
										AddField(testhelper.StorageClassName, StorageClasses[j].Name).
										AddField(testhelper.StorageClassProvisioner, StorageClasses[j].Provisioner).
										AddField(testhelper.PersistentVolumeClaimName, put.Spec.Volumes[pvIndex].PersistentVolumeClaim.ClaimName))
									continue
								}
								check.LogError("Pod %q: Non local storage not recommended in single node clusters.", put)
								nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Non local storage not recommended in single node clusters.", false).
									AddField(testhelper.StorageClassName, StorageClasses[j].Name).
									AddField(testhelper.StorageClassProvisioner, StorageClasses[j].Provisioner).
									AddField(testhelper.PersistentVolumeClaimName, put.Spec.Volumes[pvIndex].PersistentVolumeClaim.ClaimName))
							} else {
								if StorageClasses[j].Provisioner == localStorageProvisioner || StorageClasses[j].Provisioner == lvmProvisioner {
									check.LogError("Pod %q: Local storage provisioner (no provisioner or lvms) not recommended in multinode clusters.", put)
									nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Local storage provisioner (no provisioner or lvms) not recommended in multinode clusters.", false).
										AddField(testhelper.StorageClassName, StorageClasses[j].Name).
										AddField(testhelper.StorageClassProvisioner, StorageClasses[j].Provisioner).
										AddField(testhelper.PersistentVolumeClaimName, put.Spec.Volumes[pvIndex].PersistentVolumeClaim.ClaimName))
									continue
								}
								check.LogInfo("Pod %q: Non local storage provisioner recommended in multinode clusters.", put)
								compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Non local storage provisioner recommended in multinode clusters.", false).
									AddField(testhelper.StorageClassName, StorageClasses[j].Name).
									AddField(testhelper.StorageClassProvisioner, StorageClasses[j].Provisioner).
									AddField(testhelper.PersistentVolumeClaimName, put.Spec.Volumes[pvIndex].PersistentVolumeClaim.ClaimName))
							}
						}
					}
				}
			}
			// Save as compliant pod in case it's not using any of the existing PVC/StorageClasses of the cluster.
			// Otherwise, in this cases the check will be marked as skipped.
			// ToDo: improve this function.
			if !usesPvcAndStorageClass {
				check.LogInfo("Pod %q not configured to use local storage", put)
				compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod not configured to use local storage.", true))
			}
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}
