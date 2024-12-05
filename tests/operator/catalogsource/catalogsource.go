package catalogsource

import (
	olmpkgv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
)

func SkipPMBasedOnChannel(channels []olmpkgv1.PackageChannel, csvName string) bool {
	// This logic is in place because it is possible for an operator to pull from a multiple package manifests.
	skipPMBasedOnChannel := true
	for c := range channels {
		log.Debug("Comparing channel currentCSV %q with current CSV %q", channels[c].CurrentCSV, csvName)
		log.Debug("Number of channel entries %d", len(channels[c].Entries))
		for _, entry := range channels[c].Entries {
			log.Debug("Comparing entry name %q with current CSV %q", entry.Name, csvName)

			if entry.Name == csvName {
				log.Debug("Skipping package manifest based on channel entry %q", entry.Name)
				skipPMBasedOnChannel = false
				break
			}
		}

		if !skipPMBasedOnChannel {
			break
		}
	}

	return skipPMBasedOnChannel
}
