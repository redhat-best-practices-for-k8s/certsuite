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
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/stringhelper"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
)

const (
	partnerPodmanFolder      = "/root/podman"
	tmpMountDestFolder       = "/tmp/tnf-podman"
	errorCode125RetrySeconds = 15
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

// fsDiffJSON represents the JSON output of `podman diff --format json`.
// It contains slices of file paths for each event type: added, changed, and deleted.
//
// The struct is used to unmarshal the command’s output into Go values,
// allowing callers to inspect which files or directories were affected.
// Only the Added, Changed, and Deleted fields are populated; other
// event types may be ignored.
type fsDiffJSON struct {
	Changed []string `json:"changed"`
	Deleted []string `json:"deleted"`
	Added   []string `json:"added"` // Will not be checked, but let's keep it just in case.
}

// FsDiff holds state and configuration for a file system difference test.
//
// It tracks changed and deleted folders, the result code, any error encountered,
// and whether custom Podman should be used. The struct is initialized by NewFsDiffTester
// with references to a checks database entry, command client, and context. RunTest
// performs the comparison using podman diff and populates the fields accordingly.
// GetResults returns the exit status of the diff operation.
type FsDiff struct {
	check           *checksdb.Check
	result          int
	clientHolder    clientsholder.Command
	ctxt            clientsholder.Context
	useCustomPodman bool

	DeletedFolders []string
	ChangedFolders []string
	Error          error
}

// FsDiffFuncs defines the contract for running filesystem diff tests.
//
// The interface provides two methods: RunTest, which executes a filesystem
// comparison test given a set of inputs, and GetResults, which retrieves
// the outcomes of those tests as structured data. Implementations are
// expected to handle any necessary setup, execution, and result aggregation
// for filesystem differences within the cnffsdiff package.
type FsDiffFuncs interface {
	RunTest(ctx clientsholder.Context, containerUID string)
	GetResults() int
}

// NewFsDiffTester creates a new FsDiff instance for comparing filesystem differences.
//
// It takes a check definition, a command interface, a context holder, and a string identifier.
// The function initializes the necessary temporary directories and decides whether to use a custom podman setup,
// then returns an initialized *FsDiff ready for running diff tests.
func NewFsDiffTester(check *checksdb.Check, client clientsholder.Command, ctxt clientsholder.Context, ocpVersion string) *FsDiff {
	useCustomPodman := shouldUseCustomPodman(check, ocpVersion)
	check.LogDebug("Using custom podman: %v.", useCustomPodman)

	return &FsDiff{
		check:           check,
		clientHolder:    client,
		ctxt:            ctxt,
		result:          testhelper.ERROR,
		useCustomPodman: useCustomPodman,
	}
}

// shouldUseCustomPodman determines whether to use the custom podman binary or the preinstalled one.
//
// It examines the OCP version from the provided checksdb.Check and the podman path string.
// If the major version is less than 4, it returns true to indicate that the custom podman
// should be used. For versions 4.13.0 and above, it returns false as the preinstalled podman
// is sufficient. The function parses the version using NewVersion and logs any parsing errors.
// It returns a boolean indicating whether the custom podman binary should be employed.
func shouldUseCustomPodman(check *checksdb.Check, ocpVersion string) bool {
	const (
		ocpForPreinstalledPodmanMajor = 4
		ocpForPreinstalledPodmanMinor = 13
	)

	version, err := semver.NewVersion(ocpVersion)
	if err != nil {
		check.LogError("Failed to parse Openshift version %q. Using preinstalled podman.", ocpVersion)
		// Use podman preinstalled in nodes as failover.
		return false
	}

	// Major versions > 4, use podman preinstalled in nodes.
	if version.Major() > ocpForPreinstalledPodmanMajor {
		return false
	}

	if version.Major() == ocpForPreinstalledPodmanMajor {
		return version.Minor() < ocpForPreinstalledPodmanMinor
	}

	// For older versions (< 3.), use podman preinstalled in nodes.
	return false
}

// intersectTargetFolders filters a slice of folder paths, returning only those that exist in the predefined target folders.
//
// It iterates over each input path and checks whether it is present in the package-level
// targetFolders slice using StringInSlice. Paths not found trigger a warning via LogWarn
// and are omitted from the result. The function returns a new slice containing only the
// valid, intersecting paths.
func (f *FsDiff) intersectTargetFolders(src []string) []string {
	var dst []string
	for _, folder := range src {
		if stringhelper.StringInSlice(targetFolders, folder, false) {
			f.check.LogWarn("Container's folder %q is altered.", folder)
			dst = append(dst, folder)
		}
	}
	return dst
}

// runPodmanDiff executes the Podman diff command inside a container and returns its output.
//
// It receives a container identifier as input, constructs a
// Podman diff command that compares the mounted filesystem with the
// target directories, runs it using ExecCommandContainer, and captures
// the resulting diff string. If the command fails, an error is returned
// with context about the failed operation. The function returns the
// diff output as a string along with any error encountered during
// execution.
func (f *FsDiff) runPodmanDiff(containerUID string) (string, error) {
	podmanPath := "podman"
	if f.useCustomPodman {
		podmanPath = fmt.Sprintf("%s/podman", tmpMountDestFolder)
	}

	output, outerr, err := f.clientHolder.ExecCommandContainer(f.ctxt, fmt.Sprintf("chroot /host %s diff --format json %s", podmanPath, containerUID))
	if err != nil {
		return "", fmt.Errorf("can not execute command on container: %w", err)
	}
	if outerr != "" {
		return "", fmt.Errorf("stderr log received when running fsdiff test: %s", outerr)
	}
	return output, nil
}

// RunTest executes the file system diff test for a given target folder.
//
// It installs a custom podman instance, mounts the node's temporary
// directory into it, runs a podman diff against the specified target,
// and then unmounts and cleans up. The function logs progress,
// handles errors by logging warnings, and sleeps briefly between steps.
// It returns an error if any step fails or nil on success.
func (f *FsDiff) RunTest(containerUID string) {
	if f.useCustomPodman {
		err := f.installCustomPodman()
		if err != nil {
			f.Error = err
			f.result = testhelper.ERROR
			return
		}

		defer f.unmountCustomPodman()
	}

	f.check.LogInfo("Running \"podman diff\" for container id %s", containerUID)

	var output string
	var err error
	for i := range [5]int{} {
		output, err = f.runPodmanDiff(containerUID)
		if err == nil {
			break
		}
		// Retry if we get a podman error code 125, which is a known issue where the container/pod
		// has possibly gone missing or is in CrashLoopBackOff state. Adding a retry here to help
		// smooth out the test results.
		if strings.Contains(err.Error(), "command terminated with exit code 125") {
			f.check.LogWarn("Retrying \"podman diff\" due to error code 125 (attempt %d/5)", i+1)
			time.Sleep(errorCode125RetrySeconds * time.Second)
			continue
		}
		break
	}

	if err != nil {
		f.Error = err
		f.result = testhelper.ERROR
		return
	}

	diff := fsDiffJSON{}
	err = json.Unmarshal([]byte(output), &diff)
	if err != nil {
		f.Error = fmt.Errorf("failed to unmarshall podman diff's json output: %s, err: %w", output, err)
		f.result = testhelper.ERROR
		return
	}
	f.DeletedFolders = f.intersectTargetFolders(diff.Deleted)
	f.ChangedFolders = f.intersectTargetFolders(diff.Changed)
	if len(f.ChangedFolders) != 0 || len(f.DeletedFolders) != 0 {
		f.check.LogDebug("Deleted folders found in Podman diff: %s", f.DeletedFolders)
		f.check.LogDebug("Changed folders found in Podman diff: %s", f.ChangedFolders)
		f.result = testhelper.FAILURE
	} else {
		f.result = testhelper.SUCCESS
	}
}

// GetResults returns the number of differences found during the comparison.
//
// It examines all stored difference records in the FsDiff instance and
// counts how many discrepancies were detected between source and target
// filesystems. The result is returned as an integer value.
func (f *FsDiff) GetResults() int {
	return f.result
}

// execCommandContainer runs a command inside the probe pod of the container under test.
// It returns an error if any output is produced on stdout or stderr, concatenating the provided
// error string with that output and the underlying execution error.
//
// The function takes two string arguments: the first is a descriptive error message used when
// reporting failures, and the second is the command to execute. It executes the command in
// the probe pod context via ExecCommandContainer. If the command writes anything to stdout
// or stderr, that output is considered a failure and included in the returned error.
// The function returns nil only when the command exits successfully with no output.
func (f *FsDiff) execCommandContainer(cmd, errorStr string) error {
	output, outerr, err := f.clientHolder.ExecCommandContainer(f.ctxt, cmd)
	if err != nil || output != "" || outerr != "" {
		return errors.New(errorStr + fmt.Sprintf(" Stderr: %s, Stdout: %s, Err: %v", output, outerr, err))
	}

	return nil
}

// createNodeFolder creates a temporary directory structure for the node side of a filesystem comparison.
//
// It prepares a mount point by invoking an execCommandContainer call to create the required folder hierarchy.
// The function returns an error if any step in setting up the node temporary mount fails.
func (f *FsDiff) createNodeFolder() error {
	return f.execCommandContainer(fmt.Sprintf("mkdir %s", nodeTmpMountFolder),
		fmt.Sprintf("failed or unexpected output when creating folder %s.", nodeTmpMountFolder))
}

// deleteNodeFolder removes the temporary mount directory used during
// a file system diff operation.
//
// It executes a container command to remove the nodeTmpMountFolder from
// the filesystem, ensuring that any temporary data created during the
// diff process is cleaned up before the test completes.
// The function returns an error if the removal command fails.
func (f *FsDiff) deleteNodeFolder() error {
	return f.execCommandContainer(fmt.Sprintf("rmdir %s", nodeTmpMountFolder),
		fmt.Sprintf("failed or unexpected output when deleting folder %s.", nodeTmpMountFolder))
}

// mountProbePodmanFolder checks that the podman folder is correctly mounted in the container.
//
// It executes a command inside the target container to verify that the
// expected podman directory exists and is accessible, returning an error if
// the check fails. The function performs no arguments and returns only an
// error value indicating success or failure of the probe.
func (f *FsDiff) mountProbePodmanFolder() error {
	return f.execCommandContainer(fmt.Sprintf("mount --bind %s %s", partnerPodmanFolder, nodeTmpMountFolder),
		fmt.Sprintf("failed or unexpected output when mounting %s into %s.", partnerPodmanFolder, nodeTmpMountFolder))
}

// unmountProbePodmanFolder removes the probe mount for a Podman folder during diff cleanup.
//
// It executes the container command to unmount the temporary probe directory
// that was mounted into the target pod. The function returns an error if the
// unmount operation fails, otherwise it returns nil. This is used as part of
// the FsDiff teardown process to ensure no leftover mounts remain after a test.
func (f *FsDiff) unmountProbePodmanFolder() error {
	return f.execCommandContainer(fmt.Sprintf("umount %s", nodeTmpMountFolder),
		fmt.Sprintf("failed or unexpected output when unmounting %s.", nodeTmpMountFolder))
}

// installCustomPodman installs the custom podman folder needed for FsDiff operations.
//
// It creates a temporary node folder, mounts the podman probe directory into it,
// and cleans up any previous installations. On success it returns nil; on failure
// it returns an error describing what went wrong.
func (f *FsDiff) installCustomPodman() error {
	// We need to create the destination folder first.
	f.check.LogInfo("Creating temp folder %s", nodeTmpMountFolder)
	if err := f.createNodeFolder(); err != nil {
		return err
	}

	// Mount podman from partner probe pod into /host/tmp/...
	f.check.LogInfo("Mounting %s into %s", partnerPodmanFolder, nodeTmpMountFolder)
	if mountErr := f.mountProbePodmanFolder(); mountErr != nil {
		// We need to delete the temp folder previously created as mount point.
		if deleteErr := f.deleteNodeFolder(); deleteErr != nil {
			return fmt.Errorf("failed to mount folder %s: %s, failed to delete %s: %s",
				partnerPodmanFolder, mountErr, nodeTmpMountFolder, deleteErr)
		}

		return mountErr
	}

	return nil
}

// unmountCustomPodman cleans up podman mount artifacts for the current FsDiff instance.
//
// It logs the start of the unmount process, attempts to unmount the probe podman folder,
// logs completion of that step, and finally removes the node temporary mount directory.
// No parameters are accepted and it does not return a value.
func (f *FsDiff) unmountCustomPodman() {
	// Unmount podman folder from host.
	f.check.LogInfo("Unmounting folder %s", nodeTmpMountFolder)
	if err := f.unmountProbePodmanFolder(); err != nil {
		// Here, there's no point on trying to remove the temp folder used as mount point, as
		// that probably will not work either.
		f.Error = err
		f.result = testhelper.ERROR
		return
	}

	f.check.LogInfo("Deleting folder %s", nodeTmpMountFolder)
	if err := f.deleteNodeFolder(); err != nil {
		f.Error = err
		f.result = testhelper.ERROR
	}
}
