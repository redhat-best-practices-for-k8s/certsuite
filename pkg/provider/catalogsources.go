package provider

import (
	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

func GetCatalogSourceBundleCount(env *TestEnvironment, cs *olmv1Alpha.CatalogSource) int {
	// Now that we know the catalog source, we are going to count up all of the relatedImages
	// that are associated with the catalog source. This will give us the number of bundles that
	// are available in the catalog source.

	totalRelatedBundles := 0
	for _, pm := range env.AllPackageManifests {
		// Skip if the package manifest is not associated with the catalog source
		if pm.Status.CatalogSource != cs.Name || pm.Status.CatalogSourceNamespace != cs.Namespace {
			continue
		}

		// Count up the number of related bundles
		for c := range pm.Status.Channels {
			totalRelatedBundles += len(pm.Status.Channels[c].Entries)
		}
	}

	return totalRelatedBundles
}
