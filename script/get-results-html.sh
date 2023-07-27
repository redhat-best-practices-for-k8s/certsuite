#!/usr/bin/env bash

set -x
git clone --depth=1 --branch="$1" "https://github.com/test-network-function/parser.git" temp-html
cp temp-html/html/results.html cnf-certification-test/results/html/.
rm -rf temp-html
