package ishyperthread

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
)

const (
	expectedValue        = 2
	isHyperThreadCommand = "lscpu "
)

func IsHyperThread(env *provider.TestEnvironment, nodeName string) (bool, error) {
	o := clientsholder.GetClientsHolder()
	ctx := clientsholder.NewContext(env.DebugPods[nodeName].Namespace, env.DebugPods[nodeName].Name, env.DebugPods[nodeName].Spec.Containers[0].Name)
	cmdValue, errStr, err := o.ExecCommandContainer(ctx, isHyperThreadCommand)
	if err != nil || errStr != "" {
		return false, fmt.Errorf("cannot execute %s on debug pod %s, err=%s, stderr=%s", isHyperThreadCommand, env.DebugPods[nodeName], err, errStr)
	}
	re := regexp.MustCompile(`Thread\(s\) per core:\s+(\d+)`)
	match := re.FindStringSubmatch(cmdValue)
	num := 0
	if len(match) == expectedValue {
		num, _ = strconv.Atoi(match[1])
	}
	return num > 1, nil
}

func extractNumber(str string) int {
	re := regexp.MustCompile(`\d+`)

	// Find all matches in the string
	matches := re.FindAllString(str, -1)
	num := 0
	// Loop through the matches and convert them to integers (assuming there's only one number)
	for _, match := range matches {
		num, _ = strconv.Atoi(match)
	}
	return num
}

func IsBareMetal(providerID string) bool {
	// Check if the node's providerID indicates it's a baremetalhost
	return strings.HasPrefix(providerID, "baremetalhost://")
}
