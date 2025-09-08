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

// parseSysctlSystemOutput parses sysctl output into a map of key-value pairs
//
// The function takes the raw text returned by "sysctl --system" and splits it
// line by line. It ignores comment lines that start with an asterisk, then uses
// a regular expression to extract keys and values from standard assignments
// such as "kernel.yama.ptrace_scope = 0". Each extracted key and value is
// stored in a map which the function returns.
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

// GetSysctlSettings Retrieves system configuration values from a node's probe pod
//
// This function runs the command "chroot /host sysctl --system" inside a
// designated probe container to collect kernel settings for a specified node.
// It captures standard output and parses each line into key/value pairs,
// ignoring comments or non‑matching lines. The resulting map of setting names
// to values is returned, with an error if the command fails or produces
// unexpected output.
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
