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

package lifecycle

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/ownerreference"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/podrecreation"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/podsets"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/scaling"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/tolerations"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/volumes"
	"github.com/test-network-function/cnf-certification-test/pkg/checksdb"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"github.com/test-network-function/cnf-certification-test/pkg/postmortem"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
	corev1 "k8s.io/api/core/v1"
)

const (
	timeout                    = 300 * time.Second
	timeoutPodRecreationPerPod = time.Minute
	timeoutPodSetReady         = 7 * time.Minute
	minWorkerNodesForLifecycle = 2
	statefulSet                = "StatefulSet"
	localStorage               = "local-storage"
	intrusiveTcSkippedReason   = "This is an intrusive test case and the env var TNF_NON_INTRUSIVE_ONLY was set"
)

var (
	env provider.TestEnvironment

	beforeEachFn = func(check *checksdb.Check) error {
		logrus.Infof("Check %s: getting test environment.", check.ID)
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

//nolint:funlen
func init() {
	logrus.Debugf("Entering %s suite", common.LifecycleTestKey)

	checksGroup := checksdb.NewChecksGroup(common.LifecycleTestKey).
		WithBeforeEachFn(beforeEachFn)

	// Prestop test
	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestShutdownIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainersPreStop(c, &env)
			return nil
		}))

	// Scale CRD test
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestCrdScalingIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
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
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestStartupIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainersPostStart(c, &env)
			return nil
		}))

	// Image pull policy test
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestImagePullPolicyIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainersImagePolicy(c, &env)
			return nil
		}))

	// Readiness probe test
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestReadinessProbeIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainersReadinessProbe(c, &env)
			return nil
		}))

	// Liveness probe test
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestLivenessProbeIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainersLivenessProbe(c, &env)
			return nil
		}))

	// Startup probe test
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestStartupProbeIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testContainersStartupProbe(c, &env)
			return nil
		}))

	// Pod owner reference test
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodDeploymentBestPracticesIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodsOwnerReference(c, &env)
			return nil
		}))

	// High availability test
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodHighAvailabilityBestPractices)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle)).
		WithSkipCheckFn(skipIfNoPodSetsetsUnderTest).
		WithCheckFn(func(c *checksdb.Check) error {
			testHighAvailability(c, &env)
			return nil
		}))

	// Selector and affinity best practices test
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodNodeSelectorAndAffinityBestPractices)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(
			testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle),
			testhelper.GetPodsWithoutAffinityRequiredLabelSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodNodeSelectorAndAffinityBestPractices(env.GetPodsWithoutAffinityRequiredLabel(), c)
			return nil
		}))

	// Pod recreation test
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodRecreationIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(
			testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle),
			testhelper.GetNotIntrusiveSkipFn(&env)).
		WithSkipCheckFn(skipIfNoPodSetsetsUnderTest).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodsRecreation(c, &env)
			return nil
		}))

	// Deployment scaling test
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestDeploymentScalingIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(
			testhelper.GetNotIntrusiveSkipFn(&env),
			testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle)).
		WithSkipCheckFn(skipIfNoPodSetsetsUnderTest).
		WithCheckFn(func(c *checksdb.Check) error {
			testDeploymentScaling(&env, timeout, c)
			return nil
		}))

	// Statefulset scaling test
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestStateFulSetScalingIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(
			testhelper.GetNotIntrusiveSkipFn(&env),
			testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle)).
		WithSkipCheckFn(skipIfNoPodSetsetsUnderTest).
		WithCheckFn(func(c *checksdb.Check) error {
			testStatefulSetScaling(&env, timeout, c)
			return nil
		}))

	// Persistent volume reclaim policy test
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPersistentVolumeReclaimPolicyIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(
			testhelper.GetNoPersistentVolumesSkipFn(&env),
			testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodPersistentVolumeReclaimPolicy(c, &env)
			return nil
		}))

	// CPU Isolation test
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestCPUIsolationIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoGuaranteedPodsWithExclusiveCPUsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testCPUIsolation(c, &env)
			return nil
		}))

	// Affinity required pods test
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestAffinityRequiredPods)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoAffinityRequiredPodsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testAffinityRequiredPods(c, &env)
			return nil
		}))

	// Pod toleration bypass test
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodTolerationBypassIdentifier)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testPodTolerationBypass(c, &env)
			return nil
		}))

	// Storage provisioner test
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestStorageProvisioner)
	checksGroup.Add(checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(
			testhelper.GetNoPodsUnderTestSkipFn(&env),
			testhelper.GetNoStorageClassesSkipFn(&env),
			testhelper.GetNoPersistentVolumeClaimsSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testStorageProvisioner(c, &env)
			return nil
		}))
}

func testContainersPreStop(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " pre stop lifecycle ")

		if cut.Lifecycle == nil || (cut.Lifecycle != nil && cut.Lifecycle.PreStop == nil) {
			tnf.ClaimFilePrintf("%s does not have preStop defined", cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container does not have preStop defined", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has preStop defined", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testContainersPostStart(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " post start lifecycle ")

		if cut.Lifecycle == nil || (cut.Lifecycle != nil && cut.Lifecycle.PostStart == nil) {
			tnf.ClaimFilePrintf("%s does not have postStart defined", cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container does not have postStart defined", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has postStart defined", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testContainersImagePolicy(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " pull policy, should be ", corev1.PullIfNotPresent)
		if cut.ImagePullPolicy != corev1.PullIfNotPresent {
			tnf.Logf(logrus.WarnLevel, "%s is using %s as ImagePullPolicy", cut, cut.ImagePullPolicy)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container is not using IfNotPresent as ImagePullPolicy", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container is using IfNotPresent as ImagePullPolicy", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testContainersReadinessProbe(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " readiness probe ")
		if cut.ReadinessProbe == nil {
			tnf.Logf(logrus.WarnLevel, "%s does not have ReadinessProbe defined", cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container does not have ReadinessProbe defined", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has ReadinessProbe defined", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testContainersLivenessProbe(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " liveness probe ")
		if cut.LivenessProbe == nil {
			tnf.Logf(logrus.WarnLevel, "%s does not have LivenessProbe defined", cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container does not have LivenessProbe defined", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has LivenessProbe defined", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testContainersStartupProbe(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " startup probe ")
		if cut.StartupProbe == nil {
			tnf.Logf(logrus.WarnLevel, "%s does not have StartupProbe defined", cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container does not have StartupProbe defined", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has StartupProbe defined", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testPodsOwnerReference(check *checksdb.Check, env *provider.TestEnvironment) {
	tnf.Logf(logrus.InfoLevel, "Testing owners of CNF pod, should be replicas Set")
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.Pods {
		logrus.Debugln("check pod ", put.Namespace, " ", put.Name, " owner reference")
		o := ownerreference.NewOwnerReference(put.Pod)
		o.RunTest()
		if o.GetResults() != testhelper.SUCCESS {
			tnf.ClaimFilePrintf("%s found with non-compliant owner reference", put.String())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has non-compliant owner reference", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has compliant owner reference", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testPodNodeSelectorAndAffinityBestPractices(testPods []*provider.Pod, check *checksdb.Check) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range testPods {
		compliantPod := true
		if put.HasNodeSelector() {
			tnf.ClaimFilePrintf("ERROR: %s has a node selector. Node selector: %v", put, put.Spec.NodeSelector)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has node selector", false))
			compliantPod = false
		}
		if put.Spec.Affinity != nil && put.Spec.Affinity.NodeAffinity != nil {
			tnf.ClaimFilePrintf("ERROR: %s has a node affinity clause. Node affinity: %v", put, put.Spec.Affinity.NodeAffinity)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has node affinity", false))
			compliantPod = false
		}

		if compliantPod {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has no node selector or affinity", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

func nameInDeploymentSkipList(name, namespace string, list []configuration.SkipScalingTestDeploymentsInfo) bool {
	for _, l := range list {
		if name == l.Name && namespace == l.Namespace {
			return true
		}
	}
	return false
}

func nameInStatefulSetSkipList(name, namespace string, list []configuration.SkipScalingTestStatefulSetsInfo) bool {
	for _, l := range list {
		if name == l.Name && namespace == l.Namespace {
			return true
		}
	}
	return false
}

//nolint:dupl
func testDeploymentScaling(env *provider.TestEnvironment, timeout time.Duration, check *checksdb.Check) {
	tnf.Logf(logrus.InfoLevel, "Testing deployment scaling")
	defer env.SetNeedsRefresh()
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for i := range env.Deployments {
		if scaling.IsManaged(env.Deployments[i].Name, env.Config.ManagedDeployments) {
			if !scaling.CheckOwnerReference(env.Deployments[i].GetOwnerReferences(), env.Config.CrdFilters, env.Crds) {
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewDeploymentReportObject(env.Deployments[i].Namespace, env.Deployments[i].Name, "Deployment is not scalable", false))
				tnf.ClaimFilePrintf("%s is scaling failed due to OwnerReferences that are not scalable", env.Deployments[i].ToString())
			} else {
				logrus.Infof("%s is scaling skipped due to scalable OwnerReferences, test will run on the cr scaling", env.Deployments[i].ToString())
			}
			continue
		}
		// Skip deployment if it is allowed by config
		if nameInDeploymentSkipList(env.Deployments[i].Name, env.Deployments[i].Namespace, env.Config.SkipScalingTestDeployments) {
			tnf.ClaimFilePrintf("%s is being skipped due to configuration setting", env.Deployments[i].String())
			continue
		}

		// TestDeploymentScaling test scaling of deployment
		// This is the entry point for deployment scaling tests
		ns, name := env.Deployments[i].Namespace, env.Deployments[i].Name
		if hpa := scaling.GetResourceHPA(env.HorizontalScaler, name, ns, "Deployment"); hpa != nil {
			// if the deployment is controller by
			// horizontal scaler, then test that scaler
			// can scale the deployment
			if !scaling.TestScaleHpaDeployment(env.Deployments[i], hpa, timeout) {
				tnf.ClaimFilePrintf("Deployment has failed the HPA scale test: %s", env.Deployments[i].ToString())
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewDeploymentReportObject(env.Deployments[i].Namespace, env.Deployments[i].Name, "Deployment has failed the HPA scale test", false))
			}
			continue
		}
		// if the deployment is not controller by HPA
		// scale it directly
		if !scaling.TestScaleDeployment(env.Deployments[i].Deployment, timeout) {
			tnf.ClaimFilePrintf("Deployment has failed the non-HPA scale test: %s", env.Deployments[i].ToString())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewDeploymentReportObject(env.Deployments[i].Namespace, env.Deployments[i].Name, "Deployment has failed the non-HPA scale test", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewDeploymentReportObject(env.Deployments[i].Namespace, env.Deployments[i].Name, "Deployment is scalable", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testScaleCrd(env *provider.TestEnvironment, timeout time.Duration, check *checksdb.Check) {
	tnf.Logf(logrus.InfoLevel, "Testing custom resource scaling")
	defer env.SetNeedsRefresh()
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for i := range env.ScaleCrUnderTest {
		groupResourceSchema := env.ScaleCrUnderTest[i].GroupResourceSchema
		scaleCr := env.ScaleCrUnderTest[i].Scale
		if hpa := scaling.GetResourceHPA(env.HorizontalScaler, scaleCr.Name, scaleCr.Namespace, scaleCr.Kind); hpa != nil {
			if !scaling.TestScaleHPACrd(&scaleCr, hpa, groupResourceSchema, timeout) {
				tnf.ClaimFilePrintf("cr has failed the scaling test: %s", scaleCr.GetName())
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewCrdReportObject(scaleCr.Namespace, scaleCr.Name, "cr has failed the HPA scaling test", false))
			}
			continue
		}
		if !scaling.TestScaleCrd(&scaleCr, groupResourceSchema, timeout) {
			tnf.ClaimFilePrintf("CR has failed the non-HPA scale test: %s", scaleCr.GetName())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewCrdReportObject(scaleCr.Namespace, scaleCr.Name, "CR has failed the non-HPA scale test", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewCrdReportObject(scaleCr.Namespace, scaleCr.Name, "CR is scalable", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

//nolint:dupl
func testStatefulSetScaling(env *provider.TestEnvironment, timeout time.Duration, check *checksdb.Check) {
	tnf.Logf(logrus.InfoLevel, "Testing statefulset scaling")
	defer env.SetNeedsRefresh()
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for i := range env.StatefulSets {
		if scaling.IsManaged(env.StatefulSets[i].Name, env.Config.ManagedStatefulsets) {
			if !scaling.CheckOwnerReference(env.StatefulSets[i].GetOwnerReferences(), env.Config.CrdFilters, env.Crds) {
				tnf.ClaimFilePrintf("%s is scaling failed due to OwnerReferences that are not scalable", env.Deployments[i].ToString())
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewStatefulSetReportObject(env.StatefulSets[i].Namespace, env.StatefulSets[i].Name, "StatefulSet has OwnerReferences that are not scalable", false))
			} else {
				logrus.Infof("%s is scaling skipped due to scalable OwnerReferences, test will run on te cr scaling", env.StatefulSets[i].ToString())
			}
			continue
		}
		// Skip statefulset if it is allowed by config
		if nameInStatefulSetSkipList(env.StatefulSets[i].Name, env.StatefulSets[i].Namespace, env.Config.SkipScalingTestStatefulSets) {
			tnf.ClaimFilePrintf("%s is being skipped due to configuration setting", env.StatefulSets[i].String())
			continue
		}

		// TeststatefulsetScaling test scaling of statefulset
		// This is the entry point for statefulset scaling tests
		ns, name := env.StatefulSets[i].Namespace, env.StatefulSets[i].Name
		if hpa := scaling.GetResourceHPA(env.HorizontalScaler, name, ns, "StatefulSet"); hpa != nil {
			// if the statefulset is controller by
			// horizontal scaler, then test that scaler
			// can scale the statefulset
			if !scaling.TestScaleHpaStatefulSet(env.StatefulSets[i].StatefulSet, hpa, timeout) {
				tnf.ClaimFilePrintf("StatefulSet has failed the scaling test: %s", env.StatefulSets[i].ToString())
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewStatefulSetReportObject(env.StatefulSets[i].Namespace, env.StatefulSets[i].Name, "StatefulSet has failed the HPA scaling test", false))
			}
			continue
		}
		// if the statefulset is not controller by HPA
		// scale it directly
		if !scaling.TestScaleStatefulSet(env.StatefulSets[i].StatefulSet, timeout) {
			tnf.ClaimFilePrintf("StatefulSet has failed the scaling test: %s", env.StatefulSets[i].ToString())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewStatefulSetReportObject(env.StatefulSets[i].Namespace, env.StatefulSets[i].Name, "StatefulSet has failed the non-HPA scale test", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewStatefulSetReportObject(env.StatefulSets[i].Namespace, env.StatefulSets[i].Name, "StatefulSet is scalable", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testHighAvailability
func testHighAvailability(check *checksdb.Check, env *provider.TestEnvironment) {
	tnf.Logf(logrus.InfoLevel, "Should set pod replica number greater than 1")

	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, dp := range env.Deployments {
		if dp.Spec.Replicas == nil || *(dp.Spec.Replicas) <= 1 {
			tnf.ClaimFilePrintf("Deployment found without valid high availability: %s", dp.ToString())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewDeploymentReportObject(dp.Namespace, dp.Name, "Deployment found without valid high availability", false))
			continue
		}

		// Skip any AffinityRequired pods
		//nolint:goconst
		if dp.Spec.Template.Labels["AffinityRequired"] == "true" {
			continue
		}

		if dp.Spec.Template.Spec.Affinity == nil ||
			dp.Spec.Template.Spec.Affinity.PodAntiAffinity == nil {
			tnf.ClaimFilePrintf("Deployment found without valid high availability: %s", dp.ToString())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewDeploymentReportObject(dp.Namespace, dp.Name, "Deployment found without valid high availability", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewDeploymentReportObject(dp.Namespace, dp.Name, "Deployment has valid high availability", true))
		}
	}
	for _, st := range env.StatefulSets {
		if st.Spec.Replicas == nil || *(st.Spec.Replicas) <= 1 {
			tnf.ClaimFilePrintf("StatefulSet found without valid high availability: %s", st.ToString())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewStatefulSetReportObject(st.Namespace, st.Name, "StatefulSet found without valid high availability", false))
			continue
		}

		// Skip any AffinityRequired pods
		if st.Spec.Template.Labels["AffinityRequired"] == "true" {
			continue
		}

		if st.Spec.Template.Spec.Affinity == nil ||
			st.Spec.Template.Spec.Affinity.PodAntiAffinity == nil {
			tnf.ClaimFilePrintf("StatefulSet found without valid high availability: %s", st.ToString())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewStatefulSetReportObject(st.Namespace, st.Name, "StatefulSet found without valid high availability", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewStatefulSetReportObject(st.Namespace, st.Name, "StatefulSet has valid high availability", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

// testPodsRecreation tests that pods belonging to deployments and statefulsets are re-created and ready in case a node is lost
func testPodsRecreation(check *checksdb.Check, env *provider.TestEnvironment) { //nolint:funlen,gocyclo
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	needsPostMortemInfo := true
	defer func() {
		if needsPostMortemInfo {
			tnf.ClaimFilePrintf(postmortem.Log())
		}
		// Since we are possible exiting early, we need to make sure we set the result at the end of the function.
		check.SetResult(compliantObjects, nonCompliantObjects)
	}()
	tnf.Logf(logrus.InfoLevel, "Testing node draining effect of deployment")
	tnf.Logf(logrus.InfoLevel, "Testing initial state for deployments")
	defer env.SetNeedsRefresh()

	// Before draining any node, wait until all podsets are ready. The timeout depends on the number of podsets to check.
	// timeout = k-mins + (1min * (num-deployments + num-statefulsets))
	allPodsetsReadyTimeout := timeoutPodSetReady + time.Minute*time.Duration(len(env.Deployments)+len(env.StatefulSets))
	claimsLog, notReadyDeployments, notReadyStatefulSets := podsets.WaitForAllPodSetsReady(env, allPodsetsReadyTimeout)
	if len(notReadyDeployments) > 0 || len(notReadyStatefulSets) > 0 {
		tnf.ClaimFilePrintf("%s", claimsLog.GetLogLines())
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
			logrus.Errorf("%s has been found with node selector(s): %v", put.String(), put.Spec.NodeSelector)
		}
	}
	if len(podsWithNodeAssignment) > 0 {
		logrus.Errorf("Pod(s) have been found to contain a node assignment and cannot perform the pod-recreation test: %v", podsWithNodeAssignment)
		for _, pod := range podsWithNodeAssignment {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(pod.Namespace, pod.Name, "Pod has node assignment.", false))
		}

		return
	}

	for nodeName := range podsets.GetAllNodesForAllPodSets(env.Pods) {
		defer podrecreation.CordonCleanup(nodeName) //nolint:gocritic // The defer in loop is intentional, calling the cleanup function once per node
		err := podrecreation.CordonHelper(nodeName, podrecreation.Cordon)
		if err != nil {
			logrus.Errorf("error cordoning the node: %s", nodeName)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(nodeName, "Node cordoning failed", false))
			return
		}
		tnf.Logf(logrus.InfoLevel, fmt.Sprintf("Draining and Cordoning node %s: ", nodeName))
		logrus.Debugf("node: %s cordoned", nodeName)
		count, err := podrecreation.CountPodsWithDelete(env.Pods, nodeName, podrecreation.NoDelete)
		if err != nil {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(nodeName, "Getting pods list to drain failed", false))
			return
		}
		nodeTimeout := timeoutPodSetReady + timeoutPodRecreationPerPod*time.Duration(count)
		logrus.Debugf("draining node: %s with timeout: %s", nodeName, nodeTimeout)
		_, err = podrecreation.CountPodsWithDelete(env.Pods, nodeName, podrecreation.DeleteForeground)
		if err != nil {
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewNodeReportObject(nodeName, "Draining node failed", false))
			return
		}

		claimsLog, notReadyDeployments, notReadyStatefulSets := podsets.WaitForAllPodSetsReady(env, nodeTimeout)
		if len(notReadyDeployments) > 0 || len(notReadyStatefulSets) > 0 {
			tnf.ClaimFilePrintf("%s", claimsLog.GetLogLines())
			for _, dep := range notReadyDeployments {
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewDeploymentReportObject(dep.Namespace, dep.Name, "Deployment not ready after draining node "+nodeName, false))
			}
			for _, sts := range notReadyStatefulSets {
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewStatefulSetReportObject(sts.Namespace, sts.Name, "Statefulset not ready after draining node "+nodeName, false))
			}
			return
		}

		err = podrecreation.CordonHelper(nodeName, podrecreation.Uncordon)
		if err != nil {
			logrus.Fatalf("error uncordoning the node: %s", nodeName)
		}
	}

	// If everything went well for all nodes, the nonCompliantObjects should be empty. We need to
	// manually add all the deps/sts into the compliant object lists so the check is marked as skipped.
	// ToDo: Improve this.
	if len(nonCompliantObjects) == 0 {
		for _, dep := range env.Deployments {
			compliantObjects = append(compliantObjects, testhelper.NewDeploymentReportObject(dep.Namespace, dep.Name, "Deployment's pods successfully re-schedulled after node draining.", true))
		}

		for _, sts := range env.StatefulSets {
			compliantObjects = append(compliantObjects, testhelper.NewStatefulSetReportObject(sts.Namespace, sts.Name, "Statefulset's pods successfully re-schedulled after node draining.", true))
		}
	}

	needsPostMortemInfo = false
}

func testPodPersistentVolumeReclaimPolicy(check *checksdb.Check, env *provider.TestEnvironment) {
	tnf.Logf(logrus.InfoLevel, "Testing PersistentVolumes for reclaim policy to be set to delete")
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	// Look through all of the pods, matching their persistent volumes to the list of overall cluster PVs and checking their reclaim status.
	for _, put := range env.Pods {
		compliantPod := true
		// Loop through all of the volumes attached to the pod.
		for pvIndex := range put.Spec.Volumes {
			// Skip any volumes that do not have a PVC.  No need to test them.
			if put.Spec.Volumes[pvIndex].PersistentVolumeClaim == nil {
				continue
			}

			// If the Pod Volume is not tied back to a PVC and corresponding PV that has a reclaim policy of DELETE.
			if !volumes.IsPodVolumeReclaimPolicyDelete(&put.Spec.Volumes[pvIndex], env.PersistentVolumes, env.PersistentVolumeClaims) {
				tnf.ClaimFilePrintf("%s contains volume: %s has been found without a reclaim policy of DELETE.", put.String(), &put.Spec.Volumes[pvIndex].Name)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod contains volume without a reclaim policy of DELETE", false).
					AddField(testhelper.PersistentVolumeName, put.Spec.Volumes[pvIndex].Name).
					AddField(testhelper.PersistentVolumeClaimName, put.Spec.Volumes[pvIndex].PersistentVolumeClaim.ClaimName))
				compliantPod = false
				break
			}
		}

		if compliantPod {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod complies with volume reclaim policy rules", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testCPUIsolation(check *checksdb.Check, env *provider.TestEnvironment) {
	tnf.Logf(logrus.InfoLevel, "Testing pods for CPU isolation requirements")

	// Individual requirements we are looking for:
	//  - CPU Requests and Limits must be in the form of whole units
	// - Resource Requests and Limits must be provided and identical

	// Additional checks if the above pass
	// - 'runtimeClassName' must be specified
	// - Annotations must be provided disabling CPU and IRQ load-balancing.

	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, put := range env.GetGuaranteedPodsWithExclusiveCPUs() {
		if !put.IsCPUIsolationCompliant() {
			tnf.ClaimFilePrintf("%s is not CPU isolated", put.String())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is not CPU isolated", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is CPU isolated", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testAffinityRequiredPods(check *checksdb.Check, env *provider.TestEnvironment) {
	tnf.Logf(logrus.InfoLevel, "Testing affinity required pods for ")

	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, put := range env.GetAffinityRequiredPods() {
		// Check if the pod is Affinity compliant.
		result, err := put.IsAffinityCompliant()
		if !result {
			tnf.ClaimFilePrintf(err.Error())
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is not Affinity compliant", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod is Affinity compliant", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testPodTolerationBypass(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, put := range env.Pods {
		podIsCompliant := true
		for _, t := range put.Spec.Tolerations {
			// Check if the tolerations fall outside the 'default' and are modified versions
			// Take also into account the qosClass applied to the pod
			if tolerations.IsTolerationModified(t, put.Status.QOSClass) {
				tnf.ClaimFilePrintf("%s has been found with non-default toleration %s/%s which is not allowed.", put.String(), t.Key, t.Effect)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has non-default toleration", false).
					AddField(testhelper.TolerationKey, t.Key).
					AddField(testhelper.TolerationEffect, string(t.Effect)))
				podIsCompliant = false
			}
		}

		if podIsCompliant {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has default toleration", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

//nolint:funlen
func testStorageProvisioner(check *checksdb.Check, env *provider.TestEnvironment) {
	const localStorageProvisioner = "kubernetes.io/no-provisioner"
	const lvmProvisioner = "topolvm.io"
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	var StorageClasses = env.StorageClassList
	var Pvc = env.PersistentVolumeClaims
	snoSingleLocalStorageProvisionner := ""
	for _, put := range env.Pods {
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
							tnf.ClaimFilePrintf("%s pvc_name: %s, storageclass_name: %s, provisioner_name: %s", put.String(), put.Spec.Volumes[pvIndex].PersistentVolumeClaim.ClaimName,
								StorageClasses[j].Name, StorageClasses[j].Provisioner)

							if env.IsSNO() {
								// For SNO, only one local storage provisionner is allowed. The first local storage provisioner for this pod is assumed to be the only local storage provisioner allowed in the cluster.
								if snoSingleLocalStorageProvisionner == "" &&
									(StorageClasses[j].Provisioner == localStorageProvisioner ||
										StorageClasses[j].Provisioner == lvmProvisioner) {
									snoSingleLocalStorageProvisionner = StorageClasses[j].Provisioner
								}
								if StorageClasses[j].Provisioner == snoSingleLocalStorageProvisionner {
									compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Local storage (no provisioner or lvms) is recommended for SNO clusters.", false).
										AddField(testhelper.StorageClassName, StorageClasses[j].Name).
										AddField(testhelper.StorageClassProvisioner, StorageClasses[j].Provisioner).
										AddField(testhelper.PersistentVolumeClaimName, put.Spec.Volumes[pvIndex].PersistentVolumeClaim.ClaimName))
									continue
								}
								if StorageClasses[j].Provisioner == localStorageProvisioner || StorageClasses[j].Provisioner == lvmProvisioner {
									nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name,
										"A single type of local storage cluster is recommended for single node clusters. Use lvms or kubernetes noprovisioner, but not both.", false).
										AddField(testhelper.StorageClassName, StorageClasses[j].Name).
										AddField(testhelper.StorageClassProvisioner, StorageClasses[j].Provisioner).
										AddField(testhelper.PersistentVolumeClaimName, put.Spec.Volumes[pvIndex].PersistentVolumeClaim.ClaimName))
									continue
								}
								nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Non local storage not recommended in single node clusters.", false).
									AddField(testhelper.StorageClassName, StorageClasses[j].Name).
									AddField(testhelper.StorageClassProvisioner, StorageClasses[j].Provisioner).
									AddField(testhelper.PersistentVolumeClaimName, put.Spec.Volumes[pvIndex].PersistentVolumeClaim.ClaimName))
							} else {
								if StorageClasses[j].Provisioner == localStorageProvisioner || StorageClasses[j].Provisioner == lvmProvisioner {
									nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Local storage provisioner (no provisioner or lvms) not recommended in multinode clusters.", false).
										AddField(testhelper.StorageClassName, StorageClasses[j].Name).
										AddField(testhelper.StorageClassProvisioner, StorageClasses[j].Provisioner).
										AddField(testhelper.PersistentVolumeClaimName, put.Spec.Volumes[pvIndex].PersistentVolumeClaim.ClaimName))
									continue
								}
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
				compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod not configured to use local storage.", true))
			}
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}
