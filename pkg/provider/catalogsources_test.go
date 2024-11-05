package provider

import (
	"testing"

	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmpkgv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetCatalogSourceBundleCount(t *testing.T) {
	generateEnv := func(channelEntries []olmpkgv1.ChannelEntry) *TestEnvironment {
		return &TestEnvironment{
			AllPackageManifests: []*olmpkgv1.PackageManifest{
				{
					Status: olmpkgv1.PackageManifestStatus{
						CatalogSource:          "test-catalog-source",
						CatalogSourceNamespace: "test-catalog-source-namespace",
						Channels: []olmpkgv1.PackageChannel{
							{
								Entries: channelEntries,
							},
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		testEnv  *TestEnvironment
		testCS   *olmv1Alpha.CatalogSource
		expected int
	}{
		{ // Test case 1
			testEnv: generateEnv([]olmpkgv1.ChannelEntry{
				{
					Name: "test-csv.v1.0.0",
				},
				{
					Name: "test-csv.v1.0.1",
				},
			}),
			testCS: &olmv1Alpha.CatalogSource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-catalog-source",
					Namespace: "test-catalog-source-namespace",
				},
			},
			expected: 2,
		},
		{ // Test Case 2 - No matching catalog source found, expecting 0
			testEnv: generateEnv([]olmpkgv1.ChannelEntry{
				{
					Name: "test-csv.v1.0.0",
				},
				{
					Name: "test-csv.v1.0.1",
				},
			}),
			testCS: &olmv1Alpha.CatalogSource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-catalog-source2",
					Namespace: "test-catalog-source-namespace",
				},
			},
			expected: 0,
		},
		{ // Test Case 3 - No images in the catalog source, expecting 0
			testEnv: generateEnv([]olmpkgv1.ChannelEntry{}),
			testCS: &olmv1Alpha.CatalogSource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-catalog-source",
					Namespace: "test-catalog-source-namespace",
				},
			},
			expected: 0,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.expected, GetCatalogSourceBundleCount(testCase.testEnv, testCase.testCS))
	}
}
