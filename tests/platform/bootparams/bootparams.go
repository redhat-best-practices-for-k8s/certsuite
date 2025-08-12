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

package bootparams

import (
	"fmt"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/arrayhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
)

const (
	grubKernelArgsCommand = "cat /host/boot/loader/entries/$(ls /host/boot/loader/entries/ | sort | tail -n 1)"
	kernelArgscommand     = "cat /host/proc/cmdline"
)

// TestBootParamsHelper validates kernel boot parameters in a test environment.
//
// It receives a test environment, a container instance, and a logger.
// The function retrieves the current kernel command line arguments from the
// container and compares them against expected values obtained via
// GetMcKernelArguments and getGrubKernelArgs. If discrepancies are found,
// errors are logged with Errorf; non-critical mismatches generate warnings
// using Warn. Successful checks result in no error being returned.
func TestBootParamsHelper(env *provider.TestEnvironment, cut *provider.Container, logger *log.Logger) error {
	probePod := env.ProbePods[cut.NodeName]
	if probePod == nil {
		return fmt.Errorf("probe pod for container %s not found on node %s", cut, cut.NodeName)
	}
	mcKernelArgumentsMap := GetMcKernelArguments(env, cut.NodeName)
	currentKernelArgsMap, err := getCurrentKernelCmdlineArgs(env, cut.NodeName)
	if err != nil {
		return fmt.Errorf("error getting kernel cli arguments from container: %s, err=%s", cut, err)
	}
	grubKernelConfigMap, err := getGrubKernelArgs(env, cut.NodeName)
	if err != nil {
		return fmt.Errorf("error getting grub  kernel arguments for node: %s, err=%s", cut.NodeName, err)
	}
	for key, mcVal := range mcKernelArgumentsMap {
		if currentVal, ok := currentKernelArgsMap[key]; ok {
			if currentVal != mcVal {
				logger.Warn("%s KernelCmdLineArg %q does not match MachineConfig value: %q!=%q",
					cut.NodeName, key, currentVal, mcVal)
			} else {
				logger.Debug("%s KernelCmdLineArg==mcVal %q: %q==%q", cut.NodeName, key, currentVal, mcVal)
			}
		}
		if grubVal, ok := grubKernelConfigMap[key]; ok {
			if grubVal != mcVal {
				logger.Warn("%s NodeGrubKernelArgs %q does not match MachineConfig value: %q!=%q",
					cut.NodeName, key, mcVal, grubVal)
			} else {
				logger.Debug("%s NodeGrubKernelArg==mcVal %q: %q==%q", cut.NodeName, key, grubVal, mcVal)
			}
		}
	}
	return nil
}

// GetMcKernelArguments retrieves a map of kernel argument names to values
// for the given test environment and command string.
//
// It accepts a pointer to a TestEnvironment and a command string, then parses
// the command into individual arguments. The resulting list is converted to a
// map where each key is an argument name and the value is its associated
// parameter or an empty string if none. The function returns this map for use
// in test setup or validation.
func GetMcKernelArguments(env *provider.TestEnvironment, nodeName string) (aMap map[string]string) {
	mcKernelArgumentsMap := arrayhelper.ArgListToMap(env.Nodes[nodeName].Mc.Spec.KernelArguments)
	return mcKernelArgumentsMap
}

// getGrubKernelArgs retrieves GRUB kernel arguments from a container.
//
// It runs the specified command inside the test environment's container,
// parses the output to extract kernel arguments, and returns them as
// a map of argument names to values. The function accepts a TestEnvironment
// pointer and the command string to execute. On success it returns the
// map and nil error; on failure it returns an empty map and an error
// describing what went wrong.
func getGrubKernelArgs(env *provider.TestEnvironment, nodeName string) (aMap map[string]string, err error) {
	o := clientsholder.GetClientsHolder()
	ctx := clientsholder.NewContext(env.ProbePods[nodeName].Namespace, env.ProbePods[nodeName].Name, env.ProbePods[nodeName].Spec.Containers[0].Name)
	bootConfig, errStr, err := o.ExecCommandContainer(ctx, grubKernelArgsCommand)
	if err != nil || errStr != "" {
		return aMap, fmt.Errorf("cannot execute %s on probe pod %s, err=%s, stderr=%s", grubKernelArgsCommand, env.ProbePods[nodeName], err, errStr)
	}

	splitBootConfig := strings.Split(bootConfig, "\n")
	filteredBootConfig := arrayhelper.FilterArray(splitBootConfig, func(line string) bool {
		return strings.HasPrefix(line, "options")
	})
	if len(filteredBootConfig) != 1 {
		return aMap, fmt.Errorf("filteredBootConfig!=1")
	}
	grubKernelConfig := filteredBootConfig[0]
	grubSplitKernelConfig := strings.Split(grubKernelConfig, " ")
	grubSplitKernelConfig = grubSplitKernelConfig[1:]
	return arrayhelper.ArgListToMap(grubSplitKernelConfig), nil
}

// getCurrentKernelCmdlineArgs retrieves the kernel command line arguments from a container.
// 
// It executes the specified kernel argument command inside the test environment's container,
// parses the output into key/value pairs, and returns them as a map along with any error that
// occurs during execution or parsing. The function takes a TestEnvironment pointer to access
// the container context and a string specifying which kernel arguments to retrieve. It returns
// a map of argument names to values and an error if the command fails or the output cannot be parsed.
func getCurrentKernelCmdlineArgs(env *provider.TestEnvironment, nodeName string) (aMap map[string]string, err error) {
	o := clientsholder.GetClientsHolder()
	ctx := clientsholder.NewContext(env.ProbePods[nodeName].Namespace, env.ProbePods[nodeName].Name, env.ProbePods[nodeName].Spec.Containers[0].Name)
	currentKernelCmdlineArgs, errStr, err := o.ExecCommandContainer(ctx, kernelArgscommand)
	if err != nil || errStr != "" {
		return aMap, fmt.Errorf("cannot execute %s on probe pod container %s, err=%s, stderr=%s", grubKernelArgsCommand, env.ProbePods[nodeName].Name, err, errStr)
	}
	currentSplitKernelCmdlineArgs := strings.Split(strings.TrimSuffix(currentKernelCmdlineArgs, "\n"), " ")
	return arrayhelper.ArgListToMap(currentSplitKernelCmdlineArgs), nil
}
