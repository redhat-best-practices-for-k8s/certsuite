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

package crclient

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
)

const PsRegex = `(?m)^(\d+?)\s+?(\d+?)\s+?(\d+?)\s+?(.*?)$`

// Process Represents a running process inside a container
//
// This structure holds the identifier, parent identifier, namespace, and
// command line arguments for a single operating system process discovered
// within a container’s PID namespace. The fields enable callers to
// distinguish processes by their unique IDs and to trace relationships between
// child and parent processes during diagnostics.
type Process struct {
	PidNs, Pid, PPid int
	Args             string
}

const (
	DevNull           = " 2>/dev/null"
	DockerInspectPID  = "chroot /host docker inspect -f '{{.State.Pid}}' "
	RetryAttempts     = 5
	RetrySleepSeconds = 3
)

// Process.String Formats the process details into a readable string
//
// This method creates a human‑readable representation of a process by
// combining its command line arguments and identifiers. It uses string
// formatting to include the executable name, process ID, parent process ID, and
// PID namespace number in a single line. The resulting string is returned for
// logging or debugging purposes.
func (p *Process) String() string {
	return fmt.Sprintf("cmd: %s, pid: %d, ppid: %d, pidNs: %d", p.Args, p.Pid, p.PPid, p.PidNs)
}

// GetNodeProbePodContext creates a context for the first container of a probe pod on a node
//
// The function looks up the probe pod assigned to the specified node from the
// test environment. If found, it constructs a clientsholder.Context using that
// pod’s namespace, name, and its first container’s name. The returned
// context is used to execute commands inside the probe pod; if no probe pod
// exists on the node an error is returned.
func GetNodeProbePodContext(node string, env *provider.TestEnvironment) (clientsholder.Context, error) {
	probePod := env.ProbePods[node]
	if probePod == nil {
		return clientsholder.Context{}, fmt.Errorf("probe pod not found on node %s", node)
	}

	return clientsholder.NewContext(probePod.Namespace, probePod.Name, probePod.Spec.Containers[0].Name), nil
}

// GetPidFromContainer Retrieves the process ID of a container by executing a runtime-specific command
//
// The function determines which container runtime is in use and builds an
// appropriate shell command to query the container's PID, then runs that
// command inside a probe pod context. It returns the numeric PID if the command
// succeeds or an error if execution fails or the runtime is unsupported.
func GetPidFromContainer(cut *provider.Container, ctx clientsholder.Context) (int, error) {
	var pidCmd string

	switch cut.Runtime {
	case "docker":
		pidCmd = DockerInspectPID + cut.UID + DevNull
	case "docker-pullable":
		pidCmd = DockerInspectPID + cut.UID + DevNull
	case "cri-o", "containerd":
		pidCmd = "chroot /host crictl inspect --output go-template --template '{{.info.pid}}' " + cut.UID + DevNull
	default:
		log.Debug("Container runtime %s not supported yet for this test, skipping", cut.Runtime)
		return 0, fmt.Errorf("container runtime %s not supported", cut.Runtime)
	}

	ch := clientsholder.GetClientsHolder()
	outStr, errStr, err := ch.ExecCommandContainer(ctx, pidCmd)
	if err != nil {
		return 0, fmt.Errorf("cannot execute command: \" %s \"  on %s err:%s", pidCmd, cut, err)
	}
	if errStr != "" {
		return 0, fmt.Errorf("cmd: \" %s \" on %s returned %s", pidCmd, cut, errStr)
	}

	return strconv.Atoi(strings.TrimSuffix(outStr, "\n"))
}

// GetContainerPidNamespace Retrieves the PID namespace identifier for a container
//
// This function determines the process ID of a target container by executing an
// inspection command on its runtime environment. It then runs a namespace
// listing command against that PID to extract the namespace name, returning it
// as a string. Errors from context retrieval, PID extraction, or command
// execution are wrapped and returned with descriptive messages.
func GetContainerPidNamespace(testContainer *provider.Container, env *provider.TestEnvironment) (string, error) {
	// Get the container pid
	ocpContext, err := GetNodeProbePodContext(testContainer.NodeName, env)
	if err != nil {
		return "", fmt.Errorf("failed to get probe pod's context for container %s: %v", testContainer, err)
	}

	pid, err := GetPidFromContainer(testContainer, ocpContext)
	if err != nil {
		return "", fmt.Errorf("unable to get container process id due to: %v", err)
	}
	log.Debug("Obtained process id for %s is %d", testContainer, pid)

	command := fmt.Sprintf("lsns -p %d -t pid -n", pid)
	stdout, stderr, err := clientsholder.GetClientsHolder().ExecCommandContainer(ocpContext, command)
	if err != nil || stderr != "" {
		return "", fmt.Errorf("unable to run nsenter due to : %v", err)
	}

	return strings.Fields(stdout)[0], nil
}

// GetContainerProcesses Retrieves all process information from a container's PID namespace
//
// The function first determines the PID namespace of the given container, then
// queries that namespace to list every running process. It returns a slice of
// Process structures containing each process's ID, parent ID, command line and
// namespace identifier, or an error if either step fails.
func GetContainerProcesses(container *provider.Container, env *provider.TestEnvironment) ([]*Process, error) {
	pidNs, err := GetContainerPidNamespace(container, env)
	if err != nil {
		return nil, fmt.Errorf("could not get the containers' pid namespace, err: %v", err)
	}

	return GetPidsFromPidNamespace(pidNs, container)
}

// ExecCommandContainerNSEnter Executes a shell command inside a container’s namespace
//
// The function determines the PID of the target container, builds an nsenter
// command to run in that process’s namespace, and executes it on a probe pod
// with retry logic. It returns the standard output, standard error, and any
// execution error. If the probe context or PID retrieval fails, it reports an
// appropriate error.
func ExecCommandContainerNSEnter(command string,
	aContainer *provider.Container) (outStr, errStr string, err error) {
	env := provider.GetTestEnvironment()
	ctx, err := GetNodeProbePodContext(aContainer.NodeName, &env)
	if err != nil {
		return "", "", fmt.Errorf("failed to get probe pod's context for container %s: %v", aContainer, err)
	}

	ch := clientsholder.GetClientsHolder()

	// Get the container PID to build the nsenter command
	containerPid, err := GetPidFromContainer(aContainer, ctx)
	if err != nil {
		return "", "", fmt.Errorf("cannot get PID from: %s, err: %v", aContainer, err)
	}

	// Add the container PID and the specific command to run with nsenter
	nsenterCommand := "nsenter -t " + strconv.Itoa(containerPid) + " -n " + command

	// Run the nsenter command on the probe pod with retry logic
	for attempt := 1; attempt <= RetryAttempts; attempt++ {
		outStr, errStr, err = ch.ExecCommandContainer(ctx, nsenterCommand)
		if err == nil {
			break
		}
		if attempt < RetryAttempts {
			time.Sleep(RetrySleepSeconds * time.Second)
		}
	}
	if err != nil {
		return "", "", fmt.Errorf("cannot execute command: \" %s \"  on %s err:%s", command, aContainer, err)
	}

	return outStr, errStr, err
}

// GetPidsFromPidNamespace Retrieves processes running in a specific PID namespace
//
// The function runs a ps command inside the probe pod on the container's node
// to list all processes with their namespaces, then filters those whose pidns
// matches the supplied string. It parses each line of output, converting
// numeric fields to integers and constructs Process objects for matching
// entries. The resulting slice of Process pointers is returned; if any error
// occurs during execution or parsing, an error value is returned.
func GetPidsFromPidNamespace(pidNamespace string, container *provider.Container) (p []*Process, err error) {
	const command = "trap \"\" SIGURG ; ps -e -o pidns,pid,ppid,args"
	env := provider.GetTestEnvironment()
	ctx, err := GetNodeProbePodContext(container.NodeName, &env)
	if err != nil {
		return nil, fmt.Errorf("failed to get probe pod's context for container %s: %v", container, err)
	}

	stdout, stderr, err := clientsholder.GetClientsHolder().ExecCommandContainer(ctx, command)
	if err != nil || stderr != "" {
		return nil, fmt.Errorf("command %q failed to run in probe pod=%s (node=%s): %v", command, ctx.GetPodName(), container.NodeName, err)
	}

	re := regexp.MustCompile(PsRegex)
	matches := re.FindAllStringSubmatch(stdout, -1)
	// If we do not find a successful log, we fail
	for _, v := range matches {
		// Matching only the right PidNs
		if pidNamespace != v[1] {
			continue
		}
		aPidNs, err := strconv.Atoi(v[1])
		if err != nil {
			log.Error("could not convert string %s to integer, err=%s", v[1], err)
			continue
		}
		aPid, err := strconv.Atoi(v[2])
		if err != nil {
			log.Error("could not convert string %s to integer, err=%s", v[2], err)
			continue
		}
		aPPid, err := strconv.Atoi(v[3])
		if err != nil {
			log.Error("could not convert string %s to integer, err=%s", v[3], err)
			continue
		}
		p = append(p, &Process{PidNs: aPidNs, Pid: aPid, Args: v[4], PPid: aPPid})
	}
	return p, nil
}
