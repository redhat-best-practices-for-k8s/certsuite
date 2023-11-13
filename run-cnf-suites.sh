#!/usr/bin/env bash

# [debug] uncomment line below to print out the statements as they are being
# executed.
set -x

# defaults
export OUTPUT_LOC="$PWD/cnf-certification-test"

usage() {
	echo "$0 [-o OUTPUT_LOC] [-l LABEL...] [-s run from webpage]"
	echo "Call the script and list the test suites to run"
	echo "  e.g."
	echo "    $0 [ARGS] -l \"access-control,lifecycle\""
	echo "  will run the access-control and lifecycle suites"
	echo "    $0 [ARGS] -l all will run all the tests"
	echo "    $0 [ARGS] -s true will run the test from server"
	echo ""
	echo "Allowed suites are listed in the README."
	echo ""
	echo "The specs can be listed with $0 -L|--list [-l LABEL...] [-s run from webpage]"
}

usage_error() {
	usage
	exit 1
}

TIMEOUT=24h0m0s
LABEL=''
LIST=false
SERVER_RUN=false
BASEDIR=$(dirname "$(realpath "$0")")

# Parse args beginning with "-".
while [[ $1 == -* ]]; do
	case "$1" in
	-h | --help | -\?)
		usage
		exit 0
		;;
	-L | --list)
		LIST=true
		;;
	-o)
		if (($# > 1)); then
			OUTPUT_LOC=$2
			shift
		else
			echo >&2 '-o requires an argument'
			exit 1
		fi
		;;
	-s)
		SERVER_RUN=true
		;;
	-l | --label)
		while (("$#" >= 2)) && ! [[ $2 = --* ]] && ! [[ $2 = -* ]]; do
			LABEL="$LABEL $2"
			shift
		done
		;;
	-*)
		echo >&2 "invalid option: $1"
		usage_error
		;;
	esac
	shift
done

# Strips the leading whitespace.
LABEL="$(echo -e "${LABEL}" | sed -e 's/^[[:space:]]*//')"

if [[ $LABEL == "all" ]]; then
	LABEL='common,extended,faredge,telco'
fi
# List the specs (filtering by suite).
if [ "$LIST" = true ]; then
	cd "$BASEDIR"/cnf-certification-test || exit 1
	./cnf-certification-test.test \
		--list \
		--label-filter="$LABEL"
	cd ..
	exit 0
fi

# Specify Junit report file name.
EXTRA_ARGS="\
--timeout=$TIMEOUT \
-claimloc $OUTPUT_LOC \
"

if [ "$SERVER_RUN" = "true" ]; then
	EXTRA_ARGS="$EXTRA_ARGS -serverMode"
fi

echo "Label: $LABEL"
if [[ $LABEL == "all" ]]; then
	LABEL='common,extended,faredge,telco'
fi

echo "Running with label filter '$LABEL'"
echo "Report will be output to '$OUTPUT_LOC'"
echo "Extra arguments '${EXTRA_ARGS}'"
LABEL_STRING=''

if [ -z "$LABEL" ] && { [ -z "$SERVER_RUN" ] || [ "$SERVER_RUN" == "false" ]; }; then
	echo "No test label (-l) was set, so only diagnostic functions will run."
else
	LABEL_STRING="--label-filter=${LABEL}"
fi

cd "$BASEDIR"/cnf-certification-test || exit 1

# configuring special pipeline mode
# The exit status of a pipeline is the exit status of the last command in the pipeline, unless the pipefail option
# is enabled (see The Set Builtin). If pipefail is enabled, the pipeline's return status is the value of the last (rightmost)
# command to exit with a non-zero status, or zero if all commands exit successfully.
set -o pipefail

# Do not double quote.
# SC2086: Double quote to prevent globbing and word splitting.
# shellcheck disable=SC2086
./cnf-certification-test \
	"${LABEL_STRING}" \
	${EXTRA_ARGS} |& tee $OUTPUT_LOC/tnf-execution.log

# preserving the exit status
RESULT=$?

# revert to normal mode
set +o pipefail

# exit with retrieved exit status
exit $RESULT
