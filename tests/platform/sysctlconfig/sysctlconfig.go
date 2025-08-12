// Copyright (C) 2020-2024 Red Hat, Inc.
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

package sysctlconfig

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
)

// parseSysctlSystemOutput parses the output of "sysctl --system" and returns a map of sysctl keys to their values.
//
// It takes a single string argument containing the raw command output.
// The function splits the input into lines, ignores comments and empty lines,
// extracts key-value pairs using regular expressions, and stores them in a map.
// The returned map has string keys representing sysctl parameters and string
// values representing the corresponding settings. If parsing fails for a line,
// that entry is omitted from the result.
func parseSysctlSystemOutput(sysctlSystemOutput string) map[string]string {
	retval := make(map[string]string)
	splitConfig := strings.Split(sysctlSystemOutput, "\n")
	for _, line := range splitConfig {
		if strings.HasPrefix(line, "*") {
			continue
		}

		keyValRegexp := regexp.MustCompile(`(\S+)\s*=\s*(\S+)`) // A line is of the form "kernel.yama.ptrace_scope = 0"
		if !keyValRegexp.MatchString(line) {
			continue
		}
		regexResults := keyValRegexp.FindStringSubmatch(line)
		key := regexResults[1]
		val := regexResults[2]
		retval[key] = val
	}
	return retval
}

// GetSysctlSettings retrieves sysctl settings from a container.
//
// It accepts a TestEnvironment pointer and the name of a target container.
// The function executes a command inside that container to read all
// sysctl values, parses the output into a map of key/value strings,
// and returns this map along with any error encountered during execution
// or parsing.
func GetSysctlSettings(env *provider.TestEnvironment, nodeName string) (map[string]string, error) {
	const (
		sysctlCommand = "chroot /host sysctl --system"
	)

	o := clientsholder.GetClientsHolder()
	ctx := clientsholder.NewContext(env.ProbePods[nodeName].Namespace, env.ProbePods[nodeName].Name, env.ProbePods[nodeName].Spec.Containers[0].Name)

	outStr, errStr, err := o.ExecCommandContainer(ctx, sysctlCommand)
	if err != nil || errStr != "" {
		return nil, fmt.Errorf("failed to execute command %s in probe pod %s, err=%s, stderr=%s", sysctlCommand,
			env.ProbePods[nodeName], err, errStr)
	}

	return parseSysctlSystemOutput(outStr), nil
}
