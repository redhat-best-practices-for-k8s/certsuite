#!/usr/bin/env bash

set -x
set -e

CHANNELS=(4.10 4.9 4.8 4.7 4.6 4.5 4.4 4.3 4.2 4.1)

rm -f ./cnf-certification-test/platform/operatingsystem/files/rhcos_version_map &>/dev/null

for i in "${CHANNELS[@]}"; do
	VERSIONS=$(curl -sH 'Accept: application/json' "https://api.openshift.com/api/upgrades_info/v1/graph?channel=stable-${i}" | jq '.nodes[].version' -r)

	for VERSION in ${VERSIONS}; do
		# Look up the release version using oc adm.
		RHCOSVERSION="$(oc adm release info "${VERSION}" -o 'jsonpath={.displayVersions.machine-os.Version}')"
		if [[ -n ${RHCOSVERSION} ]]; then
			echo "$VERSION / $RHCOSVERSION" | tee -a ./cnf-certification-test/platform/operatingsystem/files/rhcos_version_map
		fi
	done
done

sort -o ./cnf-certification-test/platform/operatingsystem/files/rhcos_version_map ./cnf-certification-test/platform/operatingsystem/files/rhcos_version_map

echo OpenShift to RHCOS version mapping is in rhcos_version_map
