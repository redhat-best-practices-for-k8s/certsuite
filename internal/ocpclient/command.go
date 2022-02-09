// Copyright (C) 2020-2021 Red Hat, Inc.
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

package ocpclient

import (
	"bytes"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

type Context struct {
	Namespace     string
	Podname       string
	Containername string
}

type Command interface {
	ExecCommandContainer(Context, string) (string, string, error)
}

// ExecCommand runs command in the pod and returns buffer output.
func (ocpclient OcpClient) ExecCommandContainer(
	ctx Context, command string) (stdout, stderr string, err error) {
	commandStr := []string{"sh", "-c", command}
	var buffOut bytes.Buffer
	var buffErr bytes.Buffer
	logrus.Trace(fmt.Sprintf("execute commands on ns=%s, pod=%s container=%s", ctx.Namespace, ctx.Podname, ctx.Containername))
	req := ocpclient.Coreclient.RESTClient().
		Post().
		Namespace(ctx.Namespace).
		Resource("pods").
		Name(ctx.Podname).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: ctx.Containername,
			Command:   commandStr,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(ocpclient.RestConfig, "POST", req.URL())
	if err != nil {
		logrus.Error(err)
		return stdout, stderr, err
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: &buffOut,
		Stderr: &buffErr,
		Tty:    true,
	})
	stdout, stderr = buffOut.String(), buffErr.String()
	if err != nil {
		logrus.Error(err)
		logrus.Error(req.URL())
		logrus.Error("command: ", command)
		logrus.Error("stderr: ", stderr)
		logrus.Error("stdout: ", stdout)
		return stdout, stderr, err
	}
	return stdout, stderr, err
}
