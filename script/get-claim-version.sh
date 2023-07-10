#!/usr/bin/env bash
#set -x

cat version.json | jq .claimFormat
