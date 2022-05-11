#!/usr/bin/env bash
#set -x

GIT_PREVIOUS_RELEASE=$(comm -12 all-releases.txt latest-release-tag.txt)
echo "$GIT_PREVIOUS_RELEASE"
