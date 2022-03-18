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
	"regexp"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
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
	result       int
	ClientHolder clientsholder.Command
}

//go:generate moq -out fsdiff_moq.go . FsDiffFuncs
type FsDiffFuncs interface {
	RunTest(ctx clientsholder.Context)
	GetResults() int
}

func NewFsDiffTester(client clientsholder.Command) *FsDiff {
	return &FsDiff{
		ClientHolder: client,
		result:       testhelper.ERROR,
	}
}

func (f *FsDiff) RunTest(ctx clientsholder.Context) {
	expected := []string{varlibrpm, varlibdpkg, bin, sbin, lib, usrbin, usrsbin, usrlib}
	output, outerr, err := f.ClientHolder.ExecCommandContainer(ctx, `chroot /host podman diff --format json`)
	if err != nil {
		logrus.Errorln("can't execute command on container ", err)
		f.result = testhelper.ERROR
		return
	}
	if outerr != "" {
		f.result = testhelper.ERROR
		logrus.Errorln("error when running fsdiff test ", outerr)
		return
	}

	for _, exp := range expected {
		// panic if the expression is wrong
		r := regexp.MustCompile(exp)
		if r.MatchString(output) {
			logrus.Error("an installed package found on ", exp)
			f.result = testhelper.FAILURE
			return
		}
	}
	// see if there's a match in the output
	logrus.Traceln("the output is ", output)
	f.result = testhelper.SUCCESS
}

func (f *FsDiff) GetResults() int {
	return f.result
}
