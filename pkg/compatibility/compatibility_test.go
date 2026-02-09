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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDetermineOCPStatus(t *testing.T) {
	testCases := []struct {
		testDate       time.Time
		testVersion    string
		expectedOutput string
	}{
		{ // Test Case #1 - End of life
			testDate:       time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC),
			testVersion:    "4.6",
			expectedOutput: OCPStatusEOL, // 4.6 expires on 10/18/2021
		},
		{ // Test Case #2 - Maintenance Mode
			testDate:       time.Date(2021, 10, 17, 0, 0, 0, 0, time.UTC),
			testVersion:    "4.6",
			expectedOutput: OCPStatusMS, // 4.6 expires on 10/18/2021
		},
		{ // Test Case #3 - Invalid Version
			testVersion:    "1.3",
			expectedOutput: OCPStatusUnknown, // Version 1.3 is not valid, so nothing to check
		},
		{ // Test Case #4 - Maintenance on day of
			testDate:       time.Date(2022, 1, 27, 0, 0, 0, 0, time.UTC),
			testVersion:    "4.8",
			expectedOutput: OCPStatusMS, // 4.8 enters maintenance on 1/27/2022
		},
		{ // Test Case #5 - Maintenance in window
			testDate:       time.Date(2022, 1, 28, 0, 0, 0, 0, time.UTC),
			testVersion:    "4.8",
			expectedOutput: OCPStatusMS, // 4.8 enters maintenance on 1/27/2022
		},
		{ // Test Case #6 - GA on day of
			testDate:       time.Date(2021, 7, 27, 0, 0, 0, 0, time.UTC),
			testVersion:    "4.8",
			expectedOutput: OCPStatusGA, // 4.8 enters GA on 1/27/2021
		},
		{ // Test Case #7 - Post GA, not yet in maintenance
			testDate:       time.Date(2022, 1, 26, 0, 0, 0, 0, time.UTC),
			testVersion:    "4.8",
			expectedOutput: OCPStatusGA, // 4.8 enters GA on 1/27/2022
		},
		{ // Test Case #8 - Not in maintenance window (yet)
			testDate:       time.Date(2022, 1, 26, 0, 0, 0, 0, time.UTC),
			testVersion:    "4.8",
			expectedOutput: OCPStatusGA, // 4.8 enters maintenance on 1/27/2022
		},
		{ // Test Case #9 - Not yet in GA
			testDate:       time.Date(2021, 1, 26, 0, 0, 0, 0, time.UTC),
			testVersion:    "4.8",
			expectedOutput: OCPStatusPreGA, // 4.8 enters maintenance on 1/27/2022
		},
		{ // Test Case #10 - Extended version number x.y.z
			testDate:       time.Date(2021, 1, 26, 0, 0, 0, 0, time.UTC),
			testVersion:    "4.8.2",
			expectedOutput: OCPStatusPreGA, // 4.8 enters maintenance on 1/27/2022
		},
		{ // Test Case #11 - 4.10 does not have a Maintenance date yet
			testDate:       time.Date(2022, 6, 7, 0, 0, 0, 0, time.UTC),
			testVersion:    "4.10.12",
			expectedOutput: OCPStatusGA, // 4.8 enters maintenance on 1/27/2022
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, DetermineOCPStatus(tc.testVersion, tc.testDate))
	}
}

func TestIsRHELCompatible(t *testing.T) {
	testCases := []struct {
		testMachineVersion string
		testOCPVersion     string
		expectedOutput     bool
	}{
		{ // Test Case #1 - OCP 4.10 only accepts RHEL 8.4, fail
			testOCPVersion:     "4.10",
			testMachineVersion: "7.9",
			expectedOutput:     false,
		},
		{ // Test Case #2 - OCP 4.10 only accepts RHEL 8.4, pass
			testOCPVersion:     "4.10",
			testMachineVersion: "8.4",
			expectedOutput:     true,
		},
		{ // Test Case #3 - OCP 4.10 accepts RHEL 8.4 and 8.5, pass
			testOCPVersion:     "4.10",
			testMachineVersion: "8.5",
			expectedOutput:     true,
		},
		{ // Test Case #4 - OCP 4.8 accepts RHEL >= 7.9, pass
			testOCPVersion:     "4.8",
			testMachineVersion: "8.5",
			expectedOutput:     true,
		},
		{ // Test Case #5 - OCP 4.8 accepts RHEL >= 7.9, pass
			testOCPVersion:     "4.8",
			testMachineVersion: "7.9",
			expectedOutput:     true,
		},
		{ // Test Case #6 - OCP 4.8 accepts RHEL >= 7.9, fail
			testOCPVersion:     "4.8",
			testMachineVersion: "7.8",
			expectedOutput:     false,
		},
		{ // Test Case #7 - OCP 4.8 accepts RHEL >= 7.9, pass
			testOCPVersion:     "4.8",
			testMachineVersion: "8.4",
			expectedOutput:     true,
		},
		{ // Test Case #8 - OCP version empty, fail
			testOCPVersion:     "",
			testMachineVersion: "7.8",
			expectedOutput:     false,
		},
		{ // Test Case #9 - machine version empty, fail
			testOCPVersion:     "4.8",
			testMachineVersion: "",
			expectedOutput:     false,
		},
		{ // Test Case #10 - OCP 4.9 accepts RHEL 7.9, pass
			testOCPVersion:     "4.9",
			testMachineVersion: "7.9",
			expectedOutput:     true,
		},
		{ // Test Case #11 - OCP 4.9 accepts RHEL 8.4, pass
			testOCPVersion:     "4.9",
			testMachineVersion: "8.4",
			expectedOutput:     true,
		},
		{ // Test Case #12 - OCP 4.16 accepts RHEL 8.8 and 9.2, fail with 9.4
			testOCPVersion:     "4.16",
			testMachineVersion: "9.4",
			expectedOutput:     false,
		},
		{ // Test Case #13 - OCP 4.16 accepts RHEL 8.8 and 9.2, pass with 9.2
			testOCPVersion:     "4.16",
			testMachineVersion: "9.2",
			expectedOutput:     true,
		},
		{ // Test Case #14 - OCP 4.16 does not accept RHEL 8.4, fail
			testOCPVersion:     "4.16",
			testMachineVersion: "8.4",
			expectedOutput:     false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, IsRHELCompatible(tc.testMachineVersion, tc.testOCPVersion))
	}
}

func TestIsRHCOSCompatible(t *testing.T) {
	testCases := []struct {
		testMachineVersion string
		testOCPVersion     string
		expectedOutput     bool
	}{
		{ // Test Case #1 - OCP 4.10 only accepts RHCOS version 4.10, pass
			testOCPVersion:     "4.10",
			testMachineVersion: "4.10",
			expectedOutput:     true,
		},
		{ // Test Case #2 - OCP 4.10 only accepts RHCOS version 4.10, fail
			testOCPVersion:     "4.10",
			testMachineVersion: "4.9",
			expectedOutput:     false,
		},
		{ // Test Case #3 - OCP 4.7 accepts anything 4.7+, pass
			testOCPVersion:     "4.7",
			testMachineVersion: "4.9",
			expectedOutput:     true,
		},
		{ // Test Case #4 - OCP version empty, fail
			testOCPVersion:     "",
			testMachineVersion: "7.8",
			expectedOutput:     false,
		},
		{ // Test Case #5 - machine version empty, fail
			testOCPVersion:     "4.8",
			testMachineVersion: "",
			expectedOutput:     false,
		},
		{ // Test Case #6 - OCP 4.13.0-rc.2 accepts RHCOS version 4.13.0-rc.2, pass
			testOCPVersion:     "4.13.0-rc.2",
			testMachineVersion: "4.13.0-rc.2",
			expectedOutput:     true,
		},
		{
			testOCPVersion:     "4.16.0-rc.1",
			testMachineVersion: "4.16.0-rc.1",
			expectedOutput:     true,
		},
		{ // Test Case #8 - OCP 4.20.2 accepts RHCOS version 4.20.2, pass
			testOCPVersion:     "4.20.2",
			testMachineVersion: "4.20.2",
			expectedOutput:     true,
		},
		{ // Test Case #9 - OCP 4.20.2 accepts RHCOS version 4.20.0, pass
			testOCPVersion:     "4.20.2",
			testMachineVersion: "4.20.0",
			expectedOutput:     true,
		},
		{ // Test Case #10 - OCP 4.20.0 accepts RHCOS version 4.20.2, pass
			testOCPVersion:     "4.20.0",
			testMachineVersion: "4.20.2",
			expectedOutput:     true,
		},
		{ // Test Case #11 - OCP 4.21.2 accepts RHCOS version 4.21.2, pass
			testOCPVersion:     "4.21.2",
			testMachineVersion: "4.21.2",
			expectedOutput:     true,
		},
		{ // Test Case #12 - OCP 4.21.2 accepts RHCOS version 4.21.0, pass
			testOCPVersion:     "4.21.2",
			testMachineVersion: "4.21.0",
			expectedOutput:     true,
		},
		{ // Test Case #13 - OCP 4.21.0 accepts RHCOS version 4.21.2, pass
			testOCPVersion:     "4.21.0",
			testMachineVersion: "4.21.2",
			expectedOutput:     true,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedOutput, IsRHCOSCompatible(tc.testMachineVersion, tc.testOCPVersion))
	}
}
