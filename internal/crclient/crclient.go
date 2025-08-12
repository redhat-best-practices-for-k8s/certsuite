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

// Process represents a single operating system process within the container environment, capturing its identifier, parent identifier, namespace, and command arguments.
//
// It is used by the client to map running processes back to their originating containers and namespaces, facilitating diagnostics and audit logging. The fields are:
//
// - Pid: the unique process ID.
// - PPid: the parent process ID.
// - PidNs: the PID namespace identifier for the container.
// - Args: a string containing the command line arguments of the process.
//
// The String method formats these values into a human‑readable representation.
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

// String returns a human-readable representation of the Process.
//
// It formats key attributes of the Process into a single string,
// typically including identifiers such as the process ID and
// other relevant metadata. The returned value is suitable for
// logging or debugging purposes.
func (p *Process) String() string {
	return fmt.Sprintf("cmd: %s, pid: %d, ppid: %d, pidNs: %d", p.Args, p.Pid, p.PPid, p.PidNs)
}

// GetNodeProbePodContext creates a context for the probe pod on a node.
//
// It accepts a node name and a test environment, locates the first
// container of the probe pod running on that node, and returns a
// clientsholder.Context that can be used to execute shell commands
// against the container. If no suitable pod is found or an error
// occurs during context creation, it returns an error.
func GetNodeProbePodContext(node string, env *provider.TestEnvironment) (clientsholder.Context, error) {
	probePod := env.ProbePods[node]
	if probePod == nil {
		return clientsholder.Context{}, fmt.Errorf("probe pod not found on node %s", node)
	}

	return clientsholder.NewContext(probePod.Namespace, probePod.Name, probePod.Spec.Containers[0].Name), nil
}

// GetPidFromContainer retrieves the process ID of a running container.
//
// GetPidFromContainer extracts the PID of the specified container by executing
// an inspection command inside the container environment and parsing the
// output. It takes a Container reference and a context holding client
// connections, returning the integer PID or an error if the operation fails.
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

// GetContainerPidNamespace retrieves the PID namespace of a container.
//
// It takes a Container and TestEnvironment as inputs, determines the
// container's process ID via GetPidFromContainer or exec,
// then reads /proc/<pid>/ns/pid to obtain the namespace identifier.
// The function returns the namespace string or an error if any step fails.
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

// GetContainerProcesses retrieves the list of processes running inside a Docker container.
//
// It accepts a Container and TestEnvironment, obtains the PID namespace of the container,
// then enumerates all PIDs within that namespace to build Process structs.
// The function returns a slice of pointers to Process or an error if any step fails.
func GetContainerProcesses(container *provider.Container, env *provider.TestEnvironment) ([]*Process, error) {
	pidNs, err := GetContainerPidNamespace(container, env)
	if err != nil {
		return nil, fmt.Errorf("could not get the containers' pid namespace, err: %v", err)
	}

	return GetPidsFromPidNamespace(pidNs, container)
}

// ExecCommandContainerNSEnter executes a command inside a container's namespace using nsenter.
//
// It takes the name of a command and a pointer to a Container struct.
// The function retrieves the process ID of the target container, constructs an nsenter
// command that enters the container’s PID namespace, and runs the specified command
// there. It returns the combined standard output and error streams as a string,
// along with any execution errors. If the container cannot be located or the
// command fails to run, it returns an appropriate error message.
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

// GetPidsFromPidNamespace retrieves the process IDs belonging to a specified PID namespace within a container.
//
// It executes a command inside the given container to list all processes
// that belong to the supplied PID namespace. The function parses the
// output, converts each entry into a Process struct and returns a slice
// of these structs. If any step fails—such as executing the command,
// parsing the output, or converting values—the function returns an
// error along with a nil slice. The first argument is the target PID
// namespace name; the second argument is a pointer to the container
// in which the query should run. The return value is a slice of Process
// pointers representing each matching process and an error if one occurred.
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
