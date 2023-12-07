// Copyright (C) 2020-2022 Red Hat, Inc.
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

package clientsholder

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/test-network-function/cnf-certification-test/internal/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
)

//go:generate moq -out command_moq.go . Command
type Command interface {
	ExecCommandContainer(Context, string) (string, string, error)
}

// ExecCommand runs command in the pod and returns buffer output.
func (clientsholder *ClientsHolder) ExecCommandContainer(
	ctx Context, command string) (stdout, stderr string, err error) {
	commandStr := []string{"sh", "-c", command}
	var buffOut bytes.Buffer
	var buffErr bytes.Buffer
	log.Debug(fmt.Sprintf("execute command on ns=%s, pod=%s container=%s, cmd: %s", ctx.GetNamespace(), ctx.GetPodName(), ctx.GetContainerName(), strings.Join(commandStr, " ")))
	req := clientsholder.K8sClient.CoreV1().RESTClient().
		Post().
		Namespace(ctx.GetNamespace()).
		Resource("pods").
		Name(ctx.GetPodName()).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: ctx.GetContainerName(),
			Command:   commandStr,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(clientsholder.RestConfig, "POST", req.URL())
	if err != nil {
		log.Error("%v", err)
		return stdout, stderr, err
	}
	err = exec.StreamWithContext(context.TODO(), remotecommand.StreamOptions{
		Stdout: &buffOut,
		Stderr: &buffErr,
	})
	stdout, stderr = buffOut.String(), buffErr.String()
	if err != nil {
		log.Error("%v", err)
		log.Error("%v", req.URL())
		log.Error("command: %s", command)
		log.Error("stderr: %s", stderr)
		log.Error("stdout: %s", stdout)
		return stdout, stderr, err
	}
	return stdout, stderr, err
}
