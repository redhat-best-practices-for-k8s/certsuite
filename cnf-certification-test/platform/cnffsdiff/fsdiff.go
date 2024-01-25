// Copyright (C) 2020-2024 Red Hat, Inc.
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
	"errors"
	"fmt"

	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/internal/log"
	"github.com/test-network-function/cnf-certification-test/pkg/stringhelper"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
)

const (
	partnerPodmanFolder = "/root/podman"
	tmpMountDestFolder  = "/tmp/tnf-podman"
)

var (
	nodeTmpMountFolder = "/host" + tmpMountDestFolder

	// targetFolders stores all the targetFolders that shouldn't have been
	// modified in the container. All of them exist on UBI.
	targetFolders = []string{
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
	}
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
		if stringhelper.StringInSlice(targetFolders, folder, false) {
			log.Warn("Container's folder %q is altered.", folder)
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
	err := f.installCustomPodman()
	if err != nil {
		f.Error = err
		f.result = testhelper.ERROR
		return
	}

	defer f.unmountCustomPodman()

	log.Info("Running \"podman diff\" for container id %s", containerUID)
	output, err := f.runPodmanDiff(containerUID)
	if err != nil {
		f.Error = err
		f.result = testhelper.ERROR
		return
	}

	// see if there's a match in the output
	log.Debug("Podman diff output is %s", output)

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

// Generic helper function to execute a command inside the corresponding debug pod of the
// container under test. Whatever output in stdout or stderr is considered a failure, so it will
// return the concatenation of the given errorStr with those stdout, stderr and the error string.
func (f *FsDiff) execCommandContainer(cmd, errorStr string) error {
	output, outerr, err := f.clientHolder.ExecCommandContainer(f.ctxt, cmd)
	if err != nil || output != "" || outerr != "" {
		return errors.New(errorStr + fmt.Sprintf(" Stderr: %s, Stdout: %s, Err: %v", output, outerr, err))
	}

	return nil
}

func (f *FsDiff) createNodeFolder() error {
	return f.execCommandContainer(fmt.Sprintf("mkdir %s", nodeTmpMountFolder),
		fmt.Sprintf("failed or unexpected output when creating folder %s.", nodeTmpMountFolder))
}

func (f *FsDiff) deleteNodeFolder() error {
	return f.execCommandContainer(fmt.Sprintf("rmdir %s", nodeTmpMountFolder),
		fmt.Sprintf("failed or unexpected output when deleting folder %s.", nodeTmpMountFolder))
}

func (f *FsDiff) mountDebugPartnerPodmanFolder() error {
	return f.execCommandContainer(fmt.Sprintf("mount --bind %s %s", partnerPodmanFolder, nodeTmpMountFolder),
		fmt.Sprintf("failed or unexpected output when mounting %s into %s.", partnerPodmanFolder, nodeTmpMountFolder))
}

func (f *FsDiff) unmountDebugPartnerPodmanFolder() error {
	return f.execCommandContainer(fmt.Sprintf("umount %s", nodeTmpMountFolder),
		fmt.Sprintf("failed or unexpected output when umounting %s.", nodeTmpMountFolder))
}

func (f *FsDiff) installCustomPodman() error {
	// We need to create the destination folder first.
	log.Info("Creating temp folder %s", nodeTmpMountFolder)
	if err := f.createNodeFolder(); err != nil {
		return err
	}

	// Mount podman from partner debug pod into /host/tmp/...
	log.Info("Mouting %s into %s", partnerPodmanFolder, nodeTmpMountFolder)
	if mountErr := f.mountDebugPartnerPodmanFolder(); mountErr != nil {
		// We need to delete the temp folder previously created as mount point.
		if deleteErr := f.deleteNodeFolder(); deleteErr != nil {
			return fmt.Errorf("failed to mount folder %s: %s, failed to delete %s: %s",
				partnerPodmanFolder, mountErr, nodeTmpMountFolder, deleteErr)
		}

		return mountErr
	}

	return nil
}

func (f *FsDiff) unmountCustomPodman() {
	// Unmount podman folder from host.
	log.Info("Unmounting folder %s", nodeTmpMountFolder)
	if err := f.unmountDebugPartnerPodmanFolder(); err != nil {
		// Here, there's no point on trying to remove the temp folder used as mount point, as
		// that probably won't work either.
		f.Error = err
		f.result = testhelper.ERROR
		return
	}

	log.Info("Deleting folder %s", nodeTmpMountFolder)
	if err := f.deleteNodeFolder(); err != nil {
		f.Error = err
		f.result = testhelper.ERROR
	}
}
