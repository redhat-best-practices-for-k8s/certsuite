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
	testCases := []struct {
		clientErr      error
		clientStdErr   string
		clientStdOut   string
		expectedResult int
	}{
		{ // test when no package is installed
			expectedResult: testhelper.SUCCESS,
			clientStdOut:   "{}",
			clientStdErr:   "",
		},
		{ // test when an error occurred when running the command
			expectedResult: testhelper.ERROR,
			clientErr:      errors.New("error executing the command"),
		},
		{ // test when an error message was returned
			expectedResult: testhelper.ERROR,
			clientErr:      nil,
			clientStdErr:   "container id not found",
		},
		{ // test when a package was installed
			expectedResult: testhelper.FAILURE,
			clientErr:      nil,
			clientStdErr:   "",
			clientStdOut: `{
				changed: [
					/usr/bin/lp,
					/usr/local,
					/usr/local/bin
				],
				added: [
					/usr/local/bin/docker-entrypoint.sh
				]
			}`,
		},
	}

	for _, tc := range testCases {
		chm := &ClientHoldersMock{
			stdout: tc.clientStdOut,
			stderr: tc.clientStdErr,
			err:    tc.clientErr,
		}

		fsdiff := NewFsDiffTester(chm)
		fsdiff.RunTest(clientsholder.Context{}, "fakeUID")
		assert.Equal(t, tc.expectedResult, fsdiff.GetResults())
	}
}
