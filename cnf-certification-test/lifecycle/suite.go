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
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/volumes"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/pkg/postmortem"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
	corev1 "k8s.io/api/core/v1"
)

const (
	timeout                    = 60 * time.Second
	timeoutPodRecreationPerPod = time.Minute
	timeoutPodSetReady         = 7 * time.Minute
	minWorkerNodesForLifecycle = 2
)

// All actual test code belongs below here.  Utilities belong above.
var _ = ginkgo.Describe(common.LifecycleTestKey, func() {
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})
	ginkgo.ReportAfterEach(results.RecordResult)

	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestShutdownIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAll(ginkgo.Skip, env.Containers)
		testContainersPreStop(&env)
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
		testhelper.SkipIfEmptyAll(ginkgo.Skip, env.AntiAffinityRequiredDeployments, env.AntiAffinityRequiredStatefulSets)
		testHighAvailability(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodNodeSelectorAndAffinityBestPractices)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if env.GetWorkerCount() < minWorkerNodesForLifecycle {
			ginkgo.Skip("Skipping pod scheduling test because invalid number of available workers.")
		}
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.NonGuaranteedPods)
		testPodNodeSelectorAndAffinityBestPractices(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodRecreationIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAll(ginkgo.Skip, env.Deployments, env.StatetfulSets)
		if env.GetWorkerCount() < minWorkerNodesForLifecycle {
			ginkgo.Skip("Skipping pod recreation scaling test because invalid number of available workers.")
		}
		// Testing pod re-creation for deployments
		testPodsRecreation(&env)
	})

	if env.IsIntrusive() {
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
			testhelper.SkipIfEmptyAny(ginkgo.Skip, env.StatetfulSets)
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
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.GuaranteedPods)
		testCPUIsolation(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestAffinityRequiredPods)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testAffinityRequiredPods(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestAffinityRequiredStatefulSets)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testAffinityRequiredStatefulSets(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestAffinityRequiredDeployments)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testAffinityRequiredDeployments(&env)
	})
})

func testContainersPreStop(env *provider.TestEnvironment) {
	badcontainers := []string{}
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " pre stop lifecycle ")

		if cut.Lifecycle == nil || (cut.Lifecycle != nil && cut.Lifecycle.PreStop == nil) {
			badcontainers = append(badcontainers, cut.Name)
			tnf.ClaimFilePrintf("%s does not have preStop defined", cut)
		}
	}
	if len(badcontainers) > 0 {
		tnf.ClaimFilePrintf("Containers have been found missing lifecycle preStop definitions: %v", badcontainers)
		ginkgo.Fail("Containers have been found missing lifecycle preStop definitions.")
	}
}

func testContainersImagePolicy(env *provider.TestEnvironment) {
	testhelper.SkipIfEmptyAll(ginkgo.Skip, env.Containers)
	badcontainers := []string{}
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " pull policy, should be ", corev1.PullIfNotPresent)
		if cut.ImagePullPolicy != corev1.PullIfNotPresent {
			badcontainers = append(badcontainers, "{"+cut.String()+": is using"+string(cut.ImagePullPolicy)+"}")
			logrus.Errorln("container ", cut.Name, " is using ", cut.ImagePullPolicy, " as image policy")
		}
	}
	if len(badcontainers) > 0 {
		tnf.ClaimFilePrintf("Containers have been found with IfNotPresent missing from image pull policy: %v", badcontainers)
		ginkgo.Fail("Containers have been found with IfNotPresent missing from image pull policy.")
	}
}

func testContainersReadinessProbe(env *provider.TestEnvironment) {
	badcontainers := []string{}
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " readiness probe ")
		if cut.ReadinessProbe == nil {
			badcontainers = append(badcontainers, cut.String())
			logrus.Errorln("container ", cut.Name, " does not have ReadinessProbe defined")
		}
	}
	if len(badcontainers) > 0 {
		tnf.ClaimFilePrintf("Containers have been found without readiness probes defined: %v", badcontainers)
		ginkgo.Fail("Containers have been found without readiness probes defined.")
	}
}

func testContainersLivenessProbe(env *provider.TestEnvironment) {
	badcontainers := []string{}
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " liveness probe ")
		if cut.LivenessProbe == nil {
			badcontainers = append(badcontainers, cut.String())
			logrus.Errorln("container ", cut.Name, " does not have livenessProbe defined")
		}
	}
	if len(badcontainers) > 0 {
		tnf.ClaimFilePrintf("Containers have been found without livenessProbe defined: %v", badcontainers)
		ginkgo.Fail("Containers have been found without livenessProbe defined.")
	}
}

func testContainersStartupProbe(env *provider.TestEnvironment) {
	badcontainers := []string{}
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " startup probe ")
		if cut.StartupProbe == nil {
			badcontainers = append(badcontainers, cut.String())
			logrus.Errorln("container ", cut.Name, " does not have startupProbe defined")
		}
	}
	if len(badcontainers) > 0 {
		tnf.ClaimFilePrintf("Containers have been found without startupProbe defined: %v", badcontainers)
		ginkgo.Fail("Containers have been found without startupProbe defined.")
	}
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
		}
	}
	if len(badPods) > 0 {
		tnf.ClaimFilePrintf("Containers were found with incorrect owner reference: %v", badPods)
		ginkgo.Fail("Containers were found with incorrect owner reference.")
	}
}

func testPodNodeSelectorAndAffinityBestPractices(env *provider.TestEnvironment) {
	var badPods []*corev1.Pod
	for _, put := range env.NonGuaranteedPods {
		if len(put.Spec.NodeSelector) != 0 {
			tnf.ClaimFilePrintf("ERROR: %s has a node selector clause. Node selector: %v", put, &put.Spec.NodeSelector)
			badPods = append(badPods, put.Pod)
		}
		// Skip any affinity required pods.  Not applicable for this test.
		if !put.AffinityRequired() && put.Spec.Affinity != nil && put.Spec.Affinity.NodeAffinity != nil {
			tnf.ClaimFilePrintf("ERROR: %s has a node affinity clause. Node affinity: %v", put, put.Spec.Affinity.NodeAffinity)
			badPods = append(badPods, put.Pod)
		}
	}
	if n := len(badPods); n > 0 {
		logrus.Debugf("Pods with nodeSelector/nodeAffinity: %+v", badPods)
		ginkgo.Fail(fmt.Sprintf("%d pods found with nodeSelector/nodeAffinity rules", n))
	}
}

//nolint:dupl
func testDeploymentScaling(env *provider.TestEnvironment, timeout time.Duration) {
	ginkgo.By("Testing deployment scaling")
	defer env.SetNeedsRefresh()
	failedDeployments := []string{}
	for i := range env.Deployments {
		// TestDeploymentScaling test scaling of deployment
		// This is the entry point for deployment scaling tests
		ns, name := env.Deployments[i].Namespace, env.Deployments[i].Name
		key := ns + name
		if hpa, ok := env.HorizontalScaler[key]; ok {
			// if the deployment is controller by
			// horizontal scaler, then test that scaler
			// can scale the deployment
			if !scaling.TestScaleHpaDeployment(env.Deployments[i], hpa, timeout) {
				failedDeployments = append(failedDeployments, env.Deployments[i].ToString())
			}
			continue
		}
		// if the deployment is not controller by HPA
		// scale it directly
		if !scaling.TestScaleDeployment(env.Deployments[i].Deployment, timeout) {
			failedDeployments = append(failedDeployments, env.Deployments[i].ToString())
		}
	}

	if len(failedDeployments) > 0 {
		tnf.ClaimFilePrintf("Deployments were found to have failed the scaling test: %v", failedDeployments)
		ginkgo.Fail("Deployments were found to have failed the scaling test.")
	}
}

//nolint:dupl
func testStatefulSetScaling(env *provider.TestEnvironment, timeout time.Duration) {
	ginkgo.By("Testing statefulset scaling")
	defer env.SetNeedsRefresh()
	failedStatetfulSets := []string{}
	for i := range env.StatetfulSets {
		// TeststatefulsetScaling test scaling of statefulset
		// This is the entry point for statefulset scaling tests
		ns, name := env.StatetfulSets[i].Namespace, env.StatetfulSets[i].Name
		key := ns + name
		if hpa, ok := env.HorizontalScaler[key]; ok {
			// if the statefulset is controller by
			// horizontal scaler, then test that scaler
			// can scale the statefulset
			if !scaling.TestScaleHpaStatefulSet(env.StatetfulSets[i].StatefulSet, hpa, timeout) {
				failedStatetfulSets = append(failedStatetfulSets, env.StatetfulSets[i].ToString())
			}
			continue
		}
		// if the statefulset is not controller by HPA
		// scale it directly
		if !scaling.TestScaleStatefulSet(env.StatetfulSets[i].StatefulSet, timeout) {
			failedStatetfulSets = append(failedStatetfulSets, env.StatetfulSets[i].ToString())
		}
	}

	if len(failedStatetfulSets) > 0 {
		tnf.ClaimFilePrintf("Statefulsets were found to have failed the scaling test: %v", failedStatetfulSets)
		ginkgo.Fail("Statefulsets were found to have failed the scaling test.")
	}
}

// testHighAvailability
func testHighAvailability(env *provider.TestEnvironment) {
	ginkgo.By("Should set pod replica number greater than 1")

	badDeployments := []string{}
	badStatefulSet := []string{}
	for _, dp := range env.AntiAffinityRequiredDeployments {
		if dp.Spec.Replicas == nil || *(dp.Spec.Replicas) <= 1 {
			badDeployments = append(badDeployments, dp.ToString())
			continue
		}
		if dp.Spec.Template.Spec.Affinity == nil ||
			dp.Spec.Template.Spec.Affinity.PodAntiAffinity == nil {
			badDeployments = append(badDeployments, dp.ToString())
		}
	}
	for _, st := range env.AntiAffinityRequiredStatefulSets {
		if st.Spec.Replicas == nil || *(st.Spec.Replicas) <= 1 {
			badStatefulSet = append(badStatefulSet, st.ToString())
			continue
		}
		if st.Spec.Template.Spec.Affinity == nil ||
			st.Spec.Template.Spec.Affinity.PodAntiAffinity == nil {
			badDeployments = append(badDeployments, st.ToString())
		}
	}

	if n := len(badDeployments); n > 0 {
		logrus.Errorf("Deployments without a valid high availability : %+v", badDeployments)
		tnf.ClaimFilePrintf("Deployments without a valid high availability : %+v", badDeployments)
		ginkgo.Fail("Deployments were found without a valid high availability.")
	}
	if n := len(badStatefulSet); n > 0 {
		logrus.Errorf("Statefulset without a valid podAntiAffinity rule: %+v", badStatefulSet)
		tnf.ClaimFilePrintf("Statefulset without a valid podAntiAffinity rule: %+v", badStatefulSet)
		ginkgo.Fail("Statefulsets were found without a valid high availability.")
	}
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
	for n := range podsets.GetAllNodesForAllPodSets(env.Pods) {
		defer podrecreation.CordonCleanup(n) //nolint:gocritic // The defer in loop is intentional, calling the cleanup function once per node
		err := podrecreation.CordonHelper(n, podrecreation.Cordon)
		if err != nil {
			logrus.Errorf("error cordoning the node: %s", n)
			ginkgo.Fail(fmt.Sprintf("Cordoning node %s failed with err: %s. Test inconclusive, skipping", n, err))
		}
		ginkgo.By(fmt.Sprintf("Draining and Cordoning node %s: ", n))
		logrus.Debugf("node: %s cordoned", n)
		count, err := podrecreation.CountPodsWithDelete(env.Pods, n, podrecreation.NoDelete)
		if err != nil {
			ginkgo.Fail(fmt.Sprintf("Getting pods list to drain in node %s failed with err: %s. Test inconclusive.", n, err))
		}
		nodeTimeout := timeoutPodSetReady + timeoutPodRecreationPerPod*time.Duration(count)
		logrus.Debugf("draining node: %s with timeout: %s", n, nodeTimeout.String())
		_, err = podrecreation.CountPodsWithDelete(env.Pods, n, podrecreation.DeleteForeground)
		if err != nil {
			ginkgo.Fail(fmt.Sprintf("Draining node %s failed with err: %s. Test inconclusive", n, err))
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

	for _, put := range env.GuaranteedPods {
		if !put.IsCPUIsolationCompliant() {
			podsMissingIsolationRequirements[put.Name] = true
		}
	}

	if n := len(podsMissingIsolationRequirements); n > 0 {
		errMsg := fmt.Sprintf("Number of guaranteed pods found that are not compliant with CPU isolation requirements: %d. See logs for more detail.", n)
		tnf.ClaimFilePrintf(errMsg)
		ginkgo.Fail(errMsg)
	}
}

func testAffinityRequiredPods(env *provider.TestEnvironment) {
	testhelper.SkipIfEmptyAny(ginkgo.Skip, env.AffinityRequiredPods)

	var podsDesiringAffinityRequiredMissingLabel []*provider.Pod
	for _, put := range env.AffinityRequiredPods {
		// Check if the pod is Affinity compliant.
		result, err := put.IsAffinityCompliant()
		if !result {
			tnf.ClaimFilePrintf(err.Error())
			podsDesiringAffinityRequiredMissingLabel = append(podsDesiringAffinityRequiredMissingLabel, put)
		}
	}

	testhelper.AddTestResultLog("Non-compliant", podsDesiringAffinityRequiredMissingLabel, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testAffinityRequiredDeployments(env *provider.TestEnvironment) {
	testhelper.SkipIfEmptyAny(ginkgo.Skip, env.AffinityRequiredDeployments)

	var deploymentsDesiringAffinityRequiredMissingLabel []*provider.Deployment
	for _, dep := range env.AffinityRequiredDeployments {
		// Check if the deployment is Affinity compliant.
		result, err := dep.IsAffinityCompliant()
		if !result {
			tnf.ClaimFilePrintf(err.Error())
			deploymentsDesiringAffinityRequiredMissingLabel = append(deploymentsDesiringAffinityRequiredMissingLabel, dep)
		}
	}

	testhelper.AddTestResultLog("Non-compliant", deploymentsDesiringAffinityRequiredMissingLabel, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testAffinityRequiredStatefulSets(env *provider.TestEnvironment) {
	testhelper.SkipIfEmptyAny(ginkgo.Skip, env.AffinityRequiredStatefulSets)

	var ssDesiringAffinityRequiredMissingLabel []*provider.StatefulSet
	for _, ss := range env.AffinityRequiredStatefulSets {
		// Check if the statefulset is Affinity compliant.
		result, err := ss.IsAffinityCompliant()
		if !result {
			tnf.ClaimFilePrintf(err.Error())
			ssDesiringAffinityRequiredMissingLabel = append(ssDesiringAffinityRequiredMissingLabel, ss)
		}
	}

	testhelper.AddTestResultLog("Non-compliant", ssDesiringAffinityRequiredMissingLabel, tnf.ClaimFilePrintf, ginkgo.Fail)
}
