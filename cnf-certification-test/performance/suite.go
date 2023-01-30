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

package performance

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/internal/crclient"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/scheduling"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

// All actual test code belongs below here.  Utilities belong above.
var _ = ginkgo.Describe(common.PerformanceTestKey, func() {
	logrus.Debugf("Entering %s suite", common.PerformanceTestKey)
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})
	ginkgo.ReportAfterEach(results.RecordResult)

	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSharedCPUPoolSchedulingPolicy)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		var nonGuaranteedPodContainers = env.GetNonGuaranteedPodContainers()
		testhelper.SkipIfEmptyAll(ginkgo.Skip, nonGuaranteedPodContainers)
		testSchedulingPolicyInCPUPool(&env, nonGuaranteedPodContainers, scheduling.SharedCPUScheduling)
	})
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestExclusiveCPUPoolSchedulingPolicy)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		var guaranteedPodContainersWithExclusiveCPUs = env.GetGuaranteedPodContainersWithExlusiveCPUs()
		testhelper.SkipIfEmptyAll(ginkgo.Skip, guaranteedPodContainersWithExclusiveCPUs)
		testSchedulingPolicyInCPUPool(&env, guaranteedPodContainersWithExclusiveCPUs, scheduling.ExclusiveCPUScheduling)
	})
})

func testSchedulingPolicyInCPUPool(env *provider.TestEnvironment,
	podContainers []*provider.Container, schedulingType string) {
	nonCompliantContainers := make(map[*provider.Container][]int)
	var compliantContainers []*provider.Container
	for _, testContainer := range podContainers {
		logrus.Infof("Processing %v", testContainer)

		// Get the pid namespace
		pidNamespace, err := crclient.GetContainerPidNamespace(testContainer, env)
		if err != nil {
			logrus.Errorf("unable to get pid namespace due to: %v", err)
			tnf.ClaimFilePrintf("Failed", "Incomplete processing for %v while getting pid namespace err %v", testContainer, err)
			continue
		}
		logrus.Debugf("Obtained pidNamespace for %s is %s", testContainer, pidNamespace)

		// Get the list of process ids running in the pid namespace
		pids := crclient.GetPidsFromPidNamespace(pidNamespace, testContainer)

		// Check for the specified priority for each processes running in that pid namespace
		if scheduling.ProcessPidsCPUScheduling(pids, testContainer, nonCompliantContainers, schedulingType) {
			compliantContainers = append(compliantContainers, testContainer)
		}
		logrus.Debugf("Processed %v", testContainer)
	}
	if len(nonCompliantContainers) != 0 {
		testhelper.AddTestResultLog("Non-compliant", nonCompliantContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
	}
	tnf.ClaimFilePrintf("Compliant", compliantContainers)
}
