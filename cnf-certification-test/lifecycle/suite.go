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
	intrusiveTcSkippedReason   = "This is an intrusive test case and the env var TNF_NON_INTRUSIVE_ONLY was set"
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
		testhelper.SkipIfEmptyAll(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		testContainersPreStop(&env)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestCrdScalingIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if !env.IsIntrusive() {
			ginkgo.Skip(intrusiveTcSkippedReason)
		}
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.ScaleCrUnderTest, "env.ScaleCrUnderTest"))
		// Note: We skip this test because 'testHighAvailability' in the lifecycle suite is already
		// testing the replicas and antiaffinity rules that should already be in place for crd.
		testScaleCrd(&env, timeout)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestStartupIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAll(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		testContainersPostStart(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestImagePullPolicyIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAll(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		testContainersImagePolicy(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestReadinessProbeIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAll(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		testContainersReadinessProbe(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestLivenessProbeIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAll(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		testContainersLivenessProbe(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestStartupProbeIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAll(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
		testContainersStartupProbe(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodDeploymentBestPracticesIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAll(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"))
		testPodsOwnerReference(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodHighAvailabilityBestPractices)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if env.GetWorkerCount() < minWorkerNodesForLifecycle {
			ginkgo.Skip("Skipping pod high availability test because invalid number of available workers.")
		}
		testhelper.SkipIfEmptyAll(ginkgo.Skip, testhelper.NewSkipObject(env.Deployments, "env.Deployments"), testhelper.NewSkipObject(env.StatefulSets, "env.StatefulSets"))
		testHighAvailability(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodNodeSelectorAndAffinityBestPractices)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if env.GetWorkerCount() < minWorkerNodesForLifecycle {
			ginkgo.Skip("Skipping pod scheduling test because invalid number of available workers.")
		}
		testPods := env.GetPodsWithoutAffinityRequiredLabel()
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(testPods, "testPods"))
		testPodNodeSelectorAndAffinityBestPractices(testPods)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodRecreationIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if !env.IsIntrusive() {
			ginkgo.Skip(intrusiveTcSkippedReason)
		}
		testhelper.SkipIfEmptyAll(ginkgo.Skip, testhelper.NewSkipObject(env.Deployments, "env.Deployments"), testhelper.NewSkipObject(env.StatefulSets, "env.StatefulSets"))
		if env.GetWorkerCount() < minWorkerNodesForLifecycle {
			ginkgo.Skip("Skipping pod recreation scaling test because invalid number of available workers.")
		}
		// Testing pod re-creation for deployments
		testPodsRecreation(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestDeploymentScalingIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if !env.IsIntrusive() {
			ginkgo.Skip(intrusiveTcSkippedReason)
		}
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Deployments, "env.Deployments"))
		if env.GetWorkerCount() < minWorkerNodesForLifecycle {
			// Note: We skip this test because 'testHighAvailability' in the lifecycle suite is already
			// testing the replicas and antiaffinity rules that should already be in place for deployments.
			ginkgo.Skip("Skipping deployment scaling test because invalid number of available workers.")
		}
		testDeploymentScaling(&env, timeout)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestStateFulSetScalingIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		if !env.IsIntrusive() {
			ginkgo.Skip(intrusiveTcSkippedReason)
		}
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.StatefulSets, "env.StatefulSets"))
		if env.GetWorkerCount() < minWorkerNodesForLifecycle {
			// Note: We skip this test because 'testHighAvailability' in the lifecycle suite is already
			// testing the replicas and antiaffinity rules that should already be in place for statefulset.
			ginkgo.Skip("Skipping statefulset scaling test because invalid number of available workers.")
		}
		testStatefulSetScaling(&env, timeout)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPersistentVolumeReclaimPolicyIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"), testhelper.NewSkipObject(env.PersistentVolumes, "env.PersistentVolumes"))
		testPodPersistentVolumeReclaimPolicy(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestCPUIsolationIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.GetGuaranteedPodsWithExclusiveCPUs(), "env.GetGuaranteedPodsWithExclusiveCPUs()"))
		testCPUIsolation(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestAffinityRequiredPods)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testAffinityRequiredPods(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestPodTolerationBypassIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"))
		testPodTolerationBypass(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestStorageProvisioner)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.Pods, "env.Pods"))
		testStorageProvisioner(&env)
	})
})

func testContainersPreStop(env *provider.TestEnvironment) {
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
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testContainersPostStart(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range env.Containers {
		logrus.Debugln("check container ", cut.String(), " post start lifecycle ")

		if cut.Lifecycle == nil || (cut.Lifecycle != nil && cut.Lifecycle.PostStart == nil) {
			tnf.ClaimFilePrintf("%s does not have postStart defined", cut)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container does not have postStart defined", false))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container has postStart defined."+
				"Attention: There is a known upstream bug where a pod with a still-running postStart lifecycle hook that is deleted may not be terminated even after "+
				"the terminationGracePeriod k8s bug link: kubernetes/kubernetes#116032", true))
		}
	}
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testContainersImagePolicy(env *provider.TestEnvironment) {
	testhelper.SkipIfEmptyAll(ginkgo.Skip, testhelper.NewSkipObject(env.Containers, "env.Containers"))
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
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testContainersReadinessProbe(env *provider.TestEnvironment) {
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
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testContainersLivenessProbe(env *provider.TestEnvironment) {
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
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testContainersStartupProbe(env *provider.TestEnvironment) {
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
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testPodsOwnerReference(env *provider.TestEnvironment) {
	ginkgo.By("Testing owners of CNF pod, should be replicas Set")
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
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testPodNodeSelectorAndAffinityBestPractices(testPods []*provider.Pod) {
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
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
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

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testScaleCrd(env *provider.TestEnvironment, timeout time.Duration) {
	ginkgo.By("Testing custom resource scaling")
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
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

//nolint:dupl
func testStatefulSetScaling(env *provider.TestEnvironment, timeout time.Duration) {
	ginkgo.By("Testing statefulset scaling")
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

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

// testHighAvailability
func testHighAvailability(env *provider.TestEnvironment) {
	ginkgo.By("Should set pod replica number greater than 1")

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

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
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

	// Before draining any node, wait until all podsets are ready. The timeout depends on the number of podsets to check.
	// timeout = k-mins + (1min * (num-deployments + num-statefulsets))
	allPodsetsReadyTimeout := timeoutPodSetReady + time.Minute*time.Duration(len(env.Deployments)+len(env.StatefulSets))
	claimsLog, atLeastOnePodsetNotReady := podsets.WaitForAllPodSetsReady(env, allPodsetsReadyTimeout)
	if atLeastOnePodsetNotReady {
		tnf.ClaimFilePrintf("%s", claimsLog.GetLogLines())
		ginkgo.Fail("Some deployments or stateful sets are not in a good initial state. Cannot perform test.")
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
		testhelper.AddTestResultLog("Non-compliant", podsWithNodeAssignment, tnf.ClaimFilePrintf, ginkgo.Fail)
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
		logrus.Debugf("draining node: %s with timeout: %s", n, nodeTimeout)
		_, err = podrecreation.CountPodsWithDelete(env.Pods, n, podrecreation.DeleteForeground)
		if err != nil {
			ginkgo.Fail(fmt.Sprintf("Draining node %s failed with err: %v. Test inconclusive", n, err))
		}

		claimsLog, podsNotReady := podsets.WaitForAllPodSetsReady(env, nodeTimeout)
		if podsNotReady {
			tnf.ClaimFilePrintf("%s", claimsLog.GetLogLines())
			ginkgo.Fail(fmt.Sprintf("Some pods are not ready after draining the node %s", n))
		}

		err = podrecreation.CordonHelper(n, podrecreation.Uncordon)
		if err != nil {
			logrus.Fatalf("error uncordoning the node: %s", n)
		}
	}

	// Reached end of TC, which means no ginkgo.Fail() was called.
	needsPostMortemInfo = false
}

func testPodPersistentVolumeReclaimPolicy(env *provider.TestEnvironment) {
	ginkgo.By("Testing PersistentVolumes for reclaim policy to be set to delete")
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

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testCPUIsolation(env *provider.TestEnvironment) {
	ginkgo.By("Testing pods for CPU isolation requirements")

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

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testAffinityRequiredPods(env *provider.TestEnvironment) {
	ginkgo.By("Testing affinity required pods for ")
	testhelper.SkipIfEmptyAny(ginkgo.Skip, testhelper.NewSkipObject(env.GetAffinityRequiredPods(), "env.GetAffinityRequiredPods()"))

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
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testPodTolerationBypass(env *provider.TestEnvironment) {
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

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

//nolint:funlen
func testStorageProvisioner(env *provider.TestEnvironment) {
	const localStorageProvisioner = "kubernetes.io/no-provisioner"
	const lvmProvisioner = "topolvm.io"
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	var StorageClasses = env.StorageClassList
	var Pvc = env.PersistentVolumeClaims
	snoSingleLocalStorageProvisionner := ""
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
						if Pvc[i].Spec.StorageClassName != nil && StorageClasses[j].Name == *Pvc[i].Spec.StorageClassName {
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
		}
	}
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}
