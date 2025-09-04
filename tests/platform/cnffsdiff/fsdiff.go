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

// fsDiffJSON Parses podman diff JSON output into separate lists of changed, added, and deleted paths
//
// This struct holds three slices of strings that represent file or folder paths
// reported by the podman diff command. The "changed" slice contains paths
// modified in a container, "deleted" lists removed items, and "added" tracks
// new creations. Only the changed and deleted fields are used for comparison
// logic, while added is retained for completeness.
type fsDiffJSON struct {
	Changed []string `json:"changed"`
	Deleted []string `json:"deleted"`
	Added   []string `json:"added"` // Will not be checked, but let's keep it just in case.
}

// FsDiff Tracks file system differences in a container
//
// This structure stores the results of running a podman diff against a
// container, capturing any folders that have been changed or deleted from a
// predefined target list. It also holds references to the check context,
// command client, and execution context used during the test, along with flags
// for custom podman usage and an error field for failure reporting. The result
// integer indicates success, failure, or error status after the test runs.
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

// FsDiffFuncs provides file system diff functionality
//
// This interface defines two operations: one that initiates a diff test within
// a specified container context, and another that retrieves the result status
// of that test as an integer code. The RunTest method accepts execution context
// and container identifier parameters to perform the comparison, while
// GetResults returns an integer indicating success or failure of the last run.
type FsDiffFuncs interface {
	RunTest(ctx clientsholder.Context, containerUID string)
	GetResults() int
}

// NewFsDiffTester Creates a tester for filesystem differences in containers
//
// It determines whether to use a custom podman based on the OpenShift version,
// logs this decision, and initializes an FsDiff structure with the provided
// check, client holder, context, and result state. The returned object is ready
// to run tests that compare container file systems.
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

// shouldUseCustomPodman determines whether a custom podman binary should be used
//
// The function parses the OpenShift version string to decide if the
// preinstalled podman on each node is suitable. For versions below 4.13 it
// selects a custom, precompiled podman that works with older RHEL 8.x based
// clusters; for newer releases or parsing failures it defaults to the node’s
// built‑in podman. The result is returned as a boolean.
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

// FsDiff.intersectTargetFolders Filters a list of folders to those that are monitored
//
// The function iterates over the supplied slice, checking each path against a
// predefined set of target directories. If a match is found, it logs a warning
// and adds the folder to the result slice. The resulting slice contains only
// paths that belong to the monitored set.
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

// FsDiff.runPodmanDiff Runs podman diff and returns its JSON output
//
// This method constructs the path to podman, optionally using a custom binary
// if configured. It then executes a chrooted command inside the host
// environment to obtain a diff of the container’s filesystem in JSON format.
// The function captures standard output and errors, returning the output string
// or an error if execution fails.
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

// FsDiff.RunTest Executes podman diff to detect container file system changes
//
// The method runs the "podman diff" command on a specified container,
// optionally installing a custom podman binary if configured. It retries up to
// five times when encountering exit code 125 errors and parses the JSON output
// into deleted and changed folder lists. If any target folders are found
// altered or removed, the test fails; otherwise it succeeds.
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

// FsDiff.GetResults provides the current result value
//
// The method simply retrieves and returns the integer field that holds the diff
// outcome. No parameters are required, and it does not modify any state. The
// returned value reflects the number of differences detected by the FsDiff
// instance.
func (f *FsDiff) GetResults() int {
	return f.result
}

// FsDiff.execCommandContainer Executes a shell command inside the probe pod and reports any output as an error
//
// It runs the supplied command in the container associated with FsDiff,
// capturing both stdout and stderr. If the command fails or produces any
// output, it returns an error that includes the provided error string plus the
// captured outputs and underlying execution error. Otherwise, it returns nil to
// indicate success.
func (f *FsDiff) execCommandContainer(cmd, errorStr string) error {
	output, outerr, err := f.clientHolder.ExecCommandContainer(f.ctxt, cmd)
	if err != nil || output != "" || outerr != "" {
		return errors.New(errorStr + fmt.Sprintf(" Stderr: %s, Stdout: %s, Err: %v", output, outerr, err))
	}

	return nil
}

// FsDiff.createNodeFolder Creates a temporary folder on the node for mounting purposes
//
// The method runs a container command to make a directory at the path defined
// by nodeTmpMountFolder. It uses execCommandContainer to capture any output or
// errors, returning an error if the command fails or produces unexpected
// output.
func (f *FsDiff) createNodeFolder() error {
	return f.execCommandContainer(fmt.Sprintf("mkdir %s", nodeTmpMountFolder),
		fmt.Sprintf("failed or unexpected output when creating folder %s.", nodeTmpMountFolder))
}

// FsDiff.deleteNodeFolder Removes the temporary mount directory on the target node
//
// This method issues a command to delete the folder designated by the constant
// nodeTmpMountFolder using the execCommandContainer helper. It expects no
// output from the command; any stdout, stderr or execution error results in an
// informative error being returned. The function is invoked during setup and
// teardown of custom Podman mounts to clean up the temporary directory.
func (f *FsDiff) deleteNodeFolder() error {
	return f.execCommandContainer(fmt.Sprintf("rmdir %s", nodeTmpMountFolder),
		fmt.Sprintf("failed or unexpected output when deleting folder %s.", nodeTmpMountFolder))
}

// FsDiff.mountProbePodmanFolder Binds a partner pod's podman directory into the node's temporary mount point
//
// This method runs a bind‑mount command inside the container to expose the
// partner probe's podman folder at the node’s temporary location. It
// constructs the mount command with the source and destination paths, executes
// it via execCommandContainer, and returns any error from that execution. If
// the command succeeds, no value is returned.
func (f *FsDiff) mountProbePodmanFolder() error {
	return f.execCommandContainer(fmt.Sprintf("mount --bind %s %s", partnerPodmanFolder, nodeTmpMountFolder),
		fmt.Sprintf("failed or unexpected output when mounting %s into %s.", partnerPodmanFolder, nodeTmpMountFolder))
}

// FsDiff.unmountProbePodmanFolder Unmounts the probe podman mount folder from within the container
//
// The method runs a command inside the container to unmount the temporary host
// folder used for probing filesystem differences. It reports any error or
// unexpected output, propagating it back to the caller. The operation is part
// of cleaning up after tests and returns an error if the unmount fails.
func (f *FsDiff) unmountProbePodmanFolder() error {
	return f.execCommandContainer(fmt.Sprintf("umount %s", nodeTmpMountFolder),
		fmt.Sprintf("failed or unexpected output when unmounting %s.", nodeTmpMountFolder))
}

// FsDiff.installCustomPodman prepares a temporary mount point for custom podman
//
// This method creates a temporary directory, mounts the partner probe podman's
// podman binary into that directory, and cleans up if mounting fails. It logs
// each step and returns an error if any operation fails. The setup is used
// before running podman diff in tests.
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

// FsDiff.unmountCustomPodman Unmounts the temporary Podman mount directory
//
// The function logs that it is unmounting a specific folder, then attempts to
// unmount it using a helper command. If the unmount fails, it records an error
// and stops further cleanup. Finally, it deletes the now-unmounted folder,
// recording any errors encountered during deletion.
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
