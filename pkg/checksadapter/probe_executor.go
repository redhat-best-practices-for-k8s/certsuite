// Copyright (C) 2020-2026 Red Hat, Inc.
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

package checksadapter

import (
	"context"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider"
	corev1 "k8s.io/api/core/v1"
)

// ProbeExecutorAdapter adapts certsuite's execution model to checks.ProbeExecutor interface.
type ProbeExecutorAdapter struct {
	env *provider.TestEnvironment
}

// ExecCommand executes a command in the probe pod and returns stdout, stderr, and error.
func (p *ProbeExecutorAdapter) ExecCommand(ctx context.Context, pod *corev1.Pod, command string) (stdout, stderr string, err error) {
	// Use certsuite's existing pod execution infrastructure
	clients := clientsholder.GetClientsHolder()

	// Create a context for the command execution using the constructor
	containerName := "container-00" // Default probe container name
	if len(pod.Spec.Containers) > 0 {
		containerName = pod.Spec.Containers[0].Name
	}

	cmdCtx := clientsholder.NewContext(pod.Namespace, pod.Name, containerName)

	// Execute the command
	return clients.ExecCommandContainer(cmdCtx, command)
}
