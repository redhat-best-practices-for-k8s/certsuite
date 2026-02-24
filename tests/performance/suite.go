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

package performance

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/crclient"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/scheduling"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/accesscontrol/resources"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/common"
	"github.com/redhat-best-practices-for-k8s/certsuite/tests/identifiers"
)

const (
	maxNumberOfExecProbes     = 10
	minExecProbePeriodSeconds = 10
)

var (
	env provider.TestEnvironment

	beforeEachFn = func(check *checksdb.Check) error {
		env = provider.GetTestEnvironment()
		return nil
	}

	skipIfNoGuaranteedPodContainersWithExclusiveCPUs = func() (bool, string) {
		var guaranteedPodContainersWithExclusiveCPUs = env.GetGuaranteedPodContainersWithExclusiveCPUs()
		if len(guaranteedPodContainersWithExclusiveCPUs) == 0 {
			return true, "There are no guaranteed pods with exclusive CPUs to check."
		}
		return false, ""
	}

	skipIfNoNonGuaranteedPodContainersWithoutHostPID = func() (bool, string) {
		var nonGuaranteedPodContainers = env.GetNonGuaranteedPodContainersWithoutHostPID()
		if len(nonGuaranteedPodContainers) == 0 {
			return true, "There are no non-guaranteed pods without HostPID to check."
		}
		return false, ""
	}

	skipIfNoGuaranteedPodContainersWithExclusiveCPUsWithoutHostPID = func() (bool, string) {
		var guaranteedPodContainersWithExclusiveCPUs = env.GetGuaranteedPodContainersWithExclusiveCPUsWithoutHostPID()
		if len(guaranteedPodContainersWithExclusiveCPUs) == 0 {
			return true, "There are no guaranteed pods without exclusive CPUs and without HostPID to check."
		}
		return false, ""
	}

	skipIfNoGuaranteedPodContainersWithIsolatedCPUsWithoutHostPID = func() (bool, string) {
		var guaranteedPodContainersWithIsolatedCPUs = env.GetGuaranteedPodContainersWithIsolatedCPUsWithoutHostPID()
		if len(guaranteedPodContainersWithIsolatedCPUs) == 0 {
			return true, "There are no guaranteed pods with isolated CPUs and without HostPID to check."
		}
		return false, ""
	}
)

func LoadChecks() {
	log.Debug("Loading %s suite checks", common.PerformanceTestKey)

	checksGroup := checksdb.NewChecksGroup(common.PerformanceTestKey).
		WithBeforeEachFn(beforeEachFn)

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestExclusiveCPUPoolIdentifier)).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testExclusiveCPUPool(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestRtAppNoExecProbes)).
		WithSkipCheckFn(skipIfNoGuaranteedPodContainersWithExclusiveCPUs).
		WithCheckFn(func(c *checksdb.Check) error {
			testRtAppsNoExecProbes(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestSharedCPUPoolSchedulingPolicy)).
		WithSkipCheckFn(skipIfNoNonGuaranteedPodContainersWithoutHostPID).
		WithCheckFn(func(c *checksdb.Check) error {
			testSchedulingPolicyInCPUPool(c, &env, env.GetNonGuaranteedPodContainersWithoutHostPID(), scheduling.SharedCPUScheduling)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestExclusiveCPUPoolSchedulingPolicy)).
		WithSkipCheckFn(skipIfNoGuaranteedPodContainersWithExclusiveCPUsWithoutHostPID).
		WithCheckFn(func(c *checksdb.Check) error {
			testSchedulingPolicyInCPUPool(c, &env, env.GetGuaranteedPodContainersWithExclusiveCPUsWithoutHostPID(), scheduling.ExclusiveCPUScheduling)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestIsolatedCPUPoolSchedulingPolicy)).
		WithSkipCheckFn(skipIfNoGuaranteedPodContainersWithIsolatedCPUsWithoutHostPID).
		WithCheckFn(func(c *checksdb.Check) error {
			testSchedulingPolicyInCPUPool(c, &env, env.GetGuaranteedPodContainersWithIsolatedCPUsWithoutHostPID(), scheduling.ExclusiveCPUScheduling)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestLimitedUseOfExecProbesIdentifier)).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testLimitedUseOfExecProbes(c, &env)
			return nil
		}))

	checksGroup.Add(checksdb.NewCheck(identifiers.GetTestIDAndLabels(identifiers.TestCPUPinningNoExecProbes)).
		WithSkipCheckFn(skipIfNoGuaranteedPodContainersWithExclusiveCPUs).
		WithCheckFn(func(c *checksdb.Check) error {
			cpuPinnedPods := env.GetGuaranteedPodsWithExclusiveCPUs()
			testCPUPinningNoExecProbes(c, cpuPinnedPods)
			return nil
		}))
}

//nolint:funlen
func testLimitedUseOfExecProbes(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	counter := 0
	for _, put := range env.Pods {
		for _, cut := range put.Containers {
			check.LogInfo("Testing Container %q", cut)
			if cut.LivenessProbe != nil && cut.LivenessProbe.Exec != nil {
				counter++
				if cut.LivenessProbe.PeriodSeconds >= minExecProbePeriodSeconds {
					check.LogInfo("Container %q has a LivenessProbe with PeriodSeconds greater than %d (%d seconds)",
						cut, minExecProbePeriodSeconds, cut.LivenessProbe.PeriodSeconds)

					compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(put.Namespace, put.Name,
						cut.Name, fmt.Sprintf("LivenessProbe exec probe has a PeriodSeconds greater than 10 (%d seconds)",
							cut.LivenessProbe.PeriodSeconds), true))
				} else {
					check.LogError("Container %q has a LivenessProbe with PeriodSeconds less than %d (%d seconds)",
						cut, minExecProbePeriodSeconds, cut.LivenessProbe.PeriodSeconds)

					nonCompliantObjects = append(nonCompliantObjects,
						testhelper.NewContainerReportObject(put.Namespace, put.Name,
							cut.Name, fmt.Sprintf("LivenessProbe exec probe has a PeriodSeconds that is not greater than 10 (%d seconds)",
								cut.LivenessProbe.PeriodSeconds), false))
				}
			}
			if cut.StartupProbe != nil && cut.StartupProbe.Exec != nil {
				counter++
				if cut.StartupProbe.PeriodSeconds >= minExecProbePeriodSeconds {
					check.LogInfo("Container %q has a StartupProbe with PeriodSeconds greater than %d (%d seconds)",
						cut, minExecProbePeriodSeconds, cut.LivenessProbe.PeriodSeconds)

					compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(put.Namespace, put.Name,
						cut.Name, fmt.Sprintf("StartupProbe exec probe has a PeriodSeconds greater than 10 (%d seconds)",
							cut.StartupProbe.PeriodSeconds), true))
				} else {
					check.LogError("Container %q has a StartupProbe with PeriodSeconds less than %d (%d seconds)",
						cut, minExecProbePeriodSeconds, cut.LivenessProbe.PeriodSeconds)

					nonCompliantObjects = append(nonCompliantObjects,
						testhelper.NewContainerReportObject(put.Namespace, put.Name,
							cut.Name, fmt.Sprintf("StartupProbe exec probe has a PeriodSeconds that is not greater than 10 (%d seconds)",
								cut.StartupProbe.PeriodSeconds), false))
				}
			}
			if cut.ReadinessProbe != nil && cut.ReadinessProbe.Exec != nil {
				counter++
				if cut.ReadinessProbe.PeriodSeconds >= minExecProbePeriodSeconds {
					check.LogInfo("Container %q has a ReadinessProbe with PeriodSeconds greater than %d (%d seconds)",
						cut, minExecProbePeriodSeconds, cut.LivenessProbe.PeriodSeconds)

					compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(put.Namespace, put.Name,
						cut.Name, fmt.Sprintf("ReadinessProbe exec probe has a PeriodSeconds greater than 10 (%d seconds)",
							cut.ReadinessProbe.PeriodSeconds), true))
				} else {
					check.LogError("Container %q has a ReadinessProbe with PeriodSeconds less than %d (%d seconds)",
						cut, minExecProbePeriodSeconds, cut.LivenessProbe.PeriodSeconds)

					nonCompliantObjects = append(nonCompliantObjects,
						testhelper.NewContainerReportObject(put.Namespace, put.Name,
							cut.Name, fmt.Sprintf("ReadinessProbe exec probe has a PeriodSeconds that is not greater than 10 (%d seconds)",
								cut.ReadinessProbe.PeriodSeconds), false))
				}
			}
		}
	}

	// If there >=10 exec probes, mark the entire cluster as a failure
	if counter >= maxNumberOfExecProbes {
		check.LogError("CNF has 10 or more exec probes (nb-exec-probes=%d)", counter)
		nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject(fmt.Sprintf("CNF has 10 or more exec probes (%d exec probes)", counter), testhelper.CnfType, false))
	} else {
		check.LogInfo("CNF has less than 10 exec probes (nb-exec-probes=%d)", counter)
		compliantObjects = append(compliantObjects, testhelper.NewReportObject(fmt.Sprintf("CNF has less than 10 exec probes (%d exec probes)", counter), testhelper.CnfType, true))
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testExclusiveCPUPool(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, put := range env.Pods {
		nBExclusiveCPUPoolContainers := 0
		nBSharedCPUPoolContainers := 0
		for _, cut := range put.Containers {
			if resources.HasExclusiveCPUsAssigned(cut, check.GetLogger()) {
				nBExclusiveCPUPoolContainers++
			} else {
				nBSharedCPUPoolContainers++
			}
		}

		if nBExclusiveCPUPoolContainers > 0 && nBSharedCPUPoolContainers > 0 {
			exclusiveStr := strconv.Itoa(nBExclusiveCPUPoolContainers)
			sharedStr := strconv.Itoa(nBSharedCPUPoolContainers)

			check.LogError("Pod %q has containers whose CPUs belong to different pools. Containers in the shared cpu pool: %d "+
				"Containers in the exclusive cpu pool: %d", put, nBSharedCPUPoolContainers, nBExclusiveCPUPoolContainers)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has containers whose CPUs belong to different pools", false).
				AddField("SharedCPUPoolContainers", sharedStr).
				AddField("ExclusiveCPUPoolContainers", exclusiveStr))
		} else {
			check.LogInfo("Pod %q has no containers whose CPUs belong to different pools", put)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(put.Namespace, put.Name, "Pod has no containers whose CPUs belong to different pools", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testSchedulingPolicyInCPUPool(check *checksdb.Check, env *provider.TestEnvironment,
	podContainers []*provider.Container, schedulingType string) {
	var compliantContainersPids []*testhelper.ReportObject
	var nonCompliantContainersPids []*testhelper.ReportObject
	for _, cut := range podContainers {
		check.LogInfo("Testing Container %q", cut)

		// Get the pid namespace
		pidNamespace, err := crclient.GetContainerPidNamespace(cut, env)
		if err != nil {
			check.LogError("Unable to get pid namespace for Container %q, err: %v", cut, err)
			nonCompliantContainersPids = append(nonCompliantContainersPids,
				testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, fmt.Sprintf("Internal error, err=%s", err), false))
			continue
		}
		check.LogDebug("PID namespace for Container %q is %q", cut, pidNamespace)

		// Get the list of process ids running in the pid namespace
		processes, err := crclient.GetPidsFromPidNamespace(pidNamespace, cut)
		if err != nil {
			check.LogError("Unable to get PIDs from PID namespace %q for Container %q, err: %v", pidNamespace, cut, err)
			nonCompliantContainersPids = append(nonCompliantContainersPids,
				testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, fmt.Sprintf("Internal error, err=%s", err), false))
			continue
		}

		compliantPids, nonCompliantPids := scheduling.ProcessPidsCPUScheduling(processes, cut, schedulingType, check.GetLogger())
		// Check for the specified priority for each processes running in that pid namespace

		compliantContainersPids = append(compliantContainersPids, compliantPids...)
		nonCompliantContainersPids = append(nonCompliantContainersPids, nonCompliantPids...)
	}

	check.SetResult(compliantContainersPids, nonCompliantContainersPids)
}

func getExecProbesCmds(c *provider.Container) map[string]bool {
	cmds := map[string]bool{}

	if c.LivenessProbe != nil && c.LivenessProbe.Exec != nil {
		cmd := strings.Join(c.LivenessProbe.Exec.Command, "")
		cmd = strings.Join(strings.Fields(cmd), "")
		cmds[cmd] = true
	}

	if c.ReadinessProbe != nil && c.ReadinessProbe.Exec != nil {
		cmd := strings.Join(c.ReadinessProbe.Exec.Command, "")
		cmd = strings.Join(strings.Fields(cmd), "")
		cmds[cmd] = true
	}

	if c.StartupProbe != nil && c.StartupProbe.Exec != nil {
		cmd := strings.Join(c.StartupProbe.Exec.Command, "")
		cmd = strings.Join(strings.Fields(cmd), "")
		cmds[cmd] = true
	}

	return cmds
}

func testRtAppsNoExecProbes(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	cuts := env.GetNonGuaranteedPodContainersWithoutHostPID()
	for _, cut := range cuts {
		check.LogInfo("Testing Container %q", cut)
		if !cut.HasExecProbes() {
			check.LogInfo("Container %q does not define exec probes", cut)
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container does not define exec probes", true))
			continue
		}

		processes, err := crclient.GetContainerProcesses(cut, env)
		if err != nil {
			check.LogError("Could not determine the processes pids for container %q, err: %v", cut, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Could not determine the processes pids for container", false))
			break
		}

		notExecProbeProcesses, compliantObjectsProbes := filterProbeProcesses(processes, cut)
		compliantObjects = append(compliantObjects, compliantObjectsProbes...)
		allProcessesCompliant := true
		for _, p := range notExecProbeProcesses {
			check.LogInfo("Testing process %q", p)
			schedPolicy, _, err := scheduling.GetProcessCPUScheduling(p.Pid, cut)
			if err != nil {
				// If the process does not exist anymore it means that it has finished since the time the process list
				// was retrieved. In this case, just ignore the error and continue processing the rest of the processes.
				if strings.Contains(err.Error(), scheduling.NoProcessFoundErrMsg) {
					check.LogWarn("Container process %q disappeared", p)
					compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container process disappeared", true).
						AddField(testhelper.ProcessID, strconv.Itoa(p.Pid)).
						AddField(testhelper.ProcessCommandLine, p.Args))
					continue
				}
				check.LogError("Could not determine the scheduling policy for container %q (pid=%d), err: %v", cut, p.Pid, err)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Could not determine the scheduling policy for container", false).
					AddField(testhelper.ProcessID, strconv.Itoa(p.Pid)).
					AddField(testhelper.ProcessCommandLine, p.Args))
				allProcessesCompliant = false
				continue
			}
			if scheduling.PolicyIsRT(schedPolicy) {
				check.LogError("Container %q defines exec probes while having a RT scheduling policy for process %q", cut, p)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container defines exec probes while having a RT scheduling policy", false).
					AddField(testhelper.ProcessID, strconv.Itoa(p.Pid)))
				allProcessesCompliant = false
			}
		}

		if allProcessesCompliant {
			check.LogInfo("Container %q defines exec probes but does not have a RT scheduling policy", cut)
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container defines exec probes but does not have a RT scheduling policy", true))
		}
	}

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func filterProbeProcesses(allProcesses []*crclient.Process, cut *provider.Container) (notExecProbeProcesses []*crclient.Process, compliantObjects []*testhelper.ReportObject) {
	execProbeProcesses := []int{}
	execProbesCmds := getExecProbesCmds(cut)
	// find all exec probes by matching command line
	for _, p := range allProcesses {
		if execProbesCmds[strings.Join(strings.Fields(p.Args), "")] {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container process belongs to an exec probe (skipping verification)", true).
				AddField(testhelper.ProcessID, strconv.Itoa(p.Pid)).
				AddField(testhelper.ProcessCommandLine, p.Args))
			execProbeProcesses = append(execProbeProcesses, p.Pid)
		}
	}
	// remove all exec probes and their children from the process list
	for _, p := range allProcesses {
		if slices.Contains(execProbeProcesses, p.Pid) || slices.Contains(execProbeProcesses, p.PPid) {
			// this process is part of an exec probe (child or parent), continue
			continue
		}
		notExecProbeProcesses = append(notExecProbeProcesses, p)
	}
	return notExecProbeProcesses, compliantObjects
}

func testCPUPinningNoExecProbes(check *checksdb.Check, cpuPinnedPods []*provider.Pod) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject

	for _, cpuPinnedPod := range cpuPinnedPods {
		execProbeFound := false
		for _, cut := range cpuPinnedPod.Containers {
			check.LogInfo("Testing Container %q", cut)
			if cut.HasExecProbes() {
				check.LogError("Container %q defines an exec probe", cut)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewPodReportObject(cpuPinnedPod.Namespace, cpuPinnedPod.Name, "Exec probe is not allowed on CPU-pinned pods", false))
				execProbeFound = true
			}
		}

		if !execProbeFound {
			check.LogInfo("Pod %q does not define any exec probe", cpuPinnedPod)
			compliantObjects = append(compliantObjects, testhelper.NewPodReportObject(cpuPinnedPod.Namespace, cpuPinnedPod.Name, "No exec probes found", true))
		}
	}
	check.SetResult(compliantObjects, nonCompliantObjects)
}
