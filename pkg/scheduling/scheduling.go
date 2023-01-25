package scheduling

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	SucessChrtCommandOutputLines = 2
)

func GetSchedulingPolicyAndPriority(chrtCommandOutput string) (schedPolicy string, schedPriority int, err error) {
	/*	Sample output:
		pid 476's current scheduling policy: SCHED_OTHER
		pid 476's current scheduling priority: 0*/

	lines := strings.Split(chrtCommandOutput, "\n")

	if len(lines) != SucessChrtCommandOutputLines {
		return schedPolicy, schedPriority, fmt.Errorf("error in parsing %s", chrtCommandOutput)
	}

	policyStr := lines[0]
	prioritySubstr := lines[1]
	// Get policy value
	policyTokens := strings.Fields(policyStr)
	schedPolicy = policyTokens[len(policyTokens)-1]
	logrus.Infof("Obtained scheduling policy = %s", schedPolicy)

	// Get priority value
	priorityTokens := strings.Fields(prioritySubstr)
	schedPriority, err = strconv.Atoi(priorityTokens[len(priorityTokens)-1])
	if err != nil {
		logrus.Errorf("Error obtained during strconv %v", err)
		return schedPolicy, schedPriority, err
	}
	logrus.Infof("Obtained scheduling priority = %d", schedPriority)

	return schedPolicy, schedPriority, nil
}
