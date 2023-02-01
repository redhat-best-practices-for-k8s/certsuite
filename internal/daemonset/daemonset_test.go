package daemonset

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildImageWithVersion(t *testing.T) {
	testCases := []struct {
		repoVar         string
		supportImageVar string
		expectedOutput  string
	}{
		{
			repoVar:         "test1",
			supportImageVar: "image1",
			expectedOutput:  "test1/image1",
		},
		{
			repoVar:         "",
			supportImageVar: "",
			expectedOutput:  "quay.io/testnetworkfunction/debug-partner:latest",
		},
	}

	defer func() {
		os.Unsetenv("TNF_PARTNER_REPO")
		os.Unsetenv("SUPPORT_IMAGE")
	}()

	for _, tc := range testCases {
		os.Setenv("TNF_PARTNER_REPO", tc.repoVar)
		os.Setenv("SUPPORT_IMAGE", tc.supportImageVar)
		assert.Equal(t, tc.expectedOutput, buildImageWithVersion())
	}
}
