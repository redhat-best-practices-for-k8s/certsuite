package scheduling

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/crclient"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
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
				logrus.Errorf("Error obtained during strconv %v", err)
				return schedPolicy, InvalidPriority, err
			}
		default:
			return schedPolicy, InvalidPriority, fmt.Errorf("invalid: %s", line)
		}
	}
	return schedPolicy, schedPriority, nil
}

func ProcessPidsCPUScheduling(pids []int, testContainer *provider.Container, nonCompliantContainers map[*provider.Container][]int, check string) (hasCPUSchedulingConditionSuccess bool) {
	for _, pid := range pids {
		schedulePolicy, schedulePriority, err := GetProcessCPUSchedulingFn(pid, testContainer)
		if err != nil {
			logrus.Errorf("error getting the scheduling policy and priority : %v", err)
			return hasCPUSchedulingConditionSuccess
		}

		switch check {
		case SharedCPUScheduling:
			hasCPUSchedulingConditionSuccess = schedulePriority == 0
		case ExclusiveCPUScheduling:
			hasCPUSchedulingConditionSuccess = schedulePriority < 10 && (schedulePolicy == SchedulingRoundRobin || schedulePolicy == SchedulingFirstInFirstOut)
		case IsolatedCPUScheduling:
			hasCPUSchedulingConditionSuccess = schedulePolicy == SchedulingRoundRobin || schedulePolicy == SchedulingFirstInFirstOut
		}

		if !hasCPUSchedulingConditionSuccess {
			tnf.ClaimFilePrintf("pid %d in %v with cpu scheduling policy %s, priority %d did not satisfy cpu scheduling requirements", pid, testContainer, schedulePolicy, schedulePriority)
			nonCompliantProcessIds, ok := nonCompliantContainers[testContainer]
			if !ok {
				nonCompliantContainers[testContainer] = []int{pid}
			} else {
				nonCompliantProcessIds = append(nonCompliantProcessIds, pid)
				nonCompliantContainers[testContainer] = nonCompliantProcessIds
			}
		}
	}
	return hasCPUSchedulingConditionSuccess
}

func GetProcessCPUScheduling(pid int, testContainer *provider.Container) (schedulePolicy string, schedulePriority int, err error) {
	logrus.Infof("Checking the scheduling policy/priority in %v for pid=%d", testContainer, pid)

	command := fmt.Sprintf("chrt -p %d", pid)

	stdout, stderr, err := CrcClientExecCommandContainerNSEnter(command, testContainer)
	if err != nil || stderr != "" {
		return schedulePolicy, InvalidPriority, fmt.Errorf("unable to run nsenter for %v due to : %v", testContainer, err)
	}

	schedulePolicy, schedulePriority, err = parseSchedulingPolicyAndPriority(stdout)
	if err != nil {
		return schedulePolicy, InvalidPriority, fmt.Errorf("error getting the scheduling policy and priority for %v : %v", testContainer, err)
	}
	logrus.Infof("pid %d in %v has the cpu scheduling policy %s, scheduling priority %d", pid, testContainer, schedulePolicy, schedulePriority)

	return schedulePolicy, schedulePriority, err
}
