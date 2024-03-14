package operator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidSemanticVersioning(t *testing.T) {
	testCases := []struct {
		testVersion    string
		expectedResult bool
	}{
		{
			testVersion:    "1.1.1",
			expectedResult: true,
		},
		{
			testVersion:    "1.1.2-prerelease+meta",
			expectedResult: true,
		},
		{
			testVersion:    "1.1.2+meta",
			expectedResult: true,
		},
		{
			testVersion:    "1.0.0-alpha",
			expectedResult: true,
		},
		{
			testVersion:    "1.0.0-alpha.beta",
			expectedResult: true,
		},
		{
			testVersion:    "1.0.0-alpha.1",
			expectedResult: true,
		},
		{
			testVersion:    "1.0.0-alpha.0valid",
			expectedResult: true,
		},
		{
			testVersion:    "1.0.0-rc.1+build.1",
			expectedResult: true,
		},
		{
			testVersion:    "10.2.3-DEV-SNAPSHOT",
			expectedResult: true,
		},
		{
			testVersion:    "2.0.0+build.1848",
			expectedResult: true,
		},
		{
			testVersion:    "2.0.1-alpha.1227",
			expectedResult: true,
		},
		{
			testVersion:    "1.0.0-alpha+beta",
			expectedResult: true,
		},
		{
			testVersion:    "1.2.3----RC-SNAPSHOT.12.9.1--.12+788",
			expectedResult: true,
		},

		{
			testVersion:    "1.2.3----R-S.12.9.1--.12+meta",
			expectedResult: true,
		},
		{
			testVersion:    "1.2.3.4",
			expectedResult: false,
		},
		{
			testVersion:    "1.2.3.4.5",
			expectedResult: false,
		},
		{
			testVersion:    "1.2.3+",
			expectedResult: false,
		},
		{
			testVersion:    "hello.v1.1.1",
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedResult, isValidSemanticVersion(tc.testVersion))
	}
}
