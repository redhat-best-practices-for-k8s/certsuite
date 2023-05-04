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

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/ownerreference"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/podrecreation"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/podsets"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/scaling"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/tolerations"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/volumes"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
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
)

// All actual test code belongs below here.  Utilities belong above.
var _ = ginkgo.Describe(common.LifecycleTestKey, func() {
	logrus.Debugf("Entering %s suite", common.LifecycleTestKey)
	env := provider.GetTestEnvironment()
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})
	ginkgo.ReportAfterEach(results.RecordResult)

	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestShutdownIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAll(ginkgo.Skip, env.Containers)
		testContainersPreStop(&env)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestCrdScalingIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.ScaleCrUndetTest)
		// Note: We skip this test because 'testHighAvailability' in the lifecycle suite is already
		// testing the replicas and antiaffinity rules that should already be in place for crd.
		testScaleCrd(&env, timeout)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestStartupIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAll(ginkgo.Skip, env.Containers)
		testContainersPostStart(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestImagePullPolicyIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAll(ginkgo.Skip, env.Containers)
		testContainersImagePolicy(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestReadinessProbeIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAll(ginkgo.Skip, env.Containers)
		testContainersReadinessProbe(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestLivenessProbeIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAll(ginkgo.Skip, env.Containers)
		testContainersLivenessProbe(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestStartupProbeIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAll(ginkgo.Skip, env.Containers)
		testContainersStartupProbe(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodDeploymentBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAll(ginkgo.Skip, env.Pods)
		testPodsOwnerReference(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodHighAvailabilityBestPractices)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if env.GetWorkerCount() < minWorkerNodesForLifecycle {
			ginkgo.Skip("Skipping pod high availability test because invalid number of available workers.")
		}
		testhelper.SkipIfEmptyAll(ginkgo.Skip, env.Deployments, env.StatefulSets)
		testHighAvailability(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodNodeSelectorAndAffinityBestPractices)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if env.GetWorkerCount() < minWorkerNodesForLifecycle {
			ginkgo.Skip("Skipping pod scheduling test because invalid number of available workers.")
		}
		testPods := env.GetPodsWithoutAffinityRequiredLabel()
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testPods)
		testPodNodeSelectorAndAffinityBestPractices(testPods)
	})

	if env.IsIntrusive() {
		testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodRecreationIdentifier)
		ginkgo.It(testID, ginkgo.Label(tags...), func() {
			testhelper.SkipIfEmptyAll(ginkgo.Skip, env.Deployments, env.StatefulSets)
			if env.GetWorkerCount() < minWorkerNodesForLifecycle {
				ginkgo.Skip("Skipping pod recreation scaling test because invalid number of available workers.")
			}
			// Testing pod re-creation for deployments
			testPodsRecreation(&env)
		})

		testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestDeploymentScalingIdentifier)
		ginkgo.It(testID, ginkgo.Label(tags...), func() {
			testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Deployments)
			if env.GetWorkerCount() < minWorkerNodesForLifecycle {
				// Note: We skip this test because 'testHighAvailability' in the lifecycle suite is already
				// testing the replicas and antiaffinity rules that should already be in place for deployments.
				ginkgo.Skip("Skipping deployment scaling test because invalid number of available workers.")
			}
			testDeploymentScaling(&env, timeout)
		})
		testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestStateFulSetScalingIdentifier)
		ginkgo.It(testID, ginkgo.Label(tags...), func() {
			testhelper.SkipIfEmptyAny(ginkgo.Skip, env.StatefulSets)
			if env.GetWorkerCount() < minWorkerNodesForLifecycle {
				// Note: We skip this test because 'testHighAvailability' in the lifecycle suite is already
				// testing the replicas and antiaffinity rules that should already be in place for statefulset.
				ginkgo.Skip("Skipping statefulset scaling test because invalid number of available workers.")
			}
			testStatefulSetScaling(&env, timeout)
		})
	}

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPersistentVolumeReclaimPolicyIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods, env.PersistentVolumes)
		testPodPersistentVolumeReclaimPolicy(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestCPUIsolationIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.GetGuaranteedPodsWithExclusiveCPUs())
		testCPUIsolation(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestAffinityRequiredPods)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testAffinityRequiredPods(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodTolerationBypassIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		testPodTolerationBypass(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestStorageRequiredPods)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testStorageRequiredPods(&env)
	})
})

func testContainersPreStop(env *provider.TestEnvironment) {
	badContainers := []string{}
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " pre stop lifecycle ")

		if cut.Lifecycle == nil || (cut.Lifecycle != nil && cut.Lifecycle.PreStop == nil) {
			badContainers = append(badContainers, cut.Name)
			tnf.ClaimFilePrintf("%s does not have preStop defined", cut)
		}
	}
	testhelper.AddTestResultLog("Non-compliant", badContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testContainersPostStart(env *provider.TestEnvironment) {
	badContainers := []string{}
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " post start lifecycle ")

		if cut.Lifecycle == nil || (cut.Lifecycle != nil && cut.Lifecycle.PostStart == nil) {
			badContainers = append(badContainers, cut.Name)
			tnf.ClaimFilePrintf("%s does not have postStart defined", cut)
		}
	}
	testhelper.AddTestResultLog("Non-compliant", badContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testContainersImagePolicy(env *provider.TestEnvironment) {
	testhelper.SkipIfEmptyAll(ginkgo.Skip, env.Containers)
	badContainers := []string{}
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " pull policy, should be ", corev1.PullIfNotPresent)
		if cut.ImagePullPolicy != corev1.PullIfNotPresent {
			badContainers = append(badContainers, "{"+cut.String()+": is using"+string(cut.ImagePullPolicy)+"}")
			logrus.Errorln("container ", cut.Name, " is using ", cut.ImagePullPolicy, " as image pull policy")
			tnf.ClaimFilePrintf("%s is using %s as ImagePullPolicy", cut.String(), cut.ImagePullPolicy)
		}
	}
	testhelper.AddTestResultLog("Non-compliant", badContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testContainersReadinessProbe(env *provider.TestEnvironment) {
	badContainers := []string{}
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " readiness probe ")
		if cut.ReadinessProbe == nil {
			badContainers = append(badContainers, cut.String())
			logrus.Errorln("container ", cut.Name, " does not have ReadinessProbe defined")
			tnf.ClaimFilePrintf("%s does not have ReadinessProbe defined", cut.String())
		}
	}
	testhelper.AddTestResultLog("Non-compliant", badContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testContainersLivenessProbe(env *provider.TestEnvironment) {
	badContainers := []string{}
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " liveness probe ")
		if cut.LivenessProbe == nil {
			badContainers = append(badContainers, cut.String())
			logrus.Errorln("container ", cut.Name, " does not have LivenessProbe defined")
			tnf.ClaimFilePrintf("%s does not have LivenessProbe defined", cut.String())
		}
	}
	testhelper.AddTestResultLog("Non-compliant", badContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testContainersStartupProbe(env *provider.TestEnvironment) {
	badContainers := []string{}
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " startup probe ")
		if cut.StartupProbe == nil {
			badContainers = append(badContainers, cut.String())
			logrus.Errorln("container ", cut.Name, " does not have startupProbe defined")
			tnf.ClaimFilePrintf("%s does not have StartupProbe defined", cut.String())
		}
	}
	testhelper.AddTestResultLog("Non-compliant", badContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testPodsOwnerReference(env *provider.TestEnvironment) {
	ginkgo.By("Testing owners of CNF pod, should be replicas Set")
	badPods := []string{}
	for _, put := range env.Pods {
		logrus.Debugln("check pod ", put.Namespace, " ", put.Name, " owner reference")
		o := ownerreference.NewOwnerReference(put.Pod)
		o.RunTest()
		if o.GetResults() != testhelper.SUCCESS {
			badPods = append(badPods, put.String())
			tnf.ClaimFilePrintf("%s found with non-compliant owner reference", put.String())
		}
	}
	testhelper.AddTestResultLog("Non-compliant", badPods, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testPodNodeSelectorAndAffinityBestPractices(testPods []*provider.Pod) {
	var badPods []*corev1.Pod
	for _, put := range testPods {
		if put.HasNodeSelector() {
			tnf.ClaimFilePrintf("ERROR: %s has a node selector. Node selector: %v", put, &put.Spec.NodeSelector)
			badPods = append(badPods, put.Pod)
		}
		if put.Spec.Affinity != nil && put.Spec.Affinity.NodeAffinity != nil {
			tnf.ClaimFilePrintf("ERROR: %s has a node affinity clause. Node affinity: %v", put, put.Spec.Affinity.NodeAffinity)
			badPods = append(badPods, put.Pod)
		}
	}
	testhelper.AddTestResultLog("Non-compliant", badPods, tnf.ClaimFilePrintf, ginkgo.Fail)
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
func testDeploymentScaling(env *provider.TestEnvironment, timeout time.Duration) {
	ginkgo.By("Testing deployment scaling")
	defer env.SetNeedsRefresh()
	failedDeployments := []string{}
	for i := range env.Deployments {
		if scaling.IsManaged(env.Deployments[i].Name, env.Config.ManagedDeployments) {
			if !scaling.CheckOwnerReference(env.Deployments[i].GetOwnerReferences(), env.Config.CrdFilters, env.Crds) {
				failedDeployments = append(failedDeployments, env.Deployments[i].ToString())
				tnf.ClaimFilePrintf("%s is scaling failed due to OwnerReferences that are not scalable", env.Deployments[i].ToString())
			} else {
				logrus.Infof("%s is scaling skipped due to scalable OwnerReferences, test will run on te cr scaling", env.Deployments[i].ToString())
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
				failedDeployments = append(failedDeployments, env.Deployments[i].ToString())
				tnf.ClaimFilePrintf("Deployment has failed the HPA scale test: %s", env.Deployments[i].ToString())
			}
			continue
		}
		// if the deployment is not controller by HPA
		// scale it directly
		if !scaling.TestScaleDeployment(env.Deployments[i].Deployment, timeout) {
			failedDeployments = append(failedDeployments, env.Deployments[i].ToString())
			tnf.ClaimFilePrintf("Deployment has failed the non-HPA scale test: %s", env.Deployments[i].ToString())
		}
	}

	testhelper.AddTestResultLog("Non-compliant", failedDeployments, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testScaleCrd(env *provider.TestEnvironment, timeout time.Duration) {
	ginkgo.By("Testing deployment scaling")
	defer env.SetNeedsRefresh()
	failedcrd := []string{}
	for i := range env.ScaleCrUndetTest {
		groupResourceSchema := env.ScaleCrUndetTest[i].GroupResourceSchema
		scaleCr := env.ScaleCrUndetTest[i].Scale
		if hpa := scaling.GetResourceHPA(env.HorizontalScaler, scaleCr.Name, scaleCr.Namespace, scaleCr.Kind); hpa != nil {
			if !scaling.TestScaleHPACrd(&scaleCr, hpa, groupResourceSchema, timeout) {
				tnf.ClaimFilePrintf("cr found to have failed the scaling test: %s", scaleCr.GetName())
				failedcrd = append(failedcrd, scaleCr.ToString())
			}
			continue
		}
		if !scaling.TestScaleCrd(&scaleCr, groupResourceSchema, timeout) {
			failedcrd = append(failedcrd, scaleCr.ToString())
			tnf.ClaimFilePrintf("CR has failed the non-HPA scale test: %s", scaleCr.GetName())
		}
	}
	testhelper.AddTestResultLog("Non-compliant", failedcrd, tnf.ClaimFilePrintf, ginkgo.Fail)
}

//nolint:dupl
func testStatefulSetScaling(env *provider.TestEnvironment, timeout time.Duration) {
	ginkgo.By("Testing statefulset scaling")
	defer env.SetNeedsRefresh()
	failedStatefulSets := []string{}
	for i := range env.StatefulSets {
		if scaling.IsManaged(env.StatefulSets[i].Name, env.Config.ManagedStatefulsets) {
			if !scaling.CheckOwnerReference(env.StatefulSets[i].GetOwnerReferences(), env.Config.CrdFilters, env.Crds) {
				failedStatefulSets = append(failedStatefulSets, env.StatefulSets[i].ToString())
				tnf.ClaimFilePrintf("%s is scaling failed due to OwnerReferences that are not scalable", env.Deployments[i].ToString())
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
				tnf.ClaimFilePrintf("StatefulSet found to have failed the scaling test: %s", env.StatefulSets[i].ToString())
				failedStatefulSets = append(failedStatefulSets, env.StatefulSets[i].ToString())
			}
			continue
		}
		// if the statefulset is not controller by HPA
		// scale it directly
		if !scaling.TestScaleStatefulSet(env.StatefulSets[i].StatefulSet, timeout) {
			tnf.ClaimFilePrintf("StatefulSet found to have failed the scaling test: %s", env.StatefulSets[i].ToString())
			failedStatefulSets = append(failedStatefulSets, env.StatefulSets[i].ToString())
		}
	}

	testhelper.AddTestResultLog("Non-compliant", failedStatefulSets, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testHighAvailability
func testHighAvailability(env *provider.TestEnvironment) {
	ginkgo.By("Should set pod replica number greater than 1")

	badDeployments := []string{}
	badStatefulSet := []string{}
	for _, dp := range env.Deployments {
		if dp.Spec.Replicas == nil || *(dp.Spec.Replicas) <= 1 {
			badDeployments = append(badDeployments, dp.ToString())
			tnf.ClaimFilePrintf("Deployment found without valid high availability: %s", dp.ToString())
			continue
		}

		// Skip any AffinityRequired pods
		//nolint:goconst
		if dp.Spec.Template.Labels["AffinityRequired"] == "true" {
			continue
		}

		if dp.Spec.Template.Spec.Affinity == nil ||
			dp.Spec.Template.Spec.Affinity.PodAntiAffinity == nil {
			badDeployments = append(badDeployments, dp.ToString())
			tnf.ClaimFilePrintf("Deployment found without valid high availability: %s", dp.ToString())
		}
	}
	for _, st := range env.StatefulSets {
		if st.Spec.Replicas == nil || *(st.Spec.Replicas) <= 1 {
			badStatefulSet = append(badStatefulSet, st.ToString())
			tnf.ClaimFilePrintf("StatefulSet found without valid high availability: %s", st.ToString())
			continue
		}

		// Skip any AffinityRequired pods
		if st.Spec.Template.Labels["AffinityRequired"] == "true" {
			continue
		}

		if st.Spec.Template.Spec.Affinity == nil ||
			st.Spec.Template.Spec.Affinity.PodAntiAffinity == nil {
			badStatefulSet = append(badStatefulSet, st.ToString())
			tnf.ClaimFilePrintf("StatefulSet found without valid high availability: %s", st.ToString())
		}
	}

	testhelper.AddTestResultLog("Non-compliant", badDeployments, tnf.ClaimFilePrintf, ginkgo.Fail)
	testhelper.AddTestResultLog("Non-compliant", badStatefulSet, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testPodsRecreation tests that pods belonging to deployments and statefulsets are re-created and ready in case a node is lost
func testPodsRecreation(env *provider.TestEnvironment) { //nolint:funlen
	needsPostMortemInfo := true
	defer func() {
		if needsPostMortemInfo {
			tnf.ClaimFilePrintf(postmortem.Log())
		}
	}()
	ginkgo.By("Testing node draining effect of deployment")
	ginkgo.By("Testing initial state for deployments")
	defer env.SetNeedsRefresh()
	claimsLog, atLeastOnePodsetNotReady := podsets.WaitForAllPodSetReady(env, timeoutPodSetReady)
	if atLeastOnePodsetNotReady {
		tnf.ClaimFilePrintf("%s", claimsLog.GetLogLines())
		ginkgo.Fail("Some deployments or stateful sets are not in a good initial state. Cannot perform test.")
	}

	// Filter out pods with Node Assignments present and FAIL them.
	// We run into problems with this test when there are nodeSelectors assigned affecting where
	// pods are scheduled.  Also, they are not allowed in general, see the node-selector test case.
	// Skip the safeguard for any pods that are using a runtimeClassName.  This is potentially
	// because pods that are derived from a performance profile might have a built-in nodeSelector.
	var podsWithNodeAssignement []*provider.Pod
	for _, put := range env.Pods {
		if !put.IsRuntimeClassNameSpecified() && put.HasNodeSelector() {
			podsWithNodeAssignement = append(podsWithNodeAssignement, put)
			logrus.Errorf("%s has been found with node selector(s): %v", put.String(), put.Spec.NodeSelector)
		}
	}
	if len(podsWithNodeAssignement) > 0 {
		logrus.Errorf("Pod(s) have been found to contain a node assignment and cannot perform the pod-recreation test: %v", podsWithNodeAssignement)
		testhelper.AddTestResultLog("Non-compliant", podsWithNodeAssignement, tnf.ClaimFilePrintf, ginkgo.Fail)
	}

	for n := range podsets.GetAllNodesForAllPodSets(env.Pods) {
		defer podrecreation.CordonCleanup(n) //nolint:gocritic // The defer in loop is intentional, calling the cleanup function once per node
		err := podrecreation.CordonHelper(n, podrecreation.Cordon)
		if err != nil {
			logrus.Errorf("error cordoning the node: %s", n)
			ginkgo.Fail(fmt.Sprintf("Cordoning node %s failed with err: %v. Test inconclusive, skipping", n, err))
		}
		ginkgo.By(fmt.Sprintf("Draining and Cordoning node %s: ", n))
		logrus.Debugf("node: %s cordoned", n)
		count, err := podrecreation.CountPodsWithDelete(env.Pods, n, podrecreation.NoDelete)
		if err != nil {
			ginkgo.Fail(fmt.Sprintf("Getting pods list to drain in node %s failed with err: %v. Test inconclusive.", n, err))
		}
		nodeTimeout := timeoutPodSetReady + timeoutPodRecreationPerPod*time.Duration(count)
		logrus.Debugf("draining node: %s with timeout: %s", n, nodeTimeout.String())
		_, err = podrecreation.CountPodsWithDelete(env.Pods, n, podrecreation.DeleteForeground)
		if err != nil {
			ginkgo.Fail(fmt.Sprintf("Draining node %s failed with err: %v. Test inconclusive", n, err))
		}

		claimsLog, podsNotReady := podsets.WaitForAllPodSetReady(env, nodeTimeout)
		if podsNotReady {
			tnf.ClaimFilePrintf("%s", claimsLog.GetLogLines())
			ginkgo.Fail(fmt.Sprintf("Some pods are not ready after draining the node %s", n))
		}

		err = podrecreation.CordonHelper(n, podrecreation.Uncordon)
		if err != nil {
			logrus.Fatalf("error uncordoning the node: %s", n)
		}

		// Reached end of TC, which means no ginkgo.Fail() was called.
		needsPostMortemInfo = false
	}
}

func testPodPersistentVolumeReclaimPolicy(env *provider.TestEnvironment) {
	ginkgo.By("Testing PersistentVolumes for reclaim policy to be set to delete")
	var persistentVolumesBadReclaim []string

	// Look through all of the pods, matching their persistent volumes to the list of overall cluster PVs and checking their reclaim status.
	for _, put := range env.Pods {
		// Loop through all of the volumes attached to the pod.
		for pvIndex := range put.Spec.Volumes {
			// Skip any volumes that do not have a PVC.  No need to test them.
			if put.Spec.Volumes[pvIndex].PersistentVolumeClaim == nil {
				continue
			}

			// If the Pod Volume is not tied back to a PVC and corresponding PV that has a reclaim policy of DELETE.
			if !volumes.IsPodVolumeReclaimPolicyDelete(&put.Spec.Volumes[pvIndex], env.PersistentVolumes, env.PersistentVolumeClaims) {
				persistentVolumesBadReclaim = append(persistentVolumesBadReclaim, put.String())
				tnf.ClaimFilePrintf("%s contains volume: %s has been found without a reclaim policy of DELETE.", put.String(), &put.Spec.Volumes[pvIndex].Name)
				break
			}
		}
	}

	if n := len(persistentVolumesBadReclaim); n > 0 {
		testhelper.AddTestResultLog("Non-compliant", persistentVolumesBadReclaim, tnf.ClaimFilePrintf, ginkgo.Fail)
	}
}

func testCPUIsolation(env *provider.TestEnvironment) {
	ginkgo.By("Testing pods for CPU isolation requirements")

	// Individual requirements we are looking for:
	//  - CPU Requests and Limits must be in the form of whole units
	// - Resource Requests and Limits must be provided and identical

	// Additional checks if the above pass
	// - 'runtimeClassName' must be specified
	// - Annotations must be provided disabling CPU and IRQ load-balancing.

	podsMissingIsolationRequirements := make(map[string]bool)

	for _, put := range env.GetGuaranteedPodsWithExclusiveCPUs() {
		if !put.IsCPUIsolationCompliant() {
			podsMissingIsolationRequirements[put.Name] = true
		}
	}

	testhelper.AddTestResultLog("Non-compliant", podsMissingIsolationRequirements, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testAffinityRequiredPods(env *provider.TestEnvironment) {
	ginkgo.By("Testing affinity required pods for ")
	testhelper.SkipIfEmptyAny(ginkgo.Skip, env.GetAffinityRequiredPods())

	var podsDesiringAffinityRequiredMissingLabel []*provider.Pod
	for _, put := range env.GetAffinityRequiredPods() {
		// Check if the pod is Affinity compliant.
		result, err := put.IsAffinityCompliant()
		if !result {
			tnf.ClaimFilePrintf(err.Error())
			podsDesiringAffinityRequiredMissingLabel = append(podsDesiringAffinityRequiredMissingLabel, put)
		}
	}
	testhelper.AddTestResultLog("Non-compliant", podsDesiringAffinityRequiredMissingLabel, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testPodTolerationBypass(env *provider.TestEnvironment) {
	var podsWithRestrictedTolerationsNotDefault []string

	for _, put := range env.Pods {
		for _, t := range put.Spec.Tolerations {
			// Check if the tolerations fall outside the 'default' and are modified versions
			// Take also into account the qosClass applied to the pod
			if tolerations.IsTolerationModified(t, put.Status.QOSClass) {
				podsWithRestrictedTolerationsNotDefault = append(podsWithRestrictedTolerationsNotDefault, put.String())
				tnf.ClaimFilePrintf("%s has been found with non-default toleration %s/%s which is not allowed.", put.String(), t.Key, t.Effect)
			}
		}
	}

	testhelper.AddTestResultLog("Non-compliant", podsWithRestrictedTolerationsNotDefault, tnf.ClaimFilePrintf, ginkgo.Fail)
}
func testStorageRequiredPods(env *provider.TestEnvironment) {
	const localStorageProvisioner = "kubernetes.io/no-provisioner"
	var podsWithLocalStorage []string
	var StorageClasses = env.StorageClassList
	var Pvc = env.PersistentVolumeClaims
	for _, put := range env.Pods {
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
						if StorageClasses[j].Provisioner != localStorageProvisioner {
							continue
						}

						if Pvc[i].Spec.StorageClassName != nil && StorageClasses[j].Name == *Pvc[i].Spec.StorageClassName {
							tnf.ClaimFilePrintf("%s has been found to use a local storage enabled storageClass.\n Pvc_name: %s, Storageclass_name : %s, Provisioner_name: %s", put.String(), put.Spec.Volumes[pvIndex].PersistentVolumeClaim.ClaimName,
								StorageClasses[j].Name, StorageClasses[j].Provisioner)
							podsWithLocalStorage = append(podsWithLocalStorage, put.String())
							break
						}
						tnf.ClaimFilePrintf("%s has been not found to use a local storage enabled storageClass.\n Pvc_name: %s, Storageclass_name : %s, Provisioner_name: %s", put.String(), put.Spec.Volumes[pvIndex].PersistentVolumeClaim.ClaimName,
							StorageClasses[j].Name, StorageClasses[j].Provisioner)
					}
				}
			}
		}
	}

	if n := len(podsWithLocalStorage); n > 0 {
		testhelper.AddTestResultLog("Non-compliant", podsWithLocalStorage, tnf.ClaimFilePrintf, ginkgo.Fail)
	}
}
