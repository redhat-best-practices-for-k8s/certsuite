package compare

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_claimCompareFilesfunc(t *testing.T) {
	testCases := []struct {
		name               string
		claim1Path         string
		claim2Path         string
		expectedOutputFile string
		expectedError      string
	}{
		{
			name:          "claim1 file not found",
			claim1Path:    "not_found.json",
			claim2Path:    "testdata/claim_access_control.json",
			expectedError: "failed reading claim1 file: open not_found.json: no such file or directory",
		},
		{
			name:          "claim1 is not a json file",
			claim1Path:    "testdata/invalid.json",
			claim2Path:    "testdata/claim_access_control.json",
			expectedError: "failed to unmarshal claim1 file: invalid character 'T' looking for beginning of value",
		},
		{
			name:          "claim2 file not found",
			claim1Path:    "testdata/claim_observability.json",
			claim2Path:    "not_found.json",
			expectedError: "failed reading claim2 file: open not_found.json: no such file or directory",
		},
		{
			name:          "claim2 is not a json file",
			claim1Path:    "testdata/claim_observability.json",
			claim2Path:    "testdata/invalid.json",
			expectedError: "failed to unmarshal claim2 file: invalid character 'T' looking for beginning of value",
		},
		// claim_observability.json: claim.json from a run on a 3-nodes cluster with "-l observability".
		// claim_access_control.json: real claim.json on a 3-nodes cluster with "-l access-control".
		// In both files, the claim.configuration section has been removed. Also, some manual changes have been
		// made into nodes clus0-0 and clus0-1 in order to make the cni, csi hardware sections slightly different
		// so the differences can be seen in the report.
		{
			name:               "compare two valid claim files",
			claim1Path:         "testdata/claim_observability.json",
			claim2Path:         "testdata/claim_access_control.json",
			expectedOutputFile: "testdata/diff1.txt",
		},
		{
			name:               "compare two valid claim files in reverse order",
			claim1Path:         "testdata/claim_access_control.json",
			claim2Path:         "testdata/claim_observability.json",
			expectedOutputFile: "testdata/diff1_reverse.txt",
		},
		{
			name:               "claim files have the same content",
			claim1Path:         "testdata/claim_observability.json",
			claim2Path:         "testdata/claim_observability.json",
			expectedOutputFile: "testdata/diff2_same_claims.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create pipe so we can read the stdout output from the test function.
			originalStdout := os.Stdout
			r, w, err := os.Pipe()
			assert.Nil(t, err)

			os.Stdout = w
			// Run function under test.
			err = claimCompareFilesfunc(tc.claim1Path, tc.claim2Path)
			// Close write pipe. Needed so the io.ReadAll can detect the EOF.
			w.Close()

			// Read expected output from test file.
			var expectedOutput []byte
			var testErr error
			if tc.expectedOutputFile != "" {
				expectedOutput, testErr = os.ReadFile(tc.expectedOutputFile)
				assert.Nil(t, testErr)
			}

			if err != nil {
				assert.Equal(t, tc.expectedError, err.Error())
			} else {
				// Read the output from the pipe.
				out, _ := io.ReadAll(r)
				t.Logf("%s", string(out))
				assert.Equal(t, string(expectedOutput), string(out))
			}

			// Restore original stdout.
			os.Stdout = originalStdout
		})
	}
}
