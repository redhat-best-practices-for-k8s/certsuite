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

package cnffsdiff

import (
	"encoding/json"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
)

const (
	partnerPodmanFolder = "/root/podman"
	tmpMountDestFolder  = "/tmp/tnf-podman"
)

var (
	// targetFolders stores all the targetFolders that shouldn't have been
	// modified in the container. All of them exist on UBI.
	targetFolders = mapset.NewSet(
		"/bin",
		"/lib",
		"/lib64",
		"/sbin",
		"/usr/bin",
		"/usr/lib",
		"/usr/lib64",
		"/usr/sbin",
		"/var/lib/rpm",
		"/var/lib/dpkg",
	)
)

// fsDiffJSON is a helper struct to unmarshall the "podman diff --format json" output: a slice of
// folders/filepaths (strings) for each event type changed/added/deleted:
//
//	{"changed": ["folder1, folder2"], added": ["folder5", "folder6"], "deleted": ["folder3", "folder4"]"}
//
// We'll only care about deleted and changed types, though, as in case a folder/file is created to any of them,
// there will be two entries, one for the "added" and another for the "changed".
type fsDiffJSON struct {
	Changed []string `json:"changed"`
	Deleted []string `json:"deleted"`
	Added   []string `json:"added"` // Will not be checked, but let's keep it just in case.
}

type FsDiff struct {
	result       int
	clientHolder clientsholder.Command
	ctxt         clientsholder.Context

	DeletedFolders []string
	ChangedFolders []string
	Error          error
}

type FsDiffFuncs interface {
	RunTest(ctx clientsholder.Context, containerUID string)
	GetResults() int
}

func NewFsDiffTester(client clientsholder.Command, ctxt clientsholder.Context) *FsDiff {
	return &FsDiff{
		clientHolder: client,
		ctxt:         ctxt,
		result:       testhelper.ERROR,
	}
}

func intersectTargetFolders(src []string) []string {
	var dst []string
	for _, folder := range src {
		if targetFolders.Contains(folder) {
			logrus.Tracef("Container's folder %s is altered.", folder)
			dst = append(dst, folder)
		}
	}
	return dst
}

func (f *FsDiff) runPodmanDiff(containerUID string) (string, error) {
	output, outerr, err := f.clientHolder.ExecCommandContainer(f.ctxt, fmt.Sprintf("chroot /host %s/podman diff --format json %s", tmpMountDestFolder, containerUID))
	if err != nil {
		return "", fmt.Errorf("can not execute command on container: %w", err)
	}
	if outerr != "" {
		return "", fmt.Errorf("stderr log received when running fsdiff test: %s", outerr)
	}
	return output, nil
}

func (f *FsDiff) RunTest(containerUID string) {
	err := f.mountCustomPodman()
	if err != nil {
		f.Error = err
		f.result = testhelper.ERROR
		return
	}

	defer f.unmountCustomPodman()

	output, err := f.runPodmanDiff(containerUID)
	if err != nil {
		f.Error = err
		f.result = testhelper.ERROR
		return
	}

	// see if there's a match in the output
	logrus.Traceln("Podman diff output is ", output)

	diff := fsDiffJSON{}
	err = json.Unmarshal([]byte(output), &diff)
	if err != nil {
		f.Error = fmt.Errorf("failed to unmarshall podman diff's json output: %s, err: %w", output, err)
		f.result = testhelper.ERROR
		return
	}
	f.DeletedFolders = intersectTargetFolders(diff.Deleted)
	f.ChangedFolders = intersectTargetFolders(diff.Changed)
	if len(f.ChangedFolders) != 0 || len(f.DeletedFolders) != 0 {
		f.result = testhelper.FAILURE
	} else {
		f.result = testhelper.SUCCESS
	}
}

func (f *FsDiff) GetResults() int {
	return f.result
}

func (f *FsDiff) mountCustomPodman() error {
	nodeTmpMountFolder := "/host" + tmpMountDestFolder

	// We need to create the destination folder first.
	logrus.Infof("Creating temp folder %s", nodeTmpMountFolder)
	output, outerr, err := f.clientHolder.ExecCommandContainer(f.ctxt, fmt.Sprintf("mkdir %s", nodeTmpMountFolder))
	if err != nil {
		return fmt.Errorf("failed to create folder %s: %v", nodeTmpMountFolder, err)
	}

	if output != "" || outerr != "" {
		return fmt.Errorf("unexpected output when creating folder %s. Stdout: %s - StdErr: %s",
			nodeTmpMountFolder, output, outerr)
	}

	// Mount podman from partner debug pod into /host/tmp/...
	logrus.Infof("Mouting %s into %s", partnerPodmanFolder, nodeTmpMountFolder)
	output, outerr, err = f.clientHolder.ExecCommandContainer(f.ctxt, fmt.Sprintf("mount --bind %s %s", partnerPodmanFolder, nodeTmpMountFolder))
	if err != nil {
		return fmt.Errorf("failed to mount %s into %s: %v", partnerPodmanFolder, nodeTmpMountFolder, err)
	}

	if output != "" || outerr != "" {
		return fmt.Errorf("unexpected output when mounting %s into %s. Stdout: %s - StdErr: %s",
			partnerPodmanFolder, nodeTmpMountFolder, output, outerr)
	}

	return nil
}

func (f *FsDiff) unmountCustomPodman() {
	nodeTmpMountFolder := "/host" + tmpMountDestFolder

	// Unmount podman folder from host.
	logrus.Infof("Unmounting folder %s", nodeTmpMountFolder)
	output, outerr, err := f.clientHolder.ExecCommandContainer(f.ctxt, fmt.Sprintf("umount %s", nodeTmpMountFolder))
	if err != nil {
		f.Error = fmt.Errorf("failed to unmount %s: %v", nodeTmpMountFolder, err)
		f.result = testhelper.ERROR
		return
	}

	if output != "" || outerr != "" {
		f.Error = fmt.Errorf("unexpected output when unmounting %s. Stdout: %s - StdErr: %s",
			nodeTmpMountFolder, output, outerr)
		f.result = testhelper.ERROR
		return
	}

	logrus.Infof("Deleting folder %s", nodeTmpMountFolder)
	output, outerr, err = f.clientHolder.ExecCommandContainer(f.ctxt, fmt.Sprintf("rmdir %s", nodeTmpMountFolder))
	if err != nil {
		f.Error = fmt.Errorf("failed to delete folder %s: %v", nodeTmpMountFolder, err)
		f.result = testhelper.ERROR
		return
	}

	if output != "" || outerr != "" {
		f.Error = fmt.Errorf("unexpected output when deleting folder %s. Stdout: %s - StdErr: %s",
			nodeTmpMountFolder, output, outerr)
		f.result = testhelper.ERROR
	}
}
