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
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/testhelper"
)

var (
	// targetFolders stores all the targetFolders that shouldn't have been modified in the container.
	// All of them exist on UBI. It's a map just for convenience.
	targetFolders = map[string]bool{
		"/var/lib/rpm":  true,
		"/var/lib/dpkg": true,
		"/bin":          true,
		"/sbin":         true,
		"/lib":          true,
		"/lib64":        true,
		"/usr/bin":      true,
		"/usr/sbin":     true,
		"/usr/lib":      true,
		"/usr/lib64":    true,
	}
)

// fsDiffJSON is a helper struct to unmarshall the "podman diff --format json" output: a slice of
// folders/filepaths (strings) for each event type changed/added/deleted:
//  {"changed": ["folder1, folder2"], added": ["folder5", "folder6"], "deleted": ["folder3", "folder4"]"}
// We'll only care about deleted and changed types, though, as in case a folder/file is created to any of them,
// there will be two entries, one for the "added" and another for the "changed".
type fsDiffJSON struct {
	Changed []string `json:"changed"`
	Deleted []string `json:"deleted"`
	Added   []string `json:"added"` // Will not be checked, but let's keep it just in case.
}

type FsDiff struct {
	result         int
	ClientHolder   clientsholder.Command
	DeletedFolders []string
	ChangedFolders []string
	Error          error
}

type FsDiffFuncs interface {
	RunTest(ctx clientsholder.Context, containerUID string)
	GetResults() int
}

func NewFsDiffTester(client clientsholder.Command) *FsDiff {
	return &FsDiff{
		ClientHolder: client,
		result:       testhelper.ERROR,
	}
}

//nolint:funlen
func (f *FsDiff) RunTest(ctx clientsholder.Context, containerUID string) {
	output, outerr, err := f.ClientHolder.ExecCommandContainer(ctx, fmt.Sprintf("chroot /host podman diff --format json %s", containerUID))
	if err != nil {
		f.Error = fmt.Errorf("can't execute command on container: %w", err)
		f.result = testhelper.ERROR
		return
	}
	if outerr != "" {
		f.Error = fmt.Errorf("stderr log received when running fsdiff test: %s", outerr)
		f.result = testhelper.ERROR
		return
	}

	// see if there's a match in the output
	logrus.Traceln("the output is ", output)

	diff := fsDiffJSON{}
	err = json.Unmarshal([]byte(output), &diff)
	if err != nil {
		f.Error = fmt.Errorf("failed to unmarshall podman diff's json output: %s, err: %w", output, err)
		f.result = testhelper.ERROR
		return
	}

	// Check for deleted folders.
	for _, deletedFolder := range diff.Deleted {
		if _, exist := targetFolders[deletedFolder]; exist {
			logrus.Tracef("Container's folder %s has been deleted.", deletedFolder)
			f.DeletedFolders = append(f.DeletedFolders, deletedFolder)
		}
	}

	// Check for changed folders.
	for _, changedFolder := range diff.Changed {
		if _, exist := targetFolders[changedFolder]; exist {
			logrus.Tracef("Container's folder %s is changed.", changedFolder)
			f.ChangedFolders = append(f.ChangedFolders, changedFolder)
		}
	}

	if len(f.ChangedFolders) != 0 || len(f.DeletedFolders) != 0 {
		f.result = testhelper.FAILURE
	} else {
		f.result = testhelper.SUCCESS
	}
}

func (f *FsDiff) GetResults() int {
	return f.result
}
