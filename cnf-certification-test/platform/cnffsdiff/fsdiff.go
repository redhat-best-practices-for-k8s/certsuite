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

package cnffsdiff

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/ocpclient"
	. "github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

const (
	varlibrpm  = `(?m)[\t|\s]\/var\/lib\/rpm[.]*`
	varlibdpkg = `(?m)[\t|\s]\/var\/lib\/dpkg[.]*`
	bin        = `(?m)[\t|\s]\/bin[.]*`
	sbin       = `(?m)[\t|\s]\/sbin[.]*`
	lib        = `(?m)[\t|\s]\/lib[.]*`
	usrbin     = `(?m)[\t|\s]\/usr\/bin[.]*`
	usrsbin    = `(?m)[\t|\s]\/usr\/sbin[.]*`
	usrlib     = `(?m)[\t|\s]\/usr\/lib[.]*`
)

type FsDiff struct {
	containerId string
	command     []string
	result      int
}

func NewFsDiff(c *Container) (*FsDiff, error) {
	id := c.Status.ContainerID
	split := strings.Split(id, "://")
	uid := ""
	if len(split) > 0 {
		uid = split[len(split)-1]
	}
	if uid == "" {
		logrus.Debugln(fmt.Sprintf("could not find uid of %s/%s/%s\n", c.Namespace, c.Podname, c.Data.Name))
		return nil, errors.New("Can't instantiante FsDiff instante")
	}
	logrus.Debugln(fmt.Sprintf("uid of %s/%s/%s=%s\n", c.Namespace, c.Podname, c.Data.Name, uid))
	commands := []string{"chroot", "/host", "podman", "diff", "--format", "json", uid}
	return &FsDiff{
		containerId: uid,
		command:     commands,
		result:      tnf.ERROR,
	}, nil
}

func (f *FsDiff) RunTest(o ocpclient.Command, ctx ocpclient.Context) {

	expected := []string{varlibrpm, varlibdpkg, bin, sbin, lib, usrbin, usrsbin, usrlib}

	output, outerr, err := o.ExecCommandContainer(ctx, f.command)
	if err != nil {
		logrus.Errorln("can't execute command on container ", err)
		f.result = tnf.ERROR
		return
	}
	if outerr != "" {
		f.result = tnf.ERROR
		logrus.Errorln("error when running fsdiff test ", outerr)
		return
	}

	for _, exp := range expected {
		// panic if the expression is wrong
		r := regexp.MustCompile(exp)
		if r.Match([]byte(output)) {
			logrus.Error("an installed package found on ", exp)
			f.result = tnf.FAILURE
			return
		}
	}
	// see if there's a match in the output
	logrus.Traceln("the output is ", output)
	f.result = tnf.SUCCESS
}

func (f *FsDiff) GetResults() int {
	return f.result
}
