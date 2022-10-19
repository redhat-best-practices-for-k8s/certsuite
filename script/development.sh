#!/bin/sh

LABEL=$1

make build-cnf-tests 

./run-cnf-suites.sh -l $LABEL
