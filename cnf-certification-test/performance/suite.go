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
	"strconv"

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/resources"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/results"
	"github.com/test-network-function/cnf-certification-test/internal/crclient"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/scheduling"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

// The pods with no access to host network are considered for these tests
var _ = ginkgo.Describe(common.PerformanceTestKey, func() {
	logrus.Debugf("Entering %s suite", common.PerformanceTestKey)
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		env = provider.GetTestEnvironment()
	})
	ginkgo.ReportAfterEach(results.RecordResult)

	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestExclusiveCPUPoolIdentifier)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		testhelper.SkipIfEmptyAny(ginkgo.Skip, env.Pods)
		testExclusiveCPUPool(&env)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestRtAppNoExecProbes)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		var guaranteedPodContainersWithExclusiveCPUs = env.GetGuaranteedPodContainersWithExclusiveCPUs()
		testhelper.SkipIfEmptyAll(ginkgo.Skip, guaranteedPodContainersWithExclusiveCPUs)
		testRtAppsNoExecProbes(&env, guaranteedPodContainersWithExclusiveCPUs)
	})

	// Scheduling related tests begins here
	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSharedCPUPoolSchedulingPolicy)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		var nonGuaranteedPodContainers = env.GetNonGuaranteedPodContainersWithoutHostPID()
		testhelper.SkipIfEmptyAll(ginkgo.Skip, nonGuaranteedPodContainers)
		testSchedulingPolicyInCPUPool(&env, nonGuaranteedPodContainers, scheduling.SharedCPUScheduling)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestExclusiveCPUPoolSchedulingPolicy)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		var guaranteedPodContainersWithExclusiveCPUs = env.GetGuaranteedPodContainersWithExclusiveCPUsWithoutHostPID()
		testhelper.SkipIfEmptyAll(ginkgo.Skip, guaranteedPodContainersWithExclusiveCPUs)
		testSchedulingPolicyInCPUPool(&env, guaranteedPodContainersWithExclusiveCPUs, scheduling.ExclusiveCPUScheduling)
	})

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestIsolatedCPUPoolSchedulingPolicy)
	ginkgo.It(testID, ginkgo.Label(tags...), func() {
		var guaranteedPodContainersWithIsolatedCPUs = env.GetGuaranteedPodContainersWithIsolatedCPUsWithoutHostPID()
		testhelper.SkipIfEmptyAll(ginkgo.Skip, guaranteedPodContainersWithIsolatedCPUs)
		testSchedulingPolicyInCPUPool(&env, guaranteedPodContainersWithIsolatedCPUs, scheduling.IsolatedCPUScheduling)
	})
	// Scheduling related tests ends here
})

func testExclusiveCPUPool(env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, put := range env.Pods {
		nBExclusiveCPUPoolContainers := 0
		nBSharedCPUPoolContainers := 0
		for _, cut := range put.Containers {
			if resources.HasExclusiveCPUsAssigned(cut) {
				nBExclusiveCPUPoolContainers++
			} else {
				nBSharedCPUPoolContainers++
			}
		}

		if nBExclusiveCPUPoolContainers > 0 && nBSharedCPUPoolContainers > 0 {
			exclusiveStr := strconv.Itoa(nBExclusiveCPUPoolContainers)
			sharedStr := strconv.Itoa(nBSharedCPUPoolContainers)

			tnf.ClaimFilePrintf("Pod: %s has containers whose CPUs belong to different pools. Containers in the shared cpu pool: %d "+
				"Containers in the exclusive cpu pool: %d", put.String(), nBSharedCPUPoolContainers, nBExclusiveCPUPoolContainers)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has containers whose CPUs belong to different pools", false).
				AddField("SharedCPUPoolContainers", sharedStr).
				AddField("ExclusiveCPUPoolContainers", exclusiveStr))
		} else {
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has no containers whose CPUs belong to different pools", true))
		}
	}

	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testSchedulingPolicyInCPUPool(env *provider.TestEnvironment,
	podContainers []*provider.Container, schedulingType string) {
	var compliantContainersPids []*testhelper.ReportObject
	var nonCompliantContainersPids []*testhelper.ReportObject
	for _, testContainer := range podContainers {
		logrus.Infof("Processing %v", testContainer)

		// Get the pid namespace
		pidNamespace, err := crclient.GetContainerPidNamespace(testContainer, env)
		if err != nil {
			logrus.Errorf("unable to get pid namespace due to: %v", err)
			tnf.ClaimFilePrintf("Failed", "Incomplete processing for %v while getting pid namespace err %v", testContainer, err)
			nonCompliantContainersPids = append(nonCompliantContainersPids,
				testhelper.NewContainerReportObject(testContainer.Namespace, testContainer.Podname, testContainer.Name, fmt.Sprintf("Internal error, err=%s", err), false))
			continue
		}
		logrus.Debugf("Obtained pidNamespace for %s is %s", testContainer, pidNamespace)

		// Get the list of process ids running in the pid namespace
		processes, err := crclient.GetPidsFromPidNamespace(pidNamespace, testContainer)

		if err != nil {
			nonCompliantContainersPids = append(nonCompliantContainersPids,
				testhelper.NewContainerReportObject(testContainer.Namespace, testContainer.Podname, testContainer.Name, fmt.Sprintf("Internal error, err=%s", err), false))
		}

		compliantPids, nonCompliantPids := scheduling.ProcessPidsCPUScheduling(processes, testContainer, schedulingType)
		// Check for the specified priority for each processes running in that pid namespace

		compliantContainersPids = append(compliantContainersPids, compliantPids...)
		nonCompliantContainersPids = append(nonCompliantContainersPids, nonCompliantPids...)

		logrus.Debugf("Processed %v", testContainer)
	}

	testhelper.AddTestResultReason(compliantContainersPids, nonCompliantContainersPids, tnf.ClaimFilePrintf, ginkgo.Fail)
}

func testRtAppsNoExecProbes(env *provider.TestEnvironment, cuts []*provider.Container) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	for _, cut := range cuts {
		processes, err := crclient.GetContainerProcesses(cut, env)
		if err != nil {
			tnf.ClaimFilePrintf("Could not determine the processes pids for container %s, err: %v", cut, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Could not determine the processes pids for container", false))
			break
		}
		for _, p := range processes {
			schedPolicy, _, err := scheduling.GetProcessCPUScheduling(p.Pid, cut)
			if err != nil {
				tnf.ClaimFilePrintf("Could not determine the scheduling policy for container %s (pid=%v), err: %v", cut, p.Pid, err)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Could not determine the scheduling policy for container", false).
					AddField(testhelper.ProcessID, strconv.Itoa(p.Pid)))
				break
			}
			if scheduling.PolicyIsRT(schedPolicy) && cut.HasExecProbes() {
				tnf.ClaimFilePrintf("Pod %s/Container %s defines exec probes while having a RT scheduling policy for pid %d", cut.Podname, cut, p.Pid)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container defines exec probes while having a RT scheduling policy", false).
					AddField(testhelper.ProcessID, strconv.Itoa(p.Pid)))
			}
		}
	}
	testhelper.AddTestResultReason(compliantObjects, nonCompliantObjects, tnf.ClaimFilePrintf, ginkgo.Fail)
}
