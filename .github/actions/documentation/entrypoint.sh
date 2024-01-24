#!/usr/bin/env sh
set -e

cd "${GITHUB_WORKSPACE}"

mkdocs build
