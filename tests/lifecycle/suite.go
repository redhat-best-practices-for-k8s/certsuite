// Copyright (C) 2020-2026 Red Hat, Inc.
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
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksadapter"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/configuration"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	checksfn "github.com/redhat-best-practices-for-k8s/checks/lifecycle"
)

const (
	minWorkerNodesForLifecycle = 2
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

//nolint:funlen
func LoadChecks() {
	log.Debug("Loading %s suite checks", common.LifecycleTestKey)

	checksGroup := checksdb.NewChecksGroup(common.LifecycleTestKey).
		WithBeforeEachFn(beforeEachFn)

	// Prestop test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-container-prestop")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckPreStop).MakeCheckFn(&env)))

	// Scale CRD test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-crd-scaling")).
		WithSkipCheckFn(
			testhelper.GetNoCrdsUnderTestSkipFn(&env),
			testhelper.GetNotIntrusiveSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckCRDScaling).MakeIntrusiveCheckFn(&env)))

	// Poststart test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-container-poststart")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckPostStart).MakeCheckFn(&env)))

	// Image pull policy test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-image-pull-policy")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckImagePullPolicy).MakeCheckFn(&env)))

	// Readiness probe test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-readiness-probe")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckReadinessProbe).MakeCheckFn(&env)))

	// Liveness probe test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-liveness-probe")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckLivenessProbe).MakeCheckFn(&env)))

	// Startup probe test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-startup-probe")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckStartupProbe).MakeCheckFn(&env)))

	// Pod owner reference test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-pod-owner-type")).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckPodOwnerType).MakeCheckFn(&env)))

	// High availability test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-pod-high-availability")).
		WithSkipCheckFn(testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle)).
		WithSkipCheckFn(skipIfNoPodSetsetsUnderTest).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckHighAvailability).MakeCheckFn(&env)))

	// Selector and affinity best practices test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-pod-scheduling")).
		WithSkipCheckFn(
			testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle),
			testhelper.GetPodsWithoutAffinityRequiredLabelSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckPodScheduling).MakeCheckFn(&env)))

	// Pod recreation test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-pod-recreation")).
		WithSkipCheckFn(
			testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle),
			testhelper.GetNotIntrusiveSkipFn(&env)).
		WithSkipCheckFn(skipIfNoPodSetsetsUnderTest).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckPodRecreation).MakeIntrusiveCheckFn(&env)))

	// Deployment scaling test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-deployment-scaling")).
		WithSkipCheckFn(
			testhelper.GetNotIntrusiveSkipFn(&env),
			testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle)).
		WithSkipCheckFn(skipIfNoPodSetsetsUnderTest).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckDeploymentScaling).MakeIntrusiveCheckFn(&env)))

	// Statefulset scaling test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-statefulset-scaling")).
		WithSkipCheckFn(
			testhelper.GetNotIntrusiveSkipFn(&env),
			testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle)).
		WithSkipCheckFn(skipIfNoPodSetsetsUnderTest).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckStatefulSetScaling).MakeIntrusiveCheckFn(&env)))

	// Persistent volume reclaim policy test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-persistent-volume-reclaim-policy")).
		WithSkipCheckFn(
			testhelper.GetNoPersistentVolumesSkipFn(&env),
			testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckPVReclaimPolicy).MakeCheckFn(&env)))

	// CPU Isolation test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-cpu-isolation")).
		WithSkipCheckFn(testhelper.GetNoGuaranteedPodsWithExclusiveCPUsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckCPUIsolation).MakeCheckFn(&env)))

	// Affinity required pods test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-affinity-required-pods")).
		WithSkipCheckFn(testhelper.GetNoAffinityRequiredPodsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckAffinityRequired).MakeCheckFn(&env)))

	// Pod toleration bypass test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-pod-toleration-bypass")).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckTolerationBypass).MakeCheckFn(&env)))

	// Storage provisioner test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-storage-provisioner")).
		WithSkipCheckFn(
			testhelper.GetNoPodsUnderTestSkipFn(&env),
			testhelper.GetNoStorageClassesSkipFn(&env),
			testhelper.GetNoPersistentVolumeClaimsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckStorageProvisioner).MakeCheckFn(&env)))

	// Topology Spread Constraint test
	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("lifecycle-topology-spread-constraint")).
		WithSkipCheckFn(
			testhelper.GetNoDeploymentsUnderTestSkipFn(&env),
			testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckTopologySpreadConstraints).MakeCheckFn(&env)))
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
