package catalogsource

import (
	"testing"

	olmpkgv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"github.com/stretchr/testify/assert"
)

func TestSkipPMBasedOnChannel(t *testing.T) {
	testCases := []struct {
		testChannels []olmpkgv1.PackageChannel
		testCSVName  string
		expected     bool
	}{
		{ // Test Case #1 - Do not skip package manifest based on channel entry
			testChannels: []olmpkgv1.PackageChannel{
				{
					CurrentCSV: "test-csv.v1.0.0",
					Entries: []olmpkgv1.ChannelEntry{
						{
							Name: "test-csv.v1.0.0",
						},
					},
				},
			},
			testCSVName: "test-csv.v1.0.0",
			expected:    false,
		},
		{ // Test Case #2 - Skip package manifest based on channel entry
			testChannels: []olmpkgv1.PackageChannel{
				{
					CurrentCSV: "test-csv.v1.0.0",
					Entries: []olmpkgv1.ChannelEntry{
						{
							Name: "test-csv.v1.0.0",
						},
					},
				},
			},
			testCSVName: "test-csv.v1.0.1",
			expected:    true,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, SkipPMBasedOnChannel(tc.testChannels, tc.testCSVName))
	}
}
