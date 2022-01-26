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

// ExecCommand runs command in the pod and returns buffer output.
func (ocpclient OcpClient) ExecCommandContainer(
	namespace, podname, container string, command []string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	logrus.Debug(fmt.Sprintf("execute commands on ns=%s, pod=%s container=%s", namespace, podname, container))
	req := ocpclient.ClientConfig.RESTClient().
		Post().
		Namespace(namespace).
		Resource("pods").
		Name(podname).
		SubResource("exec")
	url := req.URL()
	logrus.Debugln("url first ", url)
	op := &v1.PodExecOptions{
		Container: container,
		Command:   []string{"date"},
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}
	req = req.VersionedParams(op, scheme.ParameterCodec)
	url = req.URL()
	logrus.Debugln("url second ", url)
	_, err := url.Parse("https://api.clus0.t5g.lab.eng.rdu2.redhat.com:6443/api/v1/namespaces/default/pods/debug-mbkt4/exec?command=chroot&command=%2Fhost&command=podman&command=diff&command=--format&command=json&command=d393ca003de01497722036281beb2385cfaf82f1854bd0ceec01217ef14c5a59&container=container-00&stderr=true&stdin=true&stdout=true")
	if err != nil {
		logrus.Error("can't parse url")
	}

	exec, err := remotecommand.NewSPDYExecutor(ocpclient.RestConfig, "POST", req.URL())
	if err != nil {
		logrus.Error(err)
		return stdout.String(), stderr.String(), err
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    true,
	})
	if err != nil {
		logrus.Error(err)
		logrus.Error(req.URL())
		logrus.Error("command ", command)
		return stdout.String(), stderr.String(), err
	}
	return stdout.String(), stderr.String(), err
}
