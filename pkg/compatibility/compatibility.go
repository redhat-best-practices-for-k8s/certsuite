// Copyright (C) 2022-2026 Red Hat, Inc.
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

package compatibility

import (
	_ "embed"
	"encoding/json"
	"slices"
	"strings"
	"sync"
	"time"

	gv "github.com/hashicorp/go-version"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/stringhelper"
)

/* Notes for this package
* Refer to these documents for more information about OCP compatibility:
*   - OCP Lifecycle Policy: https://access.redhat.com/support/policy/updates/openshift
*   - RHEL Versions in RHCOS/OCP: https://access.redhat.com/articles/6907891
*
* IMPORTANT: The RHELVersionsAccepted field refers to RHEL versions supported for WORKER NODES,
* not the RHEL version used internally by RHCOS. Worker nodes can run standard RHEL while
* control plane nodes must run RHCOS. See release notes comments below for worker node compatibility.
*
* Lifecycle dates are sourced from https://endoflife.date/api/openshift.json and auto-updated
* via the update-ocp-lifecycle GitHub Actions workflow. RHEL worker-node compatibility versions
* are manually maintained in data/rhel_compat.json.
*
* This module will help compare the running OCP version against a matrix of end of life dates.
 */

const (
	// OCP Lifecycle Statuses
	OCPStatusGA      = "generally-available"
	OCPStatusMS      = "maintenance-support"
	OCPStatusEOL     = "end-of-life"
	OCPStatusUnknown = "unknown"
	OCPStatusPreGA   = "pre-general-availability"
)

type VersionInfo struct {
	GADate  time.Time // General Availability Date
	FSEDate time.Time // Full Support Ends Date
	MSEDate time.Time // Maintenance Support Ends Date

	MinRHCOSVersion      string   // Minimum RHCOS Version supported
	RHELVersionsAccepted []string // Contains either specific versions or a minimum version eg. "7.9 or later" or "7.9 and 8.4"
}

//go:embed data/openshift_lifecycle.json
var openshiftLifecycleJSON []byte

//go:embed data/rhel_compat.json
var rhelCompatJSON []byte

// endOfLifeEntry represents a single entry from the endoflife.date API.
// The support and extendedSupport fields use json.RawMessage because the API
// returns either a date string (e.g. "2025-09-17") or a boolean (true/false).
type endOfLifeEntry struct {
	Cycle           string          `json:"cycle"`
	ReleaseDate     string          `json:"releaseDate"`
	EOL             string          `json:"eol"`
	Support         json.RawMessage `json:"support"`
	ExtendedSupport json.RawMessage `json:"extendedSupport"`
}

// parseDateOrBool attempts to parse a json.RawMessage as a date string.
// Returns the parsed time and true if successful, or zero time and false
// if the value is a boolean or unparseable.
func parseDateOrBool(raw json.RawMessage) (time.Time, bool) {
	var dateStr string
	if err := json.Unmarshal(raw, &dateStr); err == nil {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

var (
	loadOnce          sync.Once
	ocpLifeCycleDates map[string]VersionInfo
	ocpBetaVersions   []string
)

func loadData() {
	loadOnce.Do(func() {
		ocpLifeCycleDates = make(map[string]VersionInfo)

		// Parse RHEL compatibility map
		var rhelCompat map[string][]string
		if err := json.Unmarshal(rhelCompatJSON, &rhelCompat); err != nil {
			log.Error("Error parsing rhel_compat.json: %v", err)
			return
		}

		// Parse lifecycle entries
		var entries []endOfLifeEntry
		if err := json.Unmarshal(openshiftLifecycleJSON, &entries); err != nil {
			log.Error("Error parsing openshift_lifecycle.json: %v", err)
			return
		}

		for _, e := range entries {
			gaDate := parseDate(e.ReleaseDate)

			// FSEDate: "support" is a date string when full support has ended,
			// or boolean true when still in full support (no end date yet).
			fseDate, _ := parseDateOrBool(e.Support)

			// MSEDate: prefer extendedSupport date (EUS end) over standard eol.
			mseDate, hasExtended := parseDateOrBool(e.ExtendedSupport)
			if !hasExtended {
				mseDate = parseDate(e.EOL)
			}

			info := VersionInfo{
				GADate:               gaDate,
				FSEDate:              fseDate,
				MSEDate:              mseDate,
				MinRHCOSVersion:      e.Cycle,
				RHELVersionsAccepted: rhelCompat[e.Cycle],
			}
			ocpLifeCycleDates[e.Cycle] = info
			ocpBetaVersions = append(ocpBetaVersions, e.Cycle)
		}

		// Add versions from rhel_compat.json that aren't yet in the API data
		// (e.g. pre-GA versions). This ensures RHCOS beta matching works for
		// upcoming releases that have been added to rhel_compat.json but haven't
		// appeared in the endoflife.date API yet.
		for cycle, rhelVersions := range rhelCompat {
			if _, exists := ocpLifeCycleDates[cycle]; !exists {
				ocpLifeCycleDates[cycle] = VersionInfo{
					MinRHCOSVersion:      cycle,
					RHELVersionsAccepted: rhelVersions,
				}
				ocpBetaVersions = append(ocpBetaVersions, cycle)
			}
		}
	})
}

func GetLifeCycleDates() map[string]VersionInfo {
	loadData()
	return ocpLifeCycleDates
}

func BetaRHCOSVersionsFoundToMatch(machineVersion, ocpVersion string) bool {
	loadData()
	ocpVersion = FindMajorMinor(ocpVersion)
	machineVersion = FindMajorMinor(machineVersion)

	// Check if the versions exist in the beta list
	if !stringhelper.StringInSlice(ocpBetaVersions, ocpVersion, false) || !stringhelper.StringInSlice(ocpBetaVersions, machineVersion, false) {
		return false
	}

	// Check if the versions match
	return ocpVersion == machineVersion
}

func IsRHELCompatible(machineVersion, ocpVersion string) bool {
	if machineVersion == "" || ocpVersion == "" {
		return false
	}

	lifecycleInfo := GetLifeCycleDates()
	if entry, ok := lifecycleInfo[ocpVersion]; ok {
		if len(entry.RHELVersionsAccepted) == 0 {
			return false
		}

		if len(entry.RHELVersionsAccepted) >= 2 { //nolint:mnd
			// Need to be a specific major.minor version
			return slices.Contains(entry.RHELVersionsAccepted, machineVersion)
		}

		// Single version entry: accept if machine version >= the entry version
		mv, _ := gv.NewVersion(machineVersion)
		ev, _ := gv.NewVersion(entry.RHELVersionsAccepted[0])
		return mv.GreaterThanOrEqual(ev)
	}

	return false
}

func FindMajorMinor(version string) string {
	splitVersion := strings.Split(version, ".")
	return splitVersion[0] + "." + splitVersion[1]
}

func IsRHCOSCompatible(machineVersion, ocpVersion string) bool {
	if machineVersion == "" || ocpVersion == "" {
		return false
	}

	// Exception for beta versions
	if BetaRHCOSVersionsFoundToMatch(machineVersion, ocpVersion) {
		return true
	}

	// Split the incoming version on the "." and make sure we are only looking at major.minor.
	ocpVersion = FindMajorMinor(ocpVersion)

	lifecycleInfo := GetLifeCycleDates()
	if entry, ok := lifecycleInfo[ocpVersion]; ok {
		// Collect the machine version and the entry version
		mv, err := gv.NewVersion(machineVersion)
		if err != nil {
			log.Error("Error parsing machineVersion: %s err: %v", machineVersion, err)
			return false
		}
		ev, err := gv.NewVersion(entry.MinRHCOSVersion)
		if err != nil {
			log.Error("Error parsing MinRHCOSVersion: %s err: %v", entry.MinRHCOSVersion, err)
			return false
		}

		// If the machine version >= the entry version
		return mv.GreaterThanOrEqual(ev)
	}

	return false
}

func DetermineOCPStatus(version string, date time.Time) string {
	// Safeguard against empty values being passed in
	if version == "" || date.IsZero() {
		return OCPStatusUnknown
	}

	// Split the incoming version on the "." and make sure we are only looking at major.minor.
	splitVersion := strings.Split(version, ".")
	version = splitVersion[0] + "." + splitVersion[1]

	// Check if the version exists in our local map
	lifecycleDates := GetLifeCycleDates()
	if entry, ok := lifecycleDates[version]; ok {
		// Safeguard against the latest versions not having a date set for FSEDate set.
		// See the OpenShift lifecycle website link (above) for more details on this.
		if entry.FSEDate.IsZero() {
			entry.FSEDate = entry.MSEDate
		}

		// Pre-GA
		if date.Before(entry.GADate) {
			return OCPStatusPreGA
		}
		// Generally Available
		if date.Equal(entry.GADate) || date.After(entry.GADate) && date.Before(entry.FSEDate) {
			return OCPStatusGA
		}
		// Maintenance Support
		if date.Equal(entry.FSEDate) || (date.After(entry.FSEDate) && date.Before(entry.MSEDate)) {
			return OCPStatusMS
		}
		// End of Life
		if date.Equal(entry.MSEDate) || date.After(entry.MSEDate) {
			return OCPStatusEOL
		}
	}

	return OCPStatusUnknown
}

/*

Note:
You must use RHCOS machines for the control plane, and you can use either RHCOS or RHEL for compute machines.

Compatibility information gathered from the release note pages of each release:

4.11
https://docs.openshift.com/container-platform/4.11/release_notes/ocp-4-11-release-notes.html
OpenShift Container Platform 4.11 is supported on Red Hat Enterprise Linux (RHEL) 8.4 and 8.5, as well as on Red Hat Enterprise Linux CoreOS (RHCOS) 4.11.

4.10
https://docs.openshift.com/container-platform/4.10/release_notes/ocp-4-10-release-notes.html
OpenShift Container Platform 4.10 is supported on Red Hat Enterprise Linux (RHEL) 8.4 and 8.5, as well as on Red Hat Enterprise Linux CoreOS (RHCOS) 4.10.

4.9
https://docs.openshift.com/container-platform/4.9/release_notes/ocp-4-9-release-notes.html
OpenShift Container Platform 4.9 is supported on Red Hat Enterprise Linux (RHEL) 7.9 and 8.4, as well as on Red Hat Enterprise Linux CoreOS (RHCOS) 4.9.

4.8
https://docs.openshift.com/container-platform/4.8/release_notes/ocp-4-8-release-notes.html
OpenShift Container Platform 4.8 is supported on Red Hat Enterprise Linux (RHEL) 7.9 or later, as well as on Red Hat Enterprise Linux CoreOS (RHCOS) 4.8.

4.7
https://docs.openshift.com/container-platform/4.7/release_notes/ocp-4-7-release-notes.html
OpenShift Container Platform 4.7 is supported on Red Hat Enterprise Linux (RHEL) 7.9 or later, as well as Red Hat Enterprise Linux CoreOS (RHCOS) 4.7.

4.6
https://docs.openshift.com/container-platform/4.6/release_notes/ocp-4-6-release-notes.html
OpenShift Container Platform 4.6 is supported on Red Hat Enterprise Linux 7.9 or later, as well as Red Hat Enterprise Linux CoreOS (RHCOS) 4.6.

4.5
https://docs.openshift.com/container-platform/4.5/release_notes/ocp-4-5-release-notes.html
OpenShift Container Platform 4.5 is supported on RHEL 7, version 7.7 or 7.8, as well as Red Hat Enterprise Linux CoreOS (RHCOS) 4.5.

4.4
https://docs.openshift.com/container-platform/4.4/release_notes/ocp-4-4-release-notes.html
OpenShift Container Platform 4.4 is supported on Red Hat Enterprise Linux 7.6 or later, as well as Red Hat Enterprise Linux CoreOS (RHCOS) 4.4.

4.3
https://docs.openshift.com/container-platform/4.3/release_notes/ocp-4-3-release-notes.html
OpenShift Container Platform 4.3 is supported on Red Hat Enterprise Linux 7.6 or later, as well as Red Hat Enterprise Linux CoreOS 4.3.

4.2
https://docs.openshift.com/container-platform/4.2/release_notes/ocp-4-2-release-notes.html
OpenShift Container Platform 4.2 is supported on Red Hat Enterprise Linux 7.6 and later, as well as Red Hat Enterprise Linux CoreOS 4.2.

4.1
https://docs.openshift.com/container-platform/4.1/release_notes/ocp-4-1-release-notes.html
OpenShift Container Platform 4.1 is supported on Red Hat Enterprise Linux 7.6 and later, as well as Red Hat Enterprise Linux CoreOS 4.1.

*/
