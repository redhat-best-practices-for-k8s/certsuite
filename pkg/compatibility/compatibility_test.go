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
	"encoding/json"
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

func TestParseDateOrBool(t *testing.T) {
	testCases := []struct {
		name         string
		input        json.RawMessage
		expectedTime time.Time
		expectedOk   bool
	}{
		{
			name:         "date string",
			input:        json.RawMessage(`"2025-09-17"`),
			expectedTime: time.Date(2025, 9, 17, 0, 0, 0, 0, time.UTC),
			expectedOk:   true,
		},
		{
			name:         "boolean true (still in full support)",
			input:        json.RawMessage(`true`),
			expectedTime: time.Time{},
			expectedOk:   false,
		},
		{
			name:         "boolean false (no extended support)",
			input:        json.RawMessage(`false`),
			expectedTime: time.Time{},
			expectedOk:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, ok := parseDateOrBool(tc.input)
			assert.Equal(t, tc.expectedOk, ok)
			assert.Equal(t, tc.expectedTime, result)
		})
	}
}

func TestParseLifecycleData(t *testing.T) {
	dates := GetLifeCycleDates()

	// 4.18: support="2025-09-17" (date), extendedSupport="2027-02-25" (EUS date)
	v418, ok := dates["4.18"]
	assert.True(t, ok)
	assert.Equal(t, time.Date(2025, 2, 25, 0, 0, 0, 0, time.UTC), v418.GADate)
	assert.Equal(t, time.Date(2025, 9, 17, 0, 0, 0, 0, time.UTC), v418.FSEDate)
	assert.Equal(t, time.Date(2027, 2, 25, 0, 0, 0, 0, time.UTC), v418.MSEDate) // EUS end date

	// 4.20: support=true (boolean), extendedSupport=false (boolean)
	v420, ok := dates["4.20"]
	assert.True(t, ok)
	assert.Equal(t, time.Date(2025, 10, 21, 0, 0, 0, 0, time.UTC), v420.GADate)
	assert.True(t, v420.FSEDate.IsZero(), "FSEDate should be zero when support is boolean true")
	assert.Equal(t, time.Date(2027, 4, 21, 0, 0, 0, 0, time.UTC), v420.MSEDate) // falls back to eol

	// 4.17: support="2025-05-25" (date), extendedSupport=false (no EUS)
	v417, ok := dates["4.17"]
	assert.True(t, ok)
	assert.Equal(t, time.Date(2025, 5, 25, 0, 0, 0, 0, time.UTC), v417.FSEDate)
	assert.Equal(t, time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC), v417.MSEDate) // standard eol, not EUS
}

func TestIsRHELCompatibleEmptyRHELVersions(t *testing.T) {
	// A version in the lifecycle data but not in rhel_compat.json should return false, not panic
	// We test this by checking an unknown OCP version which has no RHEL compat entry
	result := IsRHELCompatible("8.4", "99.99")
	assert.False(t, result)
}

func TestPreGAVersionFromRHELCompat(t *testing.T) {
	// 4.21 is in rhel_compat.json but not in the API data yet.
	// It should still appear in the lifecycle dates with RHEL versions populated.
	dates := GetLifeCycleDates()
	v421, ok := dates["4.21"]
	assert.True(t, ok, "4.21 should be present from rhel_compat.json")
	assert.Equal(t, "4.21", v421.MinRHCOSVersion)
	assert.Equal(t, []string{"8.10", "9.4"}, v421.RHELVersionsAccepted)

	// RHCOS beta matching should work for 4.21
	assert.True(t, BetaRHCOSVersionsFoundToMatch("4.21", "4.21"))
}
