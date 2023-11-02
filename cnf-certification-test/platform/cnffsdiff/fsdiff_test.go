// Copyright (C) 2020-2023 Red Hat, Inc.
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
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/platform/cnffsdiff"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
)

type ClientHoldersMock struct {
	stdout string
	stderr string
	err    error
}

func (o ClientHoldersMock) ExecCommandContainer(_ clientsholder.Context, cmd string) (stdout, stderr string, err error) {
	// Filter out mkdir/rmdir and mount/umount commands.
	if !strings.Contains(cmd, "podman diff") {
		return "", "", nil
	}

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
		{ // test when folder /usr/lib has been removed
			expectedResult: testhelper.FAILURE,
			clientErr:      nil,
			clientStdErr:   "",
			clientStdOut: `{
				"changed": [
					"/usr"
				],
				"deleted": [
					"/usr/lib"
				]
			}`,
		},
		{ // test when a package "lp" was installed in /usr/bin and a file docker-entrypoint.sh
			// is created under /usr/local/bin
			expectedResult: testhelper.FAILURE,
			clientErr:      nil,
			clientStdErr:   "",
			clientStdOut: `{
				"changed": [
					"/usr",
					"/usr/bin",
					"/usr/local",
					"/usr/local/bin"
				],
				"added": [
					"/usr/bin/lp",
					"/usr/local/bin/docker-entrypoint.sh"
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

		fsdiff := cnffsdiff.NewFsDiffTester(chm, clientsholder.Context{})
		fsdiff.RunTest("fakeUID")
		assert.Equal(t, tc.expectedResult, fsdiff.GetResults())
	}
}

type ClientHoldersMountCustomPodmanMock struct {
	createFolderStdout string
	createFolderStderr string
	createFolderErr    error

	mountFolderStdout string
	mountFolderStderr string
	mountFolderErr    error

	// Since there are two calls to ExecCommandContainer inside fsdiff.RunTest(), we'll use a toggle bool
	// to control which call to ExecCommandContainer should work.
	MountPhaseReached bool
}

func (o *ClientHoldersMountCustomPodmanMock) ExecCommandContainer(_ clientsholder.Context, _ string) (stdout, stderr string, err error) {
	if o.MountPhaseReached {
		if o.mountFolderStdout != "" || o.mountFolderStderr != "" || o.mountFolderErr != nil {
			return o.mountFolderStdout, o.mountFolderStderr, o.mountFolderErr
		}
	} else {
		if o.createFolderStdout != "" || o.createFolderStderr != "" || o.createFolderErr != nil {
			return o.createFolderStdout, o.createFolderStderr, o.createFolderErr
		}
		o.MountPhaseReached = true
	}

	return "", "", nil
}

func TestRunTestMountFolderErrors(t *testing.T) {
	testCases := []struct {
		mockedClientshHolder *ClientHoldersMountCustomPodmanMock
		expectedError        string
	}{
		// Errors creating the mount point folder.
		{
			mockedClientshHolder: &ClientHoldersMountCustomPodmanMock{
				createFolderErr: fmt.Errorf("custom error"),
			},
			expectedError: "failed or unexpected output when creating folder /host/tmp/tnf-podman. Stderr: , Stdout: , Err: custom error",
		},
		{
			mockedClientshHolder: &ClientHoldersMountCustomPodmanMock{
				createFolderStdout: "custom stdout",
				createFolderStderr: "custom stderr",
				createFolderErr:    nil,
			},
			expectedError: "failed or unexpected output when creating folder /host/tmp/tnf-podman. Stderr: custom stdout, Stdout: custom stderr, Err: <nil>",
		},

		// Errors mounting the podman folder.
		{
			mockedClientshHolder: &ClientHoldersMountCustomPodmanMock{
				mountFolderErr: fmt.Errorf("custom error"),
			},
			expectedError: "failed to mount folder /root/podman: failed or unexpected output when mounting /root/podman into /host/tmp/tnf-podman. " +
				"Stderr: , Stdout: , Err: custom error, failed to delete /host/tmp/tnf-podman: failed or unexpected output when deleting folder /host/tmp/tnf-podman. Stderr: , Stdout: , Err: custom error",
		},
		{
			mockedClientshHolder: &ClientHoldersMountCustomPodmanMock{
				mountFolderStdout: "custom stdout",
				mountFolderStderr: "custom stderr",
				mountFolderErr:    nil,
			},
			expectedError: "failed to mount folder /root/podman: failed or unexpected output when mounting /root/podman into /host/tmp/tnf-podman. " +
				"Stderr: custom stdout, Stdout: custom stderr, Err: <nil>, failed to delete /host/tmp/tnf-podman: failed or unexpected output when deleting folder /host/tmp/tnf-podman. Stderr: custom stdout, Stdout: custom stderr, Err: <nil>",
		},
	}

	for _, tc := range testCases {
		fsdiff := cnffsdiff.NewFsDiffTester(tc.mockedClientshHolder, clientsholder.Context{})
		fsdiff.RunTest("fakeUID")
		assert.Equal(t, testhelper.ERROR, fsdiff.GetResults())
		assert.Equal(t, fsdiff.Error.Error(), tc.expectedError)
	}
}

type ClientHoldersUnmountCustomPodmanMock struct {
	unmountFolderStdout string
	unmountFolderStderr string
	unmountFolderErr    error

	deleteFolderStdout string
	deleteFolderStderr string
	deleteFolderErr    error

	// Since there are two calls to ExecCommandContainer inside fsdiff.RunTest(), we'll use a toggle bool
	// to control which call to ExecCommandContainer should work.
	DeletePhaseReached bool
}

func (o *ClientHoldersUnmountCustomPodmanMock) ExecCommandContainer(_ clientsholder.Context, cmd string) (stdout, stderr string, err error) {
	// To reach the unmount/delete folder at the end, we need to make the mount operation and the podman diff to return no errors.
	if strings.Contains(cmd, "mount --bind") || strings.Contains(cmd, "mkdir") {
		return "", "", nil
	}

	if strings.Contains(cmd, "podman diff") {
		return "{}", "", nil
	}

	if o.DeletePhaseReached {
		if o.deleteFolderStdout != "" || o.deleteFolderStderr != "" || o.deleteFolderErr != nil {
			return o.deleteFolderStdout, o.deleteFolderStderr, o.deleteFolderErr
		}
	} else {
		if o.unmountFolderStdout != "" || o.unmountFolderStderr != "" || o.unmountFolderErr != nil {
			return o.unmountFolderStdout, o.unmountFolderStderr, o.unmountFolderErr
		}
		o.DeletePhaseReached = true
	}

	return "", "", nil
}

func TestRunTestUnmountFolderErrors(t *testing.T) {
	testCases := []struct {
		mockedClientshHolder *ClientHoldersUnmountCustomPodmanMock
		expectedError        string
	}{
		// Errors unmounting the podman folder.
		{
			mockedClientshHolder: &ClientHoldersUnmountCustomPodmanMock{
				unmountFolderErr: fmt.Errorf("custom error"),
			},
			expectedError: "failed or unexpected output when umounting /host/tmp/tnf-podman. Stderr: , Stdout: , Err: custom error",
		},
		{
			mockedClientshHolder: &ClientHoldersUnmountCustomPodmanMock{
				unmountFolderStdout: "custom stdout",
				unmountFolderStderr: "custom stderr",
			},
			expectedError: "failed or unexpected output when umounting /host/tmp/tnf-podman. Stderr: custom stdout, Stdout: custom stderr, Err: <nil>",
		},

		// Errors deleting the mount point folder.
		{
			mockedClientshHolder: &ClientHoldersUnmountCustomPodmanMock{
				deleteFolderErr: fmt.Errorf("custom error"),
			},
			expectedError: "failed or unexpected output when deleting folder /host/tmp/tnf-podman. Stderr: , Stdout: , Err: custom error",
		},
		{
			mockedClientshHolder: &ClientHoldersUnmountCustomPodmanMock{
				deleteFolderStdout: "custom stdout",
				deleteFolderStderr: "custom stderr",
			},
			expectedError: "failed or unexpected output when deleting folder /host/tmp/tnf-podman. Stderr: custom stdout, Stdout: custom stderr, Err: <nil>",
		},
	}

	for _, tc := range testCases {
		fsdiff := cnffsdiff.NewFsDiffTester(tc.mockedClientshHolder, clientsholder.Context{})
		fsdiff.RunTest("fakeUID")
		assert.Equal(t, testhelper.ERROR, fsdiff.GetResults())
		assert.Equal(t, fsdiff.Error.Error(), tc.expectedError)
	}
}
