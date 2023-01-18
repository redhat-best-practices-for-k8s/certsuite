#!/bin/sh
LABEL=$1

# builds the suite executable
make build-cnf-tests

# runs the tests
./run-cnf-suites.sh -l "$LABEL"
