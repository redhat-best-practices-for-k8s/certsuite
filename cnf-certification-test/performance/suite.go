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
	"fmt"

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

const (
	SharedCPUScheduling    = "SHARED_CPU_SCHEDULING"
	ExclusiveCPUScheduling = "EXCLUSIVE_CPU_SCHEDULING"
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
		testSchedOtherPolicyInSharedCPUPool(&env, nonGuaranteedPodContainers)
	})
})

func testSchedOtherPolicyInSharedCPUPool(env *provider.TestEnvironment,
	nonGuaranteedPodContainers []*provider.Container) {
	nonCompliantContainers := make(map[*provider.Container][]int)

	for _, testContainer := range nonGuaranteedPodContainers {
		logrus.Infof("Processing %v", testContainer)

		// Get the pid namespace
		pidNamespace, err := crclient.GetContainerPidNamespace(testContainer, env)
		if err != nil {
			logrus.Errorf("unable to get pid namespace due to: %v", err)
		}
		logrus.Debugf("Obtained pidNamespace for %s is %s", testContainer, pidNamespace)

		// Get the list of process ids running in the pid namespace
		pids := crclient.GetPidsFromPidNamespace(pidNamespace, testContainer)

		// Check for the specified priority for each processes running in that pid namespace
		processPidsCPUScheduling(pids, testContainer, nonCompliantContainers, SharedCPUScheduling)
		logrus.Infof("Processed %v", testContainer)
	}
	if len(nonCompliantContainers) != 0 {
		testhelper.AddTestResultLog("Non-compliant", nonCompliantContainers, tnf.ClaimFilePrintf, ginkgo.Fail)
	}
}

func processPidsCPUScheduling(pids []int, testContainer *provider.Container, nonCompliantContainers map[*provider.Container][]int, check string) {
	for _, pid := range pids {
		_, schedulePriority, err := getProcessCPUScheduling(pid, testContainer)
		if err != nil {
			logrus.Errorf("error getting the scheduling policy and priority : %v", err)
		}

		var result bool
		switch check {
		case SharedCPUScheduling:
			result = schedulePriority == 0
		case ExclusiveCPUScheduling:
			result = schedulePriority < 10
		}

		if !result {
			nonCompliantProcessIds, ok := nonCompliantContainers[testContainer]
			if !ok {
				nonCompliantContainers[testContainer] = []int{}
			} else {
				nonCompliantProcessIds = append(nonCompliantProcessIds, pid)
				nonCompliantContainers[testContainer] = nonCompliantProcessIds
			}
		}
		logrus.Debugf("Non-compliant containers for pid=%d are : %v", pid, nonCompliantContainers)
	}
}

func getProcessCPUScheduling(pid int, testContainer *provider.Container) (schedulePolicy string, schedulePriority int, err error) {
	logrus.Infof("Checking the scheduling policy/priority in %v for pid=%d", testContainer, pid)

	command := fmt.Sprintf("chrt -p %d", pid)

	stdout, stderr, err := crclient.ExecCommandContainerNSEnter(command, testContainer)
	if err != nil || stderr != "" {
		return schedulePolicy, schedulePriority, fmt.Errorf("unable to run nsenter for %v due to : %v", testContainer, err)
	}

	schedulePolicy, schedulePriority, err = scheduling.GetSchedulingPolicyAndPriority(stdout)
	if err != nil {
		return schedulePolicy, schedulePriority, fmt.Errorf("error getting the scheduling policy and priority for %v : %v", testContainer, err)
	}
	logrus.Infof("pid %d in %v has the cpu scheduling policy %s, scheduling priority %d", pid, testContainer, schedulePolicy, schedulePriority)

	return schedulePolicy, schedulePriority, err
}
