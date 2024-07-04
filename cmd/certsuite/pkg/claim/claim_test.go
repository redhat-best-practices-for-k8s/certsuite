package claim

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsClaimFormatVersionSupported(t *testing.T) {
	testCases := []struct {
		claimFormatVersion string
		expectedError      string
	}{
		// Invalid version strings
		{
			claimFormatVersion: "",
			expectedError:      `claim file version "" is not valid: Invalid Semantic Version`,
		},
		{
			claimFormatVersion: "v0.v0.2",
			expectedError:      `claim file version "v0.v0.2" is not valid: Invalid Semantic Version`,
		},
		{
			claimFormatVersion: "v0.0.0",
			expectedError:      "claim format version v0.0.0 is not supported. Supported version is v0.4.0",
		},
		{
			claimFormatVersion: "v0.0.1",
			expectedError:      "claim format version v0.0.1 is not supported. Supported version is v0.4.0",
		},
		{
			claimFormatVersion: "v0.4.0",
			expectedError:      "",
		},
	}

	for _, tc := range testCases {
		err := CheckVersion(tc.claimFormatVersion)
		if err != nil {
			assert.Equal(t, tc.expectedError, err.Error())
		}
	}
}
