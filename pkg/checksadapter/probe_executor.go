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
	corev1 "k8s.io/api/core/v1"
)

// ProbeExecutorAdapter adapts certsuite's execution model to checks.ProbeExecutor interface.
type ProbeExecutorAdapter struct{}

// ExecCommand executes a command in the first container of the given pod.
func (p *ProbeExecutorAdapter) ExecCommand(_ context.Context, pod *corev1.Pod, command string) (stdout, stderr string, err error) {
	containerName := "container-00"
	if len(pod.Spec.Containers) > 0 {
		containerName = pod.Spec.Containers[0].Name
	}
	return p.ExecCommandInContainer(context.TODO(), pod, containerName, command)
}

// ExecCommandInContainer executes a command in a specific container of the given pod.
func (p *ProbeExecutorAdapter) ExecCommandInContainer(_ context.Context, pod *corev1.Pod, containerName, command string) (stdout, stderr string, err error) {
	clients := clientsholder.GetClientsHolder()
	cmdCtx := clientsholder.NewContext(pod.Namespace, pod.Name, containerName)
	return clients.ExecCommandContainer(cmdCtx, command)
}
