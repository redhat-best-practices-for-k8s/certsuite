#!/usr/bin/env bash

set -x
set -e

CHANNELS=(4.18 4.17 4.16 4.15 4.14 4.13 4.12 4.11 4.10 4.9 4.8 4.7 4.6 4.5 4.4 4.3 4.2 4.1)
CHANNEL_TYPES=(stable candidate)

rm -f ./tests/platform/operatingsystem/files/rhcos_version_map &>/dev/null

for i in "${CHANNELS[@]}"; do
	for j in "${CHANNEL_TYPES[@]}"; do
		VERSIONS=$(curl -sH 'Accept: application/json' "https://api.openshift.com/api/upgrades_info/v1/graph?channel=${j}-${i}" | jq '.nodes[].version' -r)
		for VERSION in ${VERSIONS}; do
			# Look up the release version using oc adm.
			if RHCOSVERSION="$(
				oc adm release \
					info "${VERSION}" \
					-o 'jsonpath={.displayVersions.machine-os.Version}'
			)"; then
				if [[ -n ${RHCOSVERSION} ]]; then
					echo "$VERSION / $RHCOSVERSION" |
						tee -a ./tests/platform/operatingsystem/files/rhcos_version_map
				fi
			else
				printf 'Continue with an error.\n'
			fi
		done
	done
done

sort -u -o ./tests/platform/operatingsystem/files/rhcos_version_map ./tests/platform/operatingsystem/files/rhcos_version_map

echo OpenShift to RHCOS version mapping is in rhcos_version_map
