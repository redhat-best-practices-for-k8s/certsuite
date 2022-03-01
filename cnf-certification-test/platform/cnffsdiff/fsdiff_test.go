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

package cnffsdiff_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	. "github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform/cnffsdiff"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
)

type ClientHoldersMock struct {
	stdout string
	stderr string
	err    error
}

func (o ClientHoldersMock) ExecCommandContainer(ctx clientsholder.Context, command string) (stdout, stderr string, err error) {
	stdout, stderr, err = o.stdout, o.stderr, o.err
	return stdout, stderr, err
}
func TestRunTest(t *testing.T) {
	fsdiff := &FsDiff{}
	o := ClientHoldersMock{
		stdout: "{}",
		stderr: "",
		err:    nil,
	}
	// test when no package is installed
	fsdiff.RunTest(o, clientsholder.Context{})
	assert.Equal(t, testhelper.SUCCESS, fsdiff.GetResults())

	// test when an error occurred when running the command
	o.err = errors.New("error executing the command")
	fsdiff.RunTest(&o, clientsholder.Context{})
	assert.Equal(t, testhelper.ERROR, fsdiff.GetResults())

	// test when an error message was returned
	o.err = nil
	o.stderr = "container id not found"
	fsdiff.RunTest(&o, clientsholder.Context{})
	assert.Equal(t, testhelper.ERROR, fsdiff.GetResults())

	// test when a package was installed
	o.err = nil
	o.stderr = ""
	o.stdout = `{
		changed: [
			/usr/bin/lp,
			/usr/local,
			/usr/local/bin
		],
		added: [
			/usr/local/bin/docker-entrypoint.sh
		]
	}`
	// "/usr/local/bin/docker-entrypoint.sh"
	fsdiff.RunTest(&o, clientsholder.Context{})
	assert.Equal(t, testhelper.FAILURE, fsdiff.GetResults())
}
