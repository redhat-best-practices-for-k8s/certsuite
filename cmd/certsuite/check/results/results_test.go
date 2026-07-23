package results

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestGetTestResultsDB(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		fileContent    string
		createFile     bool
		expectedDB     map[string]string
		expectedErrMsg string
	}{
		{
			name:       "matching log lines",
			createFile: true,
			fileContent: `some preamble log line
2024-01-01T00:00:00Z [test-case-1] Recording result "PASSED"
another line
2024-01-01T00:00:01Z [test-case-2] Recording result "FAILED"
2024-01-01T00:00:02Z [test-case-3] Recording result "SKIPPED"
`,
			expectedDB: map[string]string{
				"test-case-1": "PASSED",
				"test-case-2": "FAILED",
				"test-case-3": "SKIPPED",
			},
		},
		{
			name:        "no matches",
			createFile:  true,
			fileContent: "this is a log line without any results\nanother line\n",
			expectedDB:  map[string]string{},
		},
		{
			name:        "empty file",
			createFile:  true,
			fileContent: "",
			expectedDB:  map[string]string{},
		},
		{
			name:           "file not found",
			createFile:     false,
			expectedErrMsg: "could not open file",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var filePath string
			if tc.createFile {
				tmpFile, err := os.CreateTemp(t.TempDir(), "log-*.txt")
				require.NoError(t, err)
				_, err = tmpFile.WriteString(tc.fileContent)
				require.NoError(t, err)
				tmpFile.Close()
				filePath = tmpFile.Name()
			} else {
				filePath = filepath.Join(t.TempDir(), "nonexistent.log")
			}

			db, err := getTestResultsDB(filePath)
			if tc.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedDB, db)
			}
		})
	}
}

func TestGetExpectedTestResults(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		fileContent    string
		createFile     bool
		expectedResult map[string]string
		expectedErrMsg string
	}{
		{
			name:       "valid YAML with pass fail skip",
			createFile: true,
			fileContent: `testCases:
  pass:
    - test-a
    - test-b
  fail:
    - test-c
  skip:
    - test-d
`,
			expectedResult: map[string]string{
				"test-a": "PASSED",
				"test-b": "PASSED",
				"test-c": "FAILED",
				"test-d": "SKIPPED",
			},
		},
		{
			name:       "empty lists",
			createFile: true,
			fileContent: `testCases:
  pass: []
  fail: []
  skip: []
`,
			expectedResult: map[string]string{},
		},
		{
			name:           "file not found",
			createFile:     false,
			expectedErrMsg: "could not open template file",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var filePath string
			if tc.createFile {
				tmpFile, err := os.CreateTemp(t.TempDir(), "template-*.yaml")
				require.NoError(t, err)
				_, err = tmpFile.WriteString(tc.fileContent)
				require.NoError(t, err)
				tmpFile.Close()
				filePath = tmpFile.Name()
			} else {
				filePath = filepath.Join(t.TempDir(), "nonexistent.yaml")
			}

			result, err := getExpectedTestResults(filePath)
			if tc.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestGenerateTemplateFile(t *testing.T) {
	testCases := []struct {
		name      string
		resultsDB map[string]string
	}{
		{
			name: "map with pass fail skip results",
			resultsDB: map[string]string{
				"test-pass": "PASSED",
				"test-fail": "FAILED",
				"test-skip": "SKIPPED",
			},
		},
		{
			name:      "empty map",
			resultsDB: map[string]string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			origDir, err := os.Getwd()
			require.NoError(t, err)
			tmpDir := t.TempDir()
			err = os.Chdir(tmpDir)
			require.NoError(t, err)
			defer func() {
				require.NoError(t, os.Chdir(origDir))
			}()

			err = generateTemplateFile(tc.resultsDB)
			require.NoError(t, err)

			outputPath := filepath.Join(tmpDir, TestResultsTemplateFileName)
			data, err := os.ReadFile(outputPath)
			require.NoError(t, err)

			var parsed TestResults
			err = yaml.Unmarshal(data, &parsed)
			require.NoError(t, err)

			passCount := 0
			failCount := 0
			skipCount := 0
			for _, v := range tc.resultsDB {
				switch v {
				case resultPass:
					passCount++
				case resultFail:
					failCount++
				case resultSkip:
					skipCount++
				}
			}
			assert.Len(t, parsed.Pass, passCount)
			assert.Len(t, parsed.Fail, failCount)
			assert.Len(t, parsed.Skip, skipCount)
		})
	}
}
