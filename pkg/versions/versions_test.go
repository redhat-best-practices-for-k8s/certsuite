package versions

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
		assert.Equal(t, tc.expectedResult, IsValidSemanticVersion(tc.testVersion))
	}
}

func TestIsValidK8sVersioning(t *testing.T) {
	testCases := []struct {
		testVersion    string
		expectedResult bool
	}{
		{
			testVersion:    "v1",
			expectedResult: true,
		},
		{
			testVersion:    "v2",
			expectedResult: true,
		},
		{
			testVersion:    "v205",
			expectedResult: true,
		},
		{
			testVersion:    "v1alpha1",
			expectedResult: true,
		},
		{
			testVersion:    "v205alpha205",
			expectedResult: true,
		},
		{
			testVersion:    "v1alpha",
			expectedResult: false,
		},
		{
			testVersion:    "v1aLpha1",
			expectedResult: false,
		},
		{
			testVersion:    "v0",
			expectedResult: false,
		}, {
			testVersion:    "v1v1",
			expectedResult: false,
		},
		{
			testVersion:    "v1alpha1alpha1",
			expectedResult: true,
		},
		{
			testVersion:    "v1alpha1alpha1beta2",
			expectedResult: false,
		},
		{
			testVersion:    "v3beta11",
			expectedResult: true,
		},
		{
			testVersion:    "v3gamma101",
			expectedResult: false,
		},
		{
			testVersion:    "1.1.1",
			expectedResult: false,
		},
		{
			testVersion:    "unstable",
			expectedResult: false,
		},
		{
			testVersion:    "v",
			expectedResult: false,
		},
		{
			testVersion:    "w",
			expectedResult: false,
		},
		{
			testVersion:    "",
			expectedResult: false,
		},

		{
			testVersion:    "1alpha1",
			expectedResult: false,
		},
		{
			testVersion:    "1",
			expectedResult: false,
		},
		{
			testVersion:    "",
			expectedResult: false,
		},
		{
			testVersion:    "hello.v1alpha1",
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedResult, IsValidK8sVersion(tc.testVersion))
	}
}

func TestGitVersion(t *testing.T) {
	GitCommit = "123456"
	GitRelease = "v1.0.0"
	GitPreviousRelease = "v0.9.0"
	ClaimFormatVersion = "1.0.0"
	assert.Equal(t, "v1.0.0 (123456)", GitVersion())

	GitRelease = ""
	assert.Equal(t, "Unreleased build post v0.9.0 (123456)", GitVersion())
}
