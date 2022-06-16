// Copyright (C) 2022 Red Hat, Inc.
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
	"strings"
	"time"
)

/* Notes for this package
* Refer to this document for more information about OCP compatibility: https://access.redhat.com/support/policy/updates/openshift

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
}

var (
	ocpLifeCycleDates = map[string]VersionInfo{
		// TODO: Adjust all of these periodically to make sure they are up to date with the lifecycle
		// update documentation.

		// Full Support
		"4.10": {
			GADate:  time.Date(2022, 3, 10, 0, 0, 0, 0, time.UTC), // March 10, 2022
			FSEDate: time.Date(2022, 9, 10, 0, 0, 0, 0, time.UTC), // September 10, 2022
			MSEDate: time.Date(2023, 9, 10, 0, 0, 0, 0, time.UTC), // September 10, 2023
			// Note: FSEDate (Release of 4.11 + 3 months) is currently a "guess".  Update when available.
		},
		"4.9": {
			GADate:  time.Date(2021, 10, 18, 0, 0, 0, 0, time.UTC), // October 18, 2021
			FSEDate: time.Date(2022, 6, 8, 0, 0, 0, 0, time.UTC),   // June 8, 2022
			MSEDate: time.Date(2023, 4, 18, 0, 0, 0, 0, time.UTC),  // April 18, 2023
			// Note: FSEDate (Release of 4.10 + 3 months) is currently a "guess".  Update when available.
		},

		// Maintenance Support
		"4.8": {
			GADate:  time.Date(2021, 7, 27, 0, 0, 0, 0, time.UTC), // July 27, 2021
			FSEDate: time.Date(2022, 1, 27, 0, 0, 0, 0, time.UTC), // January 27, 2022
			MSEDate: time.Date(2023, 1, 27, 0, 0, 0, 0, time.UTC), // January 27, 2023
		},
		"4.7": {
			GADate:  time.Date(2021, 2, 24, 0, 0, 0, 0, time.UTC),  // February 24, 2021
			FSEDate: time.Date(2021, 10, 27, 0, 0, 0, 0, time.UTC), // October 27, 2021
			MSEDate: time.Date(2022, 8, 24, 0, 0, 0, 0, time.UTC),  // August 24, 2022
		},

		// End of life
		"4.6": {
			GADate:  time.Date(2020, 10, 27, 0, 0, 0, 0, time.UTC), // October 27, 2020
			FSEDate: time.Date(2021, 3, 24, 0, 0, 0, 0, time.UTC),  // March 24, 2021
			MSEDate: time.Date(2021, 10, 18, 0, 0, 0, 0, time.UTC), // October 18, 2022
		},
		"4.5": {
			GADate:  time.Date(2020, 7, 13, 0, 0, 0, 0, time.UTC),  // July 13, 2020
			FSEDate: time.Date(2020, 11, 27, 0, 0, 0, 0, time.UTC), // November 27, 2020
			MSEDate: time.Date(2021, 7, 27, 0, 0, 0, 0, time.UTC),  // July 27, 2021
		},
		"4.4": {
			GADate:  time.Date(2020, 5, 5, 0, 0, 0, 0, time.UTC),  // May 5, 2020
			FSEDate: time.Date(2020, 8, 13, 0, 0, 0, 0, time.UTC), // August 13, 2020
			MSEDate: time.Date(2021, 2, 24, 0, 0, 0, 0, time.UTC), // February 24, 2021
		},
		"4.3": {
			GADate:  time.Date(2020, 1, 23, 0, 0, 0, 0, time.UTC),  // January 23, 2020
			FSEDate: time.Date(2020, 6, 5, 0, 0, 0, 0, time.UTC),   // June 5, 2020
			MSEDate: time.Date(2020, 10, 27, 0, 0, 0, 0, time.UTC), // October 27, 2020
		},
		"4.2": {
			GADate:  time.Date(2019, 10, 16, 0, 0, 0, 0, time.UTC), // October 16, 2019
			FSEDate: time.Date(2020, 2, 23, 0, 0, 0, 0, time.UTC),  // February 23, 2020
			MSEDate: time.Date(2020, 7, 13, 0, 0, 0, 0, time.UTC),  // July 13, 2020
		},
		"4.1": {
			GADate:  time.Date(2019, 6, 4, 0, 0, 0, 0, time.UTC),   // June 4, 2019
			FSEDate: time.Date(2019, 11, 16, 0, 0, 0, 0, time.UTC), // November 16, 2019
			MSEDate: time.Date(2020, 5, 5, 0, 0, 0, 0, time.UTC),   // May 5, 2020
		},
	}
)

func GetLifeCycleDates() map[string]VersionInfo {
	return ocpLifeCycleDates
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
