// Copyright (C) 2023-2026 Red Hat, Inc.
package claim

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		fileContent    string
		createFile     bool
		expectedErrMsg string
	}{
		{
			name:       "valid claim JSON",
			createFile: true,
			fileContent: `{
				"claim": {
					"configurations": {
						"Config": null,
						"AbnormalEvents": [],
						"testOperators": []
					},
					"nodes": {
						"nodeSummary": null,
						"cniPlugins": null,
						"nodesHwInfo": null,
						"csiDriver": null
					},
					"results": {},
					"versions": {
						"claimFormat": "v0.5.0",
						"certSuite": ""
					}
				}
			}`,
		},
		{
			name:           "invalid JSON",
			createFile:     true,
			fileContent:    `{not valid json`,
			expectedErrMsg: "failed to unmarshal file",
		},
		{
			name:           "file not found",
			createFile:     false,
			expectedErrMsg: "failure reading file",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var filePath string
			if tc.createFile {
				tmpFile, err := os.CreateTemp(t.TempDir(), "claim-*.json")
				require.NoError(t, err)
				_, err = tmpFile.WriteString(tc.fileContent)
				require.NoError(t, err)
				tmpFile.Close()
				filePath = tmpFile.Name()
			} else {
				filePath = filepath.Join(t.TempDir(), "nonexistent.json")
			}

			schema, err := Parse(filePath)
			if tc.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
				assert.Nil(t, schema)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, schema)
			}
		})
	}
}

func TestIsClaimFormatVersionSupported(t *testing.T) {
	testCases := []struct {
		claimFormatVersion string
		expectedError      string
	}{
		// Invalid version strings
		{
			claimFormatVersion: "",
			expectedError:      `claim file version "" is not valid: invalid semantic version`,
		},
		{
			claimFormatVersion: "v0.v0.2",
			expectedError:      `claim file version "v0.v0.2" is not valid: invalid semantic version`,
		},
		{
			claimFormatVersion: "v0.0.0",
			expectedError:      "claim format version v0.0.0 is not supported. Supported version is v0.5.0",
		},
		{
			claimFormatVersion: "v0.0.1",
			expectedError:      "claim format version v0.0.1 is not supported. Supported version is v0.5.0",
		},
		{
			claimFormatVersion: "v0.5.0",
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
