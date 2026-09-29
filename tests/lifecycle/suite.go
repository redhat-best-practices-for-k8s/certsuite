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
		WithBeforeEachFn(checksdb.DefaultBeforeEachFn(func() { env = provider.GetTestEnvironment() }))

	// Prestop test
	checksadapter.AddCheck(checksGroup, "lifecycle-container-prestop", checksfn.CheckPreStop, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	// Scale CRD test
	checksadapter.AddIntrusiveCheck(checksGroup, "lifecycle-crd-scaling", checksfn.CheckCRDScaling, &env,
		testhelper.GetNoCrdsUnderTestSkipFn(&env),
		testhelper.GetNotIntrusiveSkipFn(&env))

	// Poststart test
	checksadapter.AddCheck(checksGroup, "lifecycle-container-poststart", checksfn.CheckPostStart, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	// Image pull policy test
	checksadapter.AddCheck(checksGroup, "lifecycle-image-pull-policy", checksfn.CheckImagePullPolicy, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	// Readiness probe test
	checksadapter.AddCheck(checksGroup, "lifecycle-readiness-probe", checksfn.CheckReadinessProbe, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	// Liveness probe test
	checksadapter.AddCheck(checksGroup, "lifecycle-liveness-probe", checksfn.CheckLivenessProbe, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	// Startup probe test
	checksadapter.AddCheck(checksGroup, "lifecycle-startup-probe", checksfn.CheckStartupProbe, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	// Pod owner reference test
	checksadapter.AddCheck(checksGroup, "lifecycle-pod-owner-type", checksfn.CheckPodOwnerType, &env,
		testhelper.GetNoPodsUnderTestSkipFn(&env))

	// High availability test
	checksadapter.AddCheck(checksGroup, "lifecycle-pod-high-availability", checksfn.CheckHighAvailability, &env,
		testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle),
		skipIfNoPodSetsetsUnderTest)

	// Selector and affinity best practices test
	checksadapter.AddCheck(checksGroup, "lifecycle-pod-scheduling", checksfn.CheckPodScheduling, &env,
		testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle),
		testhelper.GetPodsWithoutAffinityRequiredLabelSkipFn(&env))

	// Pod recreation test
	checksadapter.AddIntrusiveCheck(checksGroup, "lifecycle-pod-recreation", checksfn.CheckPodRecreation, &env,
		testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle),
		testhelper.GetNotIntrusiveSkipFn(&env),
		skipIfNoPodSetsetsUnderTest)

	// Deployment scaling test
	checksadapter.AddIntrusiveCheck(checksGroup, "lifecycle-deployment-scaling", checksfn.CheckDeploymentScaling, &env,
		testhelper.GetNotIntrusiveSkipFn(&env),
		testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle),
		skipIfNoPodSetsetsUnderTest)

	// Statefulset scaling test
	checksadapter.AddIntrusiveCheck(checksGroup, "lifecycle-statefulset-scaling", checksfn.CheckStatefulSetScaling, &env,
		testhelper.GetNotIntrusiveSkipFn(&env),
		testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle),
		skipIfNoPodSetsetsUnderTest)

	// Persistent volume reclaim policy test
	checksadapter.AddCheck(checksGroup, "lifecycle-persistent-volume-reclaim-policy", checksfn.CheckPVReclaimPolicy, &env,
		testhelper.GetNoPersistentVolumesSkipFn(&env),
		testhelper.GetNoPodsUnderTestSkipFn(&env))

	// CPU Isolation test
	checksadapter.AddCheck(checksGroup, "lifecycle-cpu-isolation", checksfn.CheckCPUIsolation, &env,
		testhelper.GetNoGuaranteedPodsWithExclusiveCPUsSkipFn(&env))

	// Affinity required pods test
	checksadapter.AddCheck(checksGroup, "lifecycle-affinity-required-pods", checksfn.CheckAffinityRequired, &env,
		testhelper.GetNoAffinityRequiredPodsSkipFn(&env))

	// Pod toleration bypass test
	checksadapter.AddCheck(checksGroup, "lifecycle-pod-toleration-bypass", checksfn.CheckTolerationBypass, &env,
		testhelper.GetNoPodsUnderTestSkipFn(&env))

	// Storage provisioner test
	checksadapter.AddCheck(checksGroup, "lifecycle-storage-provisioner", checksfn.CheckStorageProvisioner, &env,
		testhelper.GetNoPodsUnderTestSkipFn(&env),
		testhelper.GetNoStorageClassesSkipFn(&env),
		testhelper.GetNoPersistentVolumeClaimsSkipFn(&env))

	// Topology Spread Constraint test
	checksadapter.AddCheck(checksGroup, "lifecycle-topology-spread-constraint", checksfn.CheckTopologySpreadConstraints, &env,
		testhelper.GetNoDeploymentsUnderTestSkipFn(&env),
		testhelper.GetNotEnoughWorkersSkipFn(&env, minWorkerNodesForLifecycle))
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
