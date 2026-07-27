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

var (
	env provider.TestEnvironment

	beforeEachFn = func(check *checksdb.Check) error {
		env = provider.GetTestEnvironment()
		return nil
	}
)

//nolint:funlen
func LoadChecks() {
	log.Debug("Loading %s suite checks", common.PlatformAlterationTestKey)

	checksGroup := checksdb.NewChecksGroup(common.PlatformAlterationTestKey).
		WithBeforeEachFn(beforeEachFn)

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("platform-alteration-hyperthread-enable")).
		WithSkipCheckFn(
			testhelper.GetNoBareMetalNodesSkipFn(&env),
			testhelper.GetDaemonSetFailedToSpawnSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckHyperthreadEnable).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("platform-alteration-base-image")).
		WithSkipCheckFn(
			testhelper.GetNonOCPClusterSkipFn(),
			testhelper.GetDaemonSetFailedToSpawnSkipFn(&env),
			testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckUnalteredBaseImage).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("platform-alteration-tainted-node-kernel")).
		WithSkipCheckFn(testhelper.GetDaemonSetFailedToSpawnSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckTainted).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("platform-alteration-isredhat-release")).
		WithSkipCheckFn(testhelper.GetNoContainersUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckIsRedHatRelease).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("platform-alteration-is-selinux-enforcing")).
		WithSkipCheckFn(
			testhelper.GetNonOCPClusterSkipFn(),
			testhelper.GetDaemonSetFailedToSpawnSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckSELinuxEnforcing).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("platform-alteration-hugepages-config")).
		WithSkipCheckFn(
			testhelper.GetNonOCPClusterSkipFn(),
			testhelper.GetDaemonSetFailedToSpawnSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckHugepages).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("platform-alteration-boot-params")).
		WithSkipCheckFn(
			testhelper.GetNonOCPClusterSkipFn(),
			testhelper.GetDaemonSetFailedToSpawnSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckBootParams).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("platform-alteration-sysctl-config")).
		WithSkipCheckFn(
			testhelper.GetNonOCPClusterSkipFn(),
			testhelper.GetDaemonSetFailedToSpawnSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckSysctl).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("platform-alteration-service-mesh-usage")).
		WithSkipCheckFn(
			testhelper.GetNoIstioSkipFn(&env),
			testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckServiceMeshUsage).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("platform-alteration-ocp-lifecycle")).
		WithSkipCheckFn(testhelper.GetNonOCPClusterSkipFn()).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckOCPLifecycle).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("platform-alteration-ocp-node-os-lifecycle")).
		WithSkipCheckFn(testhelper.GetNonOCPClusterSkipFn()).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckOCPNodeOSLifecycle).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("platform-alteration-hugepages-2m-only")).
		WithSkipCheckFn(
			testhelper.GetNonOCPClusterSkipFn(),
			testhelper.GetNoHugepagesPodsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckHugepages2MiOnly).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("platform-alteration-hugepages-1g-only")).
		WithSkipCheckFn(
			testhelper.GetNonOCPClusterSkipFn(),
			testhelper.GetNoHugepagesPodsSkipFn(&env)).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckHugepages1GiOnly).MakeCheckFn(&env)))

	checksGroup.Add(checksdb.NewCheck(checksadapter.GetCheckIDAndLabels("platform-alteration-cluster-operator-health")).
		WithSkipCheckFn(
			testhelper.GetNonOCPClusterSkipFn(),
		).
		WithCheckFn(checksadapter.NewAdapter(checksfn.CheckClusterOperatorHealth).MakeCheckFn(&env)))
}
