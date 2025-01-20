#!/usr/bin/env bash

set -o errexit -o pipefail

CHANNELS=(4.18 4.17 4.16 4.15 4.14 4.13 4.12 4.11 4.10 4.9 4.8 4.7 4.6 4.5)
CHANNEL_TYPES=(stable candidate)

# Keep track of the number of failures we see from the 'oc adm' calls.
SUCCESSFUL_API_ATTEMPTS=0
FAILED_API_ATTEMPTS=0

#shellcheck disable=SC2002
ALREADY_PROCESSED_VERSIONS=$(cat ./tests/platform/operatingsystem/files/rhcos_version_map | awk '{print $1"/"$3}' | sort -u)

# Keep track of successful API query versions that come back missing.
MISSING_VERSIONS=()

rm -f ./tests/platform/operatingsystem/files/rhcos_version_map &>/dev/null

for i in "${CHANNELS[@]}"; do
	echo "----"
	echo "Processing Channel ${i}"
	echo "----"
	for j in "${CHANNEL_TYPES[@]}"; do
		VERSIONS=$(curl -sH 'Accept: application/json' "https://api.openshift.com/api/upgrades_info/v1/graph?channel=${j}-${i}" | jq '.nodes[].version' -r)

		echo "Processing channel ${i} ${j} with number of versions: $(echo "${VERSIONS}" | wc -l)"
		for VERSION in ${VERSIONS}; do
			# Skip versions that were already deemed missing.
			if echo "${MISSING_VERSIONS[@]}" | grep -q "${VERSION}/"; then
				echo "Version: ${VERSION} already deemed missing, skipping."
				continue
			fi

			# Skip versions we've already processed.
			if echo "${ALREADY_PROCESSED_VERSIONS}" | grep -q "${VERSION}/"; then
				echo "Version: ${VERSION} already processed, skipping."

				PROCESSED_VERSION=$(echo "${ALREADY_PROCESSED_VERSIONS}" | grep "${VERSION}/")

				# Find the version split at the '/' character.
				IFS='/' read -r -a VERSION_PARTS <<<"${PROCESSED_VERSION}"

				RHCOSVERSION=${VERSION_PARTS[1]}

				if [[ -z ${RHCOSVERSION} ]]; then
					echo "Failed to find RHCOS version for ${VERSION}."
					continue
				fi

				# Write the already processed version to the file.
				echo "${VERSION} / ${RHCOSVERSION}" |
					tee -a ./tests/platform/operatingsystem/files/rhcos_version_map
				continue
			fi

			# Look up the release version using oc adm.
			NUM_RETRIES=0
			RETRY_LIMIT=5

			echo "Looking up RHCOS version for ${VERSION}..."

			RHCOSVERSION=""

			while [[ ${NUM_RETRIES} -lt ${RETRY_LIMIT} ]]; do
				RHCOSVERSION=$(oc adm release info "${VERSION}" -o 'jsonpath={.displayVersions.machine-os.Version}')
				#shellcheck disable=SC2181
				if [[ $? -eq 0 ]] && [[ -n ${RHCOSVERSION} ]]; then
					SUCCESSFUL_API_ATTEMPTS=$((SUCCESSFUL_API_ATTEMPTS + 1))
					echo "Found RHCOS version ${RHCOSVERSION} for ${VERSION} from the API"
					break
				fi
				NUM_RETRIES=$((NUM_RETRIES + 1))
				FAILED_API_ATTEMPTS=$((FAILED_API_ATTEMPTS + 1))
				sleep 1
			done

			echo "Completed lookup for ${VERSION}."

			if [[ -n ${RHCOSVERSION} ]]; then
				echo "Found RHCOS version ${RHCOSVERSION} for ${VERSION}"
				echo "$VERSION / $RHCOSVERSION" |
					tee -a ./tests/platform/operatingsystem/files/rhcos_version_map
				break
			fi

			echo "Failed to find RHCOS version for ${VERSION}."
			MISSING_VERSIONS+=("${VERSION}/")
		done
	done
done

sort -u -o ./tests/platform/operatingsystem/files/rhcos_version_map ./tests/platform/operatingsystem/files/rhcos_version_map

echo OpenShift to RHCOS version mapping is in rhcos_version_map
echo Number of Successful API Attempts: ${SUCCESSFUL_API_ATTEMPTS}
echo Number of Failed API Attempts: ${FAILED_API_ATTEMPTS}
