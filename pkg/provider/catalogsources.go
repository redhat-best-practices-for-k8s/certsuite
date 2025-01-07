package provider

import (
	"strconv"
	"strings"

	"github.com/Masterminds/semver"
	olmv1Alpha "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
)

func GetCatalogSourceBundleCount(env *TestEnvironment, cs *olmv1Alpha.CatalogSource) int {
	// Now that we know the catalog source, we are going to count up all of the relatedImages
	// that are associated with the catalog source. This will give us the number of bundles that
	// are available in the catalog source.

	// If the OCP version is <= 4.12, we need to use the probe container to get the bundle count
	const (
		ocpMajorVersion = 4
		ocpMinorVersion = 12
	)

	// Check if the cluster is running an OCP version <= 4.12
	if env.OpenshiftVersion != "" {
		log.Info("Cluster is determined to be running Openshift version %q.", env.OpenshiftVersion)
		version, err := semver.NewVersion(env.OpenshiftVersion)
		if err != nil {
			log.Error("Failed to parse Openshift version %q.", env.OpenshiftVersion)
			return 0
		}

		if version.Major() < ocpMajorVersion || (version.Major() == ocpMajorVersion && version.Minor() <= ocpMinorVersion) {
			return getCatalogSourceBundleCountFromProbeContainer(env, cs)
		}

		// If we didn't find the bundle count via the probe container, we can attempt to use the package manifests
	}

	// If we didn't find the bundle count via the probe container, we can use the package manifests
	// to get the bundle count
	return getCatalogSourceBundleCountFromPackageManifests(env, cs)
}

func getCatalogSourceBundleCountFromProbeContainer(env *TestEnvironment, cs *olmv1Alpha.CatalogSource) int {
	// We need to use the probe container to get the bundle count
	// This is because the package manifests are not available in the cluster
	// for OCP versions <= 4.12
	o := clientsholder.GetClientsHolder()

	// Find the kubernetes service associated with the catalog source
	for _, svc := range env.AllServices {
		// Skip if the service is not associated with the catalog source
		if svc.Spec.Selector["olm.catalogSource"] != cs.Name {
			continue
		}

		log.Info("Found service %q associated with catalog source %q.", svc.Name, cs.Name)

		// Use a probe pod to get the bundle count
		for _, probePod := range env.ProbePods {
			ctx := clientsholder.NewContext(probePod.Namespace, probePod.Name, probePod.Spec.Containers[0].Name)
			cmd := "grpcurl -plaintext " + svc.Spec.ClusterIP + ":50051 api.Registry.ListBundles | jq -s 'length'"
			cmdValue, errStr, err := o.ExecCommandContainer(ctx, cmd)
			if err != nil || errStr != "" {
				log.Error("Failed to execute command %s in probe pod %s", cmd, probePod.String())
				continue
			}

			// Sanitize the command output
			cmdValue = strings.TrimSpace(cmdValue)
			cmdValue = strings.Trim(cmdValue, "\"")

			// Parse the command output
			bundleCount, err := strconv.Atoi(cmdValue)
			if err != nil {
				log.Error("Failed to convert bundle count to integer: %s", cmdValue)
				continue
			}

			// Try each probe pod until we get a valid bundle count (which should only be 1 probe pod)
			log.Info("Found bundle count via grpcurl %d for catalog source %q.", bundleCount, cs.Name)
			return bundleCount
		}
	}

	log.Warn("Warning: No services found associated with catalog source %q.", cs.Name)
	return -1
}

func getCatalogSourceBundleCountFromPackageManifests(env *TestEnvironment, cs *olmv1Alpha.CatalogSource) int {
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
