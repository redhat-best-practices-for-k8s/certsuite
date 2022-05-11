#!/usr/bin/env bash
#set -x

GIT_RELEASE=$(comm -12 all-releases.txt release-tag.txt)
echo "$GIT_RELEASE"
