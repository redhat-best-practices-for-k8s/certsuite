#!/usr/bin/env bash

RELEASE_LEVEL=$1
VERSIONS=($(curl -sH 'Accept: application/json' "https://api.openshift.com/api/upgrades_info/v1/graph?channel=stable-${RELEASE_LEVEL}&arch=amd64" | jq -r '.nodes[].version' | sort -t "." -k1,1n -k2,2n -k3,3n))
OPENSHIFT_VERSION=${VERSIONS[${#VERSIONS[@]} - 1]}
echo $OPENSHIFT_VERSION
