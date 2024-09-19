#!/usr/bin/env bash
#set -x

GIT_COMMIT=$(git rev-list -1 HEAD)

curl -s https://api.github.com/repos/redhat-best-practices-for-k8s/certsuite/releases | jq -r ".[].tag_name" | sort >all-releases.txt
git tag --points-at HEAD | sort >release-tag.txt
git tag --no-contains "${GIT_COMMIT}" | tail -n1 | sort >latest-release-tag.txt

echo "$GIT_COMMIT"
