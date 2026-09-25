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

package platform

import (
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksadapter"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	checksfn "github.com/redhat-best-practices-for-k8s/checks/platform"
)

var env provider.TestEnvironment

func LoadChecks() {
	log.Debug("Loading %s suite checks", common.PlatformAlterationTestKey)

	checksGroup := checksdb.NewChecksGroup(common.PlatformAlterationTestKey).
		WithBeforeEachFn(checksdb.DefaultBeforeEachFn(func() { env = provider.GetTestEnvironment() }))

	checksadapter.AddCheck(checksGroup, "platform-alteration-hyperthread-enable", checksfn.CheckHyperthreadEnable, &env,
		testhelper.GetNoBareMetalNodesSkipFn(&env),
		testhelper.GetDaemonSetFailedToSpawnSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "platform-alteration-base-image", checksfn.CheckUnalteredBaseImage, &env,
		testhelper.GetNonOCPClusterSkipFn(),
		testhelper.GetDaemonSetFailedToSpawnSkipFn(&env),
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "platform-alteration-tainted-node-kernel", checksfn.CheckTainted, &env,
		testhelper.GetDaemonSetFailedToSpawnSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "platform-alteration-isredhat-release", checksfn.CheckIsRedHatRelease, &env,
		testhelper.GetNoContainersUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "platform-alteration-is-selinux-enforcing", checksfn.CheckSELinuxEnforcing, &env,
		testhelper.GetNonOCPClusterSkipFn(),
		testhelper.GetDaemonSetFailedToSpawnSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "platform-alteration-hugepages-config", checksfn.CheckHugepages, &env,
		testhelper.GetNonOCPClusterSkipFn(),
		testhelper.GetDaemonSetFailedToSpawnSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "platform-alteration-boot-params", checksfn.CheckBootParams, &env,
		testhelper.GetNonOCPClusterSkipFn(),
		testhelper.GetDaemonSetFailedToSpawnSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "platform-alteration-sysctl-config", checksfn.CheckSysctl, &env,
		testhelper.GetNonOCPClusterSkipFn(),
		testhelper.GetDaemonSetFailedToSpawnSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "platform-alteration-service-mesh-usage", checksfn.CheckServiceMeshUsage, &env,
		testhelper.GetNoIstioSkipFn(&env),
		testhelper.GetNoPodsUnderTestSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "platform-alteration-ocp-lifecycle", checksfn.CheckOCPLifecycle, &env,
		testhelper.GetNonOCPClusterSkipFn())

	checksadapter.AddCheck(checksGroup, "platform-alteration-ocp-node-os-lifecycle", checksfn.CheckOCPNodeOSLifecycle, &env,
		testhelper.GetNonOCPClusterSkipFn())

	checksadapter.AddCheck(checksGroup, "platform-alteration-hugepages-2m-only", checksfn.CheckHugepages2MiOnly, &env,
		testhelper.GetNonOCPClusterSkipFn(),
		testhelper.GetNoHugepagesPodsSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "platform-alteration-hugepages-1g-only", checksfn.CheckHugepages1GiOnly, &env,
		testhelper.GetNonOCPClusterSkipFn(),
		testhelper.GetNoHugepagesPodsSkipFn(&env))

	checksadapter.AddCheck(checksGroup, "platform-alteration-cluster-operator-health", checksfn.CheckClusterOperatorHealth, &env,
		testhelper.GetNonOCPClusterSkipFn())
}
