// Copyright (C) 2023-2024 Red Hat, Inc.
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

package scheduling

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/crclient"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
)

const (
	CurrentSchedulingPolicy   = "current scheduling policy"
	CurrentSchedulingPriority = "current scheduling priority"
	newLineCharacter          = "\n"

	SharedCPUScheduling    = "SHARED_CPU_SCHEDULING"
	ExclusiveCPUScheduling = "EXCLUSIVE_CPU_SCHEDULING"
	IsolatedCPUScheduling  = "ISOLATED_CPU_SCHEDULING"

	SchedulingRoundRobin      = "SCHED_RR"
	SchedulingFirstInFirstOut = "SCHED_FIFO"

	InvalidPriority = -1
)

var (
	CrcClientExecCommandContainerNSEnter = crclient.ExecCommandContainerNSEnter
	GetProcessCPUSchedulingFn            = GetProcessCPUScheduling
)

// parseSchedulingPolicyAndPriority parses a scheduling policy string into its components.
//
// It accepts a string in the format "policy:priority" where priority is an integer.
// The function splits the input, validates the presence of both parts,
// converts the priority to an int, and returns the policy string, priority value, and any error encountered.
func parseSchedulingPolicyAndPriority(chrtCommandOutput string) (schedPolicy string, schedPriority int, err error) {
	/*	Sample output:
		pid 476's current scheduling policy: SCHED_OTHER
		pid 476's current scheduling priority: 0*/

	lines := strings.Split(chrtCommandOutput, newLineCharacter)

	for _, line := range lines {
		if line == "" {
			continue
		}

		tokens := strings.Fields(line)
		lastToken := tokens[len(tokens)-1]

		switch {
		case strings.Contains(line, CurrentSchedulingPolicy):
			schedPolicy = lastToken
		case strings.Contains(line, CurrentSchedulingPriority):
			schedPriority, err = strconv.Atoi(lastToken)
			if err != nil {
				log.Error("Error obtained during strconv %v", err)
				return schedPolicy, InvalidPriority, err
			}
		default:
			return schedPolicy, InvalidPriority, fmt.Errorf("invalid: %s", line)
		}
	}
	return schedPolicy, schedPriority, nil
}

var schedulingRequirements = map[string]string{SharedCPUScheduling: "SHARED_CPU_SCHEDULING: scheduling priority == 0",
	ExclusiveCPUScheduling: "EXCLUSIVE_CPU_SCHEDULING: scheduling priority < 10 and scheduling policy == SCHED_RR or SCHED_FIFO",
	IsolatedCPUScheduling:  "ISOLATED_CPU_SCHEDULING: scheduling policy == SCHED_RR or SCHED_FIFO"}

// ProcessPidsCPUScheduling analyzes the CPU scheduling settings of processes inside a container and returns report objects.
//
// It takes a slice of process pointers, a container reference, a string identifier, and a logger.
// For each process it retrieves the current CPU scheduling policy and priority using GetProcessCPUSchedulingFn,
// logs debug information, and records any errors encountered. The function then populates the container’s
// process values with the retrieved scheduling data, creates report objects for each process, and returns
// a slice of those reports. If an error occurs while retrieving or setting values, it is logged and
// the process is skipped in the final report list.
func ProcessPidsCPUScheduling(processes []*crclient.Process, testContainer *provider.Container, check string, logger *log.Logger) (compliantContainerPids, nonCompliantContainerPids []*testhelper.ReportObject) {
	hasCPUSchedulingConditionSuccess := false
	for _, process := range processes {
		logger.Debug("Testing process %q", process)
		schedulePolicy, schedulePriority, err := GetProcessCPUSchedulingFn(process.Pid, testContainer)
		if err != nil {
			logger.Error("Unable to get the scheduling policy and priority : %v", err)
			return compliantContainerPids, nonCompliantContainerPids
		}

		switch check {
		case SharedCPUScheduling:
			hasCPUSchedulingConditionSuccess = schedulePriority == 0
		case ExclusiveCPUScheduling:
			hasCPUSchedulingConditionSuccess = schedulePriority == 0 || (schedulePriority < 10 && (schedulePolicy == SchedulingRoundRobin || schedulePolicy == SchedulingFirstInFirstOut))
		case IsolatedCPUScheduling:
			hasCPUSchedulingConditionSuccess = schedulePriority >= 10 && (schedulePolicy == SchedulingRoundRobin || schedulePolicy == SchedulingFirstInFirstOut)
		}

		if !hasCPUSchedulingConditionSuccess {
			logger.Error("Process %q in Container %q with cpu scheduling policy=%s, priority=%d did not satisfy cpu scheduling requirements", process, testContainer, schedulePolicy, schedulePriority)
			aPidOut := testhelper.NewContainerReportObject(testContainer.Namespace, testContainer.Podname, testContainer.Name, "process does not satisfy: "+schedulingRequirements[check], false).
				SetContainerProcessValues(schedulePolicy, fmt.Sprint(schedulePriority), process.Args)
			nonCompliantContainerPids = append(nonCompliantContainerPids, aPidOut)
			continue
		}
		logger.Info("Process %q in Container %q with cpu scheduling policy=%s, priority=%d satisfies cpu scheduling requirements", process, testContainer, schedulePolicy, schedulePriority)
		aPidOut := testhelper.NewContainerReportObject(testContainer.Namespace, testContainer.Podname, testContainer.Name, "process satisfies: "+schedulingRequirements[check], true).
			SetContainerProcessValues(schedulePolicy, fmt.Sprint(schedulePriority), process.Args)
		compliantContainerPids = append(compliantContainerPids, aPidOut)
	}
	return compliantContainerPids, nonCompliantContainerPids
}

// GetProcessCPUScheduling retrieves the CPU scheduling policy and priority
// of a process running inside a container.
//
// It takes a PID and a reference to the container in which the process is
// running, executes a command inside that container to query the scheduler,
// parses the output into a scheduling policy string and an integer priority,
// and returns them along with any error encountered. If the process cannot be
// found or the scheduling information cannot be parsed, it returns an empty
// policy, zero priority, and an error describing the failure.
func GetProcessCPUScheduling(pid int, testContainer *provider.Container) (schedulePolicy string, schedulePriority int, err error) {
	log.Info("Checking the scheduling policy/priority in %v for pid=%d", testContainer, pid)

	command := fmt.Sprintf("chrt -p %d", pid)
	env := provider.GetTestEnvironment()
	ctx, err := crclient.GetNodeProbePodContext(testContainer.NodeName, &env)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get probe pod's context for container %s: %v", testContainer, err)
	}

	ch := clientsholder.GetClientsHolder()

	stdout, stderr, err := ch.ExecCommandContainer(ctx, command)
	if err != nil || stderr != "" {
		return schedulePolicy, InvalidPriority, fmt.Errorf("command %q failed to run in probe pod %s (node %s): %v (stderr: %v)",
			command, ctx.GetPodName(), testContainer.NodeName, err, stderr)
	}

	schedulePolicy, schedulePriority, err = parseSchedulingPolicyAndPriority(stdout)
	if err != nil {
		return schedulePolicy, InvalidPriority, fmt.Errorf("error getting the scheduling policy and priority for %v : %v", testContainer, err)
	}
	log.Info("pid %d in %v has the cpu scheduling policy %s, scheduling priority %d", pid, testContainer, schedulePolicy, schedulePriority)

	return schedulePolicy, schedulePriority, err
}

// PolicyIsRT reports whether the given scheduling policy name corresponds to a real‑time policy.
//
// It accepts a single string argument representing a scheduling policy and
// returns true if that policy is one of the recognized real‑time policies
// (e.g., "SCHED_FIFO" or "SCHED_RR"), otherwise it returns false.
func PolicyIsRT(schedPolicy string) bool {
	return schedPolicy == SchedulingFirstInFirstOut || schedPolicy == SchedulingRoundRobin
}
