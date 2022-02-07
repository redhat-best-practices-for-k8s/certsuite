#!/usr/bin/env bash
#set -x
GIT_RELEASE=$(comm -12 all-releases.txt release-tag.txt) 
GIT_PREVIOUS_RELEASE=$(comm -12 all-releases.txt latest-release-tag.txt)
GIT_LATEST_RELEASE=$GIT_RELEASE
if [ -z "$GIT_RELEASE" ]; then
   GIT_LATEST_RELEASE=$GIT_PREVIOUS_RELEASE
fi
echo $GIT_LATEST_RELEASE