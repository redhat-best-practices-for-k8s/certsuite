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
	"slices"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/accesscontrol/resources"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/internal/crclient"
	"github.com/test-network-function/cnf-certification-test/pkg/checksdb"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/scheduling"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
	v1 "k8s.io/api/core/v1"
)

const (
	maxNumberOfExecProbes     = 10
	minExecProbePeriodSeconds = 10
)

var (
	env provider.TestEnvironment

	beforeEachFn = func(check *checksdb.Check) error {
		logrus.Infof("Check %s: getting test environment.", check.ID)
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

//nolint:funlen
func LoadChecks() {
	logrus.Debugf("Entering %s suite", common.PerformanceTestKey)

	checksGroup := checksdb.NewChecksGroup(common.PerformanceTestKey).
		WithBeforeEachFn(beforeEachFn)

	testID, tags := identifiers.GetGinkgoTestIDAndLabels(identifiers.TestExclusiveCPUPoolIdentifier)
	check := checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testExclusiveCPUPool(c, &env)
			return nil
		})

	checksGroup.Add(check)

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestRtAppNoExecProbes)
	check = checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(skipIfNoGuaranteedPodContainersWithExclusiveCPUs).
		WithCheckFn(func(c *checksdb.Check) error {
			testRtAppsNoExecProbes(c, &env)
			return nil
		})

	checksGroup.Add(check)

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestSharedCPUPoolSchedulingPolicy)
	check = checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(skipIfNoNonGuaranteedPodContainersWithoutHostPID).
		WithCheckFn(func(c *checksdb.Check) error {
			testSchedulingPolicyInCPUPool(c, &env, env.GetNonGuaranteedPodContainersWithoutHostPID(), scheduling.SharedCPUScheduling)
			return nil
		})

	checksGroup.Add(check)

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestExclusiveCPUPoolSchedulingPolicy)
	check = checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(skipIfNoGuaranteedPodContainersWithExclusiveCPUsWithoutHostPID).
		WithCheckFn(func(c *checksdb.Check) error {
			testSchedulingPolicyInCPUPool(c, &env, env.GetGuaranteedPodContainersWithExclusiveCPUsWithoutHostPID(), scheduling.ExclusiveCPUScheduling)
			return nil
		})

	checksGroup.Add(check)

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestIsolatedCPUPoolSchedulingPolicy)
	check = checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(skipIfNoGuaranteedPodContainersWithIsolatedCPUsWithoutHostPID).
		WithCheckFn(func(c *checksdb.Check) error {
			testSchedulingPolicyInCPUPool(c, &env, env.GetGuaranteedPodContainersWithIsolatedCPUsWithoutHostPID(), scheduling.ExclusiveCPUScheduling)
			return nil
		})

	checksGroup.Add(check)

	testID, tags = identifiers.GetGinkgoTestIDAndLabels(identifiers.TestLimitedUseOfExecProbesIdentifier)
	check = checksdb.NewCheck(testID, tags).
		WithSkipCheckFn(testhelper.GetNoPodsUnderTestSkipFn(&env)).
		WithCheckFn(func(c *checksdb.Check) error {
			testLimitedUseOfExecProbes(c, &env)
			return nil
		})

	checksGroup.Add(check)
}

func CheckProbePeriodSeconds(elem *v1.Probe, cut *provider.Container, s string) bool {
	if elem.PeriodSeconds > minExecProbePeriodSeconds {
		tnf.ClaimFilePrintf("Container %s is using exec probes, PeriodSeconds of %s: %s", cut, s,
			elem.PeriodSeconds)
		return true
	}
	tnf.ClaimFilePrintf("Container %s is not using of exec probes, PeriodSeconds of %s: %s", cut, s,
		elem.PeriodSeconds)
	return false
}

func testLimitedUseOfExecProbes(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	counter := 0
	for _, put := range env.Pods {
		for _, cut := range put.Containers {
			if cut.LivenessProbe != nil && cut.LivenessProbe.Exec != nil {
				counter++
				if CheckProbePeriodSeconds(cut.LivenessProbe, cut, "LivenessProbe") {
					compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(put.Namespace, put.Name,
						cut.Name, fmt.Sprintf("LivenessProbe exec probe has a PeriodSeconds greater than 10 (%d seconds)",
							cut.LivenessProbe.PeriodSeconds), true))
				} else {
					nonCompliantObjects = append(nonCompliantObjects,
						testhelper.NewContainerReportObject(put.Namespace, put.Name,
							cut.Name, fmt.Sprintf("LivenessProbe exec probe has a PeriodSeconds that is not greater than 10 (%d seconds)",
								cut.LivenessProbe.PeriodSeconds), false))
				}
			}
			if cut.StartupProbe != nil && cut.StartupProbe.Exec != nil {
				counter++
				if CheckProbePeriodSeconds(cut.StartupProbe, cut, "StartupProbe") {
					compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(put.Namespace, put.Name,
						cut.Name, fmt.Sprintf("StartupProbe exec probe has a PeriodSeconds greater than 10 (%d seconds)",
							cut.StartupProbe.PeriodSeconds), true))
				} else {
					nonCompliantObjects = append(nonCompliantObjects,
						testhelper.NewContainerReportObject(put.Namespace, put.Name,
							cut.Name, fmt.Sprintf("StartupProbe exec probe has a PeriodSeconds that is not greater than 10 (%d seconds)",
								cut.StartupProbe.PeriodSeconds), false))
				}
			}
			if cut.ReadinessProbe != nil && cut.ReadinessProbe.Exec != nil {
				counter++
				if CheckProbePeriodSeconds(cut.ReadinessProbe, cut, "ReadinessProbe") {
					compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(put.Namespace, put.Name,
						cut.Name, fmt.Sprintf("ReadinessProbe exec probe has a PeriodSeconds greater than 10 (%d seconds)",
							cut.ReadinessProbe.PeriodSeconds), true))
				} else {
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
		tnf.ClaimFilePrintf(fmt.Sprintf("CNF has %d exec probes", counter))
		nonCompliantObjects = append(nonCompliantObjects, testhelper.NewReportObject(fmt.Sprintf("CNF has 10 or more exec probes (%d exec probes)", counter), testhelper.CnfType, false))
	} else {
		// Compliant object
		compliantObjects = append(compliantObjects, testhelper.NewReportObject(fmt.Sprintf("CNF has less than 10 exec probes (%d exec probes)", counter), testhelper.CnfType, true))
		tnf.ClaimFilePrintf(fmt.Sprintf("CNF has less than %d exec probes", counter))
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

	check.SetResult(compliantObjects, nonCompliantObjects)
}

func testSchedulingPolicyInCPUPool(check *checksdb.Check, env *provider.TestEnvironment,
	podContainers []*provider.Container, schedulingType string) {
	var compliantContainersPids []*testhelper.ReportObject
	var nonCompliantContainersPids []*testhelper.ReportObject
	for _, testContainer := range podContainers {
		logrus.Infof("Processing %v", testContainer)

		// Get the pid namespace
		pidNamespace, err := crclient.GetContainerPidNamespace(testContainer, env)
		if err != nil {
			tnf.Logf(logrus.ErrorLevel, "unable to get pid namespace for container %s, err: %v", testContainer, err)
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

const noProcessFoundErrMsg = "No such process"

func testRtAppsNoExecProbes(check *checksdb.Check, env *provider.TestEnvironment) {
	var compliantObjects []*testhelper.ReportObject
	var nonCompliantObjects []*testhelper.ReportObject
	cuts := env.GetNonGuaranteedPodContainersWithoutHostPID()
	for _, cut := range cuts {
		if !cut.HasExecProbes() {
			compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container does not define exec probes", true))
			continue
		}

		processes, err := crclient.GetContainerProcesses(cut, env)
		if err != nil {
			tnf.ClaimFilePrintf("Could not determine the processes pids for container %s, err: %v", cut, err)
			nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Could not determine the processes pids for container", false))
			break
		}

		notExecProbeProcesses, compliantObjectsProbes := filterProbeProcesses(processes, cut)
		compliantObjects = append(compliantObjects, compliantObjectsProbes...)
		allProcessesCompliant := true
		for _, p := range notExecProbeProcesses {
			schedPolicy, _, err := scheduling.GetProcessCPUScheduling(p.Pid, cut)
			if err != nil {
				// If the process does not exist anymore it means that it has finished since the time the process list
				// was retrieved. In this case, just ignore the error and continue processing the rest of the processes.
				if strings.Contains(err.Error(), noProcessFoundErrMsg) {
					compliantObjects = append(compliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container process disappeared", true).
						AddField(testhelper.ProcessID, strconv.Itoa(p.Pid)).
						AddField(testhelper.ProcessCommandLine, p.Args))
					continue
				}
				tnf.ClaimFilePrintf("Could not determine the scheduling policy for container %s (pid=%v), err: %v", cut, p.Pid, err)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Could not determine the scheduling policy for container", false).
					AddField(testhelper.ProcessID, strconv.Itoa(p.Pid)).
					AddField(testhelper.ProcessCommandLine, p.Args))
				allProcessesCompliant = false
				continue
			}
			if scheduling.PolicyIsRT(schedPolicy) {
				tnf.ClaimFilePrintf("Pod %s/Container %s defines exec probes while having a RT scheduling policy for pid %d", cut.Podname, cut, p.Pid)
				nonCompliantObjects = append(nonCompliantObjects, testhelper.NewContainerReportObject(cut.Namespace, cut.Podname, cut.Name, "Container defines exec probes while having a RT scheduling policy", false).
					AddField(testhelper.ProcessID, strconv.Itoa(p.Pid)))
				allProcessesCompliant = false
			}
		}

		if allProcessesCompliant {
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
