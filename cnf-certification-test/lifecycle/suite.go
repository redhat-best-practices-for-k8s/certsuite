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
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
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
		testhelper.SkipIfEmptyAll(ginkgo.Skip, env.Deployments, env.StatetfulSets)
		testHighAvailability(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodNodeSelectorAndAffinityBestPractices)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if env.GetWorkerCount() < minWorkerNodesForLifecycle {
			ginkgo.Skip("Skipping pod scheduling test because invalid number of available workers.")
		}
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
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

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodRequestsAndLimitsIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		testPodRequestsAndLimits(&env)
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
		testPodPersistentVolumeReclaimPolicy(&env)
	})

})

func testContainersPreStop(env *provider.TestEnvironment) {
	badcontainers := []string{}
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " pre stop lifecycle ")

		if cut.Data.Lifecycle == nil || (cut.Data.Lifecycle != nil && cut.Data.Lifecycle.PreStop == nil) {
			badcontainers = append(badcontainers, cut.Data.Name)
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
		if cut.Data.ImagePullPolicy != corev1.PullIfNotPresent {
			badcontainers = append(badcontainers, "{"+cut.String()+": is using"+string(cut.Data.ImagePullPolicy)+"}")
			logrus.Errorln("container ", cut.Data.Name, " is using ", cut.Data.ImagePullPolicy, " as image policy")
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
		if cut.Data.ReadinessProbe == nil {
			badcontainers = append(badcontainers, cut.String())
			logrus.Errorln("container ", cut.Data.Name, " does not have ReadinessProbe defined")
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
		if cut.Data.LivenessProbe == nil {
			badcontainers = append(badcontainers, cut.String())
			logrus.Errorln("container ", cut.Data.Name, " does not have livenessProbe defined")
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
		if cut.Data.StartupProbe == nil {
			badcontainers = append(badcontainers, cut.String())
			logrus.Errorln("container ", cut.Data.Name, " does not have startupProbe defined")
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
		logrus.Debugln("check pod ", put.Data.Namespace, " ", put.Data.Name, " owner reference")
		o := ownerreference.NewOwnerReference(put.Data)
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
	for _, put := range env.Pods {
		if len(put.Data.Spec.NodeSelector) != 0 {
			tnf.ClaimFilePrintf("ERROR: %s has a node selector clause. Node selector: %v", put, &put.Data.Spec.NodeSelector)
			badPods = append(badPods, put.Data)
		}
		if put.Data.Spec.Affinity != nil && put.Data.Spec.Affinity.NodeAffinity != nil {
			tnf.ClaimFilePrintf("ERROR: %s has a node affinity clause. Node affinity: %v", put, put.Data.Spec.Affinity.NodeAffinity)
			badPods = append(badPods, put.Data)
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
		deployment := env.Deployments[i]
		ns, name := deployment.Namespace, deployment.Name
		key := ns + name
		if hpa, ok := env.HorizontalScaler[key]; ok {
			// if the deployment is controller by
			// horizontal scaler, then test that scaler
			// can scale the deployment
			if !scaling.TestScaleHpaDeployment(deployment, hpa, timeout) {
				failedDeployments = append(failedDeployments, provider.DeploymentToString(deployment))
			}
			continue
		}
		// if the deployment is not controller by HPA
		// scale it directly
		if !scaling.TestScaleDeployment(deployment, timeout) {
			failedDeployments = append(failedDeployments, provider.DeploymentToString(deployment))
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
		statefulset := env.StatetfulSets[i]
		ns, name := statefulset.Namespace, statefulset.Name
		key := ns + name
		if hpa, ok := env.HorizontalScaler[key]; ok {
			// if the statefulset is controller by
			// horizontal scaler, then test that scaler
			// can scale the statefulset
			if !scaling.TestScaleHpaStatefulSet(statefulset, hpa, timeout) {
				failedStatetfulSets = append(failedStatetfulSets, provider.StatefulsetToString(statefulset))
			}
			continue
		}
		// if the statefulset is not controller by HPA
		// scale it directly
		if !scaling.TestScaleStatefulSet(statefulset, timeout) {
			failedStatetfulSets = append(failedStatetfulSets, provider.StatefulsetToString(statefulset))
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
	for _, dp := range env.Deployments {
		if dp.Spec.Replicas == nil || *(dp.Spec.Replicas) <= 1 {
			badDeployments = append(badDeployments, provider.DeploymentToString(dp))
			continue
		}
		if dp.Spec.Template.Spec.Affinity == nil ||
			dp.Spec.Template.Spec.Affinity.PodAntiAffinity == nil {
			badDeployments = append(badDeployments, provider.DeploymentToString(dp))
		}
	}
	for _, st := range env.StatetfulSets {
		if st.Spec.Replicas == nil || *(st.Spec.Replicas) <= 1 {
			badStatefulSet = append(badStatefulSet, provider.StatefulsetToString(st))
			continue
		}
		if st.Spec.Template.Spec.Affinity == nil ||
			st.Spec.Template.Spec.Affinity.PodAntiAffinity == nil {
			badDeployments = append(badDeployments, provider.StatefulsetToString(st))
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
		count, err := podrecreation.CountPodsWithDelete(n, podrecreation.NoDelete)
		if err != nil {
			ginkgo.Fail(fmt.Sprintf("Getting pods list to drain in node %s failed with err: %s. Test inconclusive.", n, err))
		}
		nodeTimeout := timeoutPodSetReady + timeoutPodRecreationPerPod*time.Duration(count)
		logrus.Debugf("draining node: %s with timeout: %s", n, nodeTimeout.String())
		_, err = podrecreation.CountPodsWithDelete(n, podrecreation.DeleteForeground)
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
	}
}

func testPodRequestsAndLimits(env *provider.TestEnvironment) {
	var containersMissingRequestsOrLimits []string
	ginkgo.By("Testing container resource requests and limits")

	// Loop through the containers, looking for containers that are missing requests or limits.
	// These need to be defined in order to pass.
	for _, cut := range env.Containers {
		badContainer := false
		// Parse the limits.
		if len(cut.Data.Resources.Limits) == 0 {
			tnf.ClaimFilePrintf("Container has been found missing resource limits: %s", cut.String())
			badContainer = true
		} else {
			if cut.Data.Resources.Limits.Cpu() == nil {
				tnf.ClaimFilePrintf("Container has been found missing CPU limits: %s", cut.String())
				badContainer = true
			}

			if cut.Data.Resources.Limits.Memory() == nil {
				tnf.ClaimFilePrintf("Container has been found missing memory limits: %s", cut.String())
				badContainer = true
			}
		}

		// Parse the requests.
		if len(cut.Data.Resources.Requests) == 0 {
			tnf.ClaimFilePrintf("Container has been found missing resource requests: %s", cut.String())
			badContainer = true
		} else {
			if cut.Data.Resources.Requests.Cpu() == nil {
				tnf.ClaimFilePrintf("Container has been found missing CPU requests: %s", cut.String())
				badContainer = true
			}

			if cut.Data.Resources.Requests.Memory() == nil {
				tnf.ClaimFilePrintf("Container has been found missing memory requests: %s", cut.String())
				badContainer = true
			}
		}

		if badContainer {
			containersMissingRequestsOrLimits = append(containersMissingRequestsOrLimits, cut.String())
		}
	}

	if n := len(containersMissingRequestsOrLimits); n > 0 {
		errMsg := fmt.Sprintf("Containers found that are missing either resource requests or limits: %d. See logs for more detail.", n)
		tnf.ClaimFilePrintf(errMsg)
		ginkgo.Fail(errMsg)
	}
}

func testPodPersistentVolumeReclaimPolicy(env *provider.TestEnvironment) {
	ginkgo.By("Testing PersistentVolumes for reclaim policy to be set to delete")
	var persistentVolumesBadReclaim []string

	// Look through all of the pods, matching their persistent volumes to the list of overall cluster PVs and checking their reclaim status.
	for _, put := range env.Pods {
		for index := range env.PersistentVolumes {
			for pvIndex := range put.Data.Spec.Volumes {
				if put.Data.Spec.Volumes[pvIndex].Name == env.PersistentVolumes[index].Name && env.PersistentVolumes[index].Spec.PersistentVolumeReclaimPolicy != corev1.PersistentVolumeReclaimDelete {
					persistentVolumesBadReclaim = append(persistentVolumesBadReclaim, env.PersistentVolumes[index].Name)
					tnf.ClaimFilePrintf("Persistent Volume: %s in namespace %s has been found without a reclaim policy of DELETE.", env.PersistentVolumes[index].Name, env.PersistentVolumes[index].Namespace)
				}
			}
		}
	}

	if n := len(persistentVolumesBadReclaim); n > 0 {
		errMsg := fmt.Sprintf("Persistent Volumes found that are missing a reclaim policy of DELETE: %d. See logs for more detail.", n)
		tnf.ClaimFilePrintf(errMsg)
		ginkgo.Fail(errMsg)
	}
}
