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

package platform

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/ocpclient"
	. "github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

const (
	command               = "salah"
	varlibrpm             = `(?m)[\t|\s]\/var\/lib\/rpm[.]*`
	varlibdpkg            = `(?m)[\t|\s]\/var\/lib\/dpkg[.]*`
	bin                   = `(?m)[\t|\s]\/bin[.]*`
	sbin                  = `(?m)[\t|\s]\/sbin[.]*`
	lib                   = `(?m)[\t|\s]\/lib[.]*`
	usrbin                = `(?m)[\t|\s]\/usr\/bin[.]*`
	usrsbin               = `(?m)[\t|\s]\/usr\/sbin[.]*`
	usrlib                = `(?m)[\t|\s]\/usr\/lib[.]*`
	successfulOutputRegex = `(?m){}`
	acceptAllRegex        = `(?m)(.|\n)+`
)

type FsDiff struct {
	containerId string
	command     []string
	result      int
}

func NewFsDiff(c *Container) *FsDiff {
	id := c.Status.ContainerID
	split := strings.Split(id, "://")
	uid := ""
	if len(split) > 0 {
		uid = split[len(split)-1]
	}
	if uid == "" {
		logrus.Debugln(fmt.Sprintf("could not find uid of %s/%s/%s\n", c.Namespace, c.Podname, &c.Data.Name))
		return nil
	}
	command := []string{"chroot", "/host", "podman", "diff", "--format", "json", uid}
	return &FsDiff{
		containerId: uid,
		command:     command,
		result:      tnf.ERROR,
	}
}

func (f *FsDiff) RunTest(o ocpclient.OcpClient, ctx *Context) {
	output, outerr, err := o.ExecCommandContainer(ctx.Namespace, ctx.Podname, ctx.Containername, f.command)
	if err != nil {
		f.result = tnf.ERROR
	}
	if outerr != "" {
		f.result = tnf.ERROR
		logrus.Errorln("error when running fsdiff test ", outerr)
	}
	// see if there's a match in the output
	logrus.Debugln("the output is ", output)
	f.result = tnf.SUCCESS
}

func (f *FsDiff) getResults() int {
	return f.result
}
