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
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/graceperiod"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/ownerreference"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/lifecycle/scaling"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"

	v1 "k8s.io/api/core/v1"
)

const (
	timeout = 60 * time.Second
)

//
// All actual test code belongs below here.  Utilities belong above.
//
var _ = ginkgo.Describe(common.LifecycleTestKey, func() {
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})
	testContainersPreStop(&env)
	testContainersImagePolicy(&env)
	testContainersReadinessProbe(&env)
	testContainersLivenessProbe(&env)
	testPodsOwnerReference(&env)
	testHighAvailability(&env)

	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodNodeSelectorAndAffinityBestPractices)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testPodNodeSelectorAndAffinityBestPractices(&env)
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestNonDefaultGracePeriodIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testGracePeriod(&env)
	})

	if env.IsIntrusive() {
		testScaling(&env, timeout)
	}
})

func testContainersPreStop(env *provider.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestShudtownIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		badcontainers := []string{}
		for _, cut := range env.Containers {
			logrus.Debugln("check container ", cut.Namespace, " ", cut.Podname, " ", cut.Data.Name, " pre stop lifecycle ")

			if cut.Data.Lifecycle == nil || (cut.Data.Lifecycle != nil && cut.Data.Lifecycle.PreStop == nil) {
				badcontainers = append(badcontainers, cut.Data.Name)
				tnf.ClaimFilePrintf("container %s does not have preStop defined", cut.StringShort())
			}
		}
		if len(badcontainers) > 0 {
			tnf.ClaimFilePrintf("bad containers %v", badcontainers)
		}
		gomega.Expect(0).To(gomega.Equal(len(badcontainers)))
	})
}

func testContainersImagePolicy(env *provider.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestImagePullPolicyIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		badcontainers := []string{}
		for _, cut := range env.Containers {
			logrus.Debugln("check container ", cut.Namespace, " ", cut.Podname, " ", cut.Data.Name, " pull policy, should be ", v1.PullIfNotPresent)
			if cut.Data.ImagePullPolicy != v1.PullIfNotPresent {
				s := cut.Namespace + ":" + cut.Podname + ":" + cut.Data.Name
				badcontainers = append(badcontainers, s)
				logrus.Errorln("container ", cut.Data.Name, " is using ", cut.Data.ImagePullPolicy, " as image policy")
			}
		}
		if len(badcontainers) > 0 {
			tnf.ClaimFilePrintf("bad containers %v", badcontainers)
		}
		gomega.Expect(0).To(gomega.Equal(len(badcontainers)))
	})
}

//nolint:dupl
func testContainersReadinessProbe(env *provider.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestReadinessProbeIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		badcontainers := []string{}
		for _, cut := range env.Containers {
			logrus.Debugln("check container ", cut.Namespace, " ", cut.Podname, " ", cut.Data.Name, " readiness probe ")
			if cut.Data.ReadinessProbe == nil {
				s := cut.Namespace + ":" + cut.Podname + ":" + cut.Data.Name
				badcontainers = append(badcontainers, s)
				logrus.Errorln("container ", cut.Data.Name, " does not have ReadinessProbe defined")
			}
		}
		if len(badcontainers) > 0 {
			tnf.ClaimFilePrintf("bad containers %v", badcontainers)
		}
		gomega.Expect(0).To(gomega.Equal(len(badcontainers)))
	})
}

//nolint:dupl
func testContainersLivenessProbe(env *provider.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestLivenessProbeIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		badcontainers := []string{}
		for _, cut := range env.Containers {
			logrus.Debugln("check container ", cut.Namespace, " ", cut.Podname, " ", cut.Data.Name, " liveness probe ")
			if cut.Data.LivenessProbe == nil {
				s := cut.Namespace + ":" + cut.Podname + ":" + cut.Data.Name
				badcontainers = append(badcontainers, s)
				logrus.Errorln("container ", cut.Data.Name, " does not have livenessProbe defined")
			}
		}
		if len(badcontainers) > 0 {
			tnf.ClaimFilePrintf("bad containers %v", badcontainers)
		}
		gomega.Expect(0).To(gomega.Equal(len(badcontainers)))
	})
}

func testPodsOwnerReference(env *provider.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodDeploymentBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		ginkgo.By("Testing owners of CNF pod, should be replicas Set")
		badPods := []string{}
		for _, put := range env.Pods {
			logrus.Debugln("check pod ", put.Namespace, " ", put.Name, " owner reference")
			o := ownerreference.NewOwnerReference(put)
			o.RunTest()
			if o.GetResults() != testhelper.SUCCESS {
				s := put.Namespace + ":" + put.Name
				badPods = append(badPods, s)
			}
		}
		if len(badPods) > 0 {
			tnf.ClaimFilePrintf("bad containers %v", badPods)
		}
		gomega.Expect(0).To(gomega.Equal(len(badPods)))
	})
}

func testPodNodeSelectorAndAffinityBestPractices(env *provider.TestEnvironment) {
	var badPods []*v1.Pod
	for _, put := range env.Pods {
		if len(put.Spec.NodeSelector) != 0 {
			tnf.ClaimFilePrintf("ERROR: Pod: %s has a node selector clause. Node selector: %v", provider.PodToString(put), &put.Spec.NodeSelector)
			badPods = append(badPods, put)
		}
		if put.Spec.Affinity != nil && put.Spec.Affinity.NodeAffinity != nil {
			tnf.ClaimFilePrintf("ERROR: Pod: %s has a node affinity clause. Node affinity: %v", provider.PodToString(put), put.Spec.Affinity.NodeAffinity)
			badPods = append(badPods, put)
		}
	}
	if n := len(badPods); n > 0 {
		logrus.Debugf("Pods with nodeSelector/nodeAffinity: %+v", badPods)
		ginkgo.Fail(fmt.Sprintf("%d pods found with nodeSelector/nodeAffinity rules", n))
	}
}

func testGracePeriod(env *provider.TestEnvironment) {
	badDeployments, deploymentLogs := graceperiod.TestTerminationGracePeriodOnDeployments(env)
	badStatefulsets, statefulsetLogs := graceperiod.TestTerminationGracePeriodOnStatefulsets(env)
	badPods, podLogs := graceperiod.TestTerminationGracePeriodOnPods(env)

	numDeps := len(badDeployments)
	if numDeps > 0 {
		logrus.Debugf("Deployments found without terminationGracePeriodSeconds param set: %+v", badDeployments)
	}
	numSts := len(badStatefulsets)
	if numSts > 0 {
		logrus.Debugf("Statefulsets found without terminationGracePeriodSeconds param set: %+v", badStatefulsets)
	}
	numPods := len(badPods)
	if numPods > 0 {
		logrus.Debugf("Pods found without terminationGracePeriodSeconds param set: %+v", badPods)
	}
	ginkgo.By("Test results for grace period on deployments")
	tnf.ClaimFilePrintf("%s", deploymentLogs)
	ginkgo.By("Test results for grace period on statefulsets")
	tnf.ClaimFilePrintf("%s", statefulsetLogs)
	ginkgo.By("Test results for grace period on unmanaged pods")
	tnf.ClaimFilePrintf("%s", podLogs)

	if numDeps > 0 || numSts > 0 || numPods > 0 {
		ginkgo.Fail(fmt.Sprintf("Found %d deployments, %d statefulsets and %d pods without terminationGracePeriodSeconds param set.", numDeps, numSts, numPods))
	}
}

func testScaling(env *provider.TestEnvironment, timeout time.Duration) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestDeploymentScalingIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		ginkgo.By("Testing deployment scaling")
		defer env.SetNeedsRefresh()

		if len(env.Deployments) == 0 {
			ginkgo.Skip("No test deployments found.")
		}
		failedDeployments := []string{}
		skippedDeployments := []string{}
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
					failedDeployments = append(failedDeployments, name)
				}
				continue
			}
			// if the deployment is not controller by HPA
			// scale it directly
			if !scaling.TestScaleDeployment(deployment, timeout) {
				failedDeployments = append(failedDeployments, name)
			}
		}

		if len(skippedDeployments) > 0 {
			tnf.ClaimFilePrintf("not ready deployments : %v", skippedDeployments)
		}
		if len(failedDeployments) > 0 {
			tnf.ClaimFilePrintf(" failed deployments: %v", failedDeployments)
		}
		gomega.Expect(0).To(gomega.Equal(len(failedDeployments)))
		gomega.Expect(0).To(gomega.Equal(len(skippedDeployments)))
	})
}

// testHighAvailability
func testHighAvailability(env *provider.TestEnvironment) {
	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestPodHighAvailabilityBestPractices)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		ginkgo.By("Should set pod replica number greater than 1")
		if len(env.Deployments) == 0 && len(env.SatetfulSets) == 0 {
			ginkgo.Skip("No test deployments/statefulset found.")
		}

		badDeployments := []string{}
		badStatefulSet := []string{}
		for _, dp := range env.Deployments {
			if dp.Spec.Replicas == nil || *(dp.Spec.Replicas) == 1 {
				badDeployments = append(badDeployments, provider.DeploymentToString(dp))
			}
		}
		for _, st := range env.SatetfulSets {
			if st.Spec.Replicas == nil || *(st.Spec.Replicas) == 1 {
				badStatefulSet = append(badStatefulSet, provider.StatefultsetToString(st))
			}
		}

		if n := len(badDeployments); n > 0 {
			logrus.Errorf("Deployments without a valid high availability : %+v", badDeployments)
			tnf.ClaimFilePrintf("Deployments without a valid high availability : %+v", badDeployments)
		}
		if n := len(badStatefulSet); n > 0 {
			logrus.Errorf("Statefulset without a valid podAntiAffinity rule: %+v", badStatefulSet)
			tnf.ClaimFilePrintf("Statefulset without a valid podAntiAffinity rule: %+v", badStatefulSet)
		}
		gomega.Expect(0).To(gomega.Equal(len(badDeployments)))
		gomega.Expect(0).To(gomega.Equal(len(badStatefulSet)))
	})
}
