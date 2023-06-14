#!/usr/bin/env bash

# [debug] uncomment line below to print out the statements as they are being
# executed.
set -x

# defaults
export OUTPUT_LOC="$PWD/cnf-certification-test"

usage() {
	echo "$0 [-o OUTPUT_LOC] [-l LABEL...]"
	echo "Call the script and list the test suites to run"
	echo "  e.g."
	echo "    $0 [ARGS] -l \"access-control,lifecycle\""
	echo "  will run the access-control and lifecycle suites"
	echo ""
	echo "Allowed suites are listed in the README."
	echo ""
	echo "The specs can be listed with $0 -L|--list [-l LABEL...]"
}

usage_error() {
	usage
	exit 1
}

TIMEOUT=24h0m0s
FOCUS=''
SKIP=''
LABEL=''
LIST=false
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
	-s | --skip)
		while (("$#" >= 2)) && ! [[ $2 = --* ]] && ! [[ $2 = -* ]]; do
			SKIP="$2|$SKIP"
			shift
		done
		;;
	-f | --focus)
		while (("$#" >= 2)) && ! [[ $2 = --* ]] && ! [[ $2 = -* ]]; do
			FOCUS="$2|$FOCUS"
			shift
		done
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

# List the specs (filtering by suite).
if [ "$LIST" = true ]; then
	cd "$BASEDIR"/cnf-certification-test || exit 1
	./cnf-certification-test.test \
		--ginkgo.dry-run \
		--ginkgo.timeout=$TIMEOUT \
		--ginkgo.v \
		--ginkgo.label-filter="$LABEL"
	cd ..
	exit 0
fi

# Specify Junit report file name.
GINKGO_ARGS="\
--ginkgo.timeout=$TIMEOUT \
-junit $OUTPUT_LOC \
-claimloc $OUTPUT_LOC \
--ginkgo.junit-report $OUTPUT_LOC/cnf-certification-tests_junit.xml \
-ginkgo.v \
-test.v\
"

FOCUS=${FOCUS%?}
SKIP=${SKIP%?}

echo "Running with focus '$FOCUS'"
echo "Running with skip '$SKIP'"
echo "Running with label filter '$LABEL'"
echo "Report will be output to '$OUTPUT_LOC'"
echo "ginkgo arguments '${GINKGO_ARGS}'"
FOCUS_STRING=''
SKIP_STRING=''
LABEL_STRING=''

if [ -n "$FOCUS" ]; then
	FOCUS_STRING=-ginkgo.focus="${FOCUS}"
	if [ -n "$SKIP" ]; then
		SKIP_STRING=-ginkgo.skip="${SKIP}"
	fi
fi

if [ -z "$FOCUS_STRING" ] && [ -z "$LABEL" ]; then
	echo "No test focus (-f) or label (-l) was set, so only diagnostic functions will run."
else
	# Add the label "common" in case no labels have been provided. This will
	# allow to filter out some either non-official TCs or TCs not intended to run
	# in CI (yet).
	if [ -n "$LABEL" ]; then
		LABEL_STRING="-ginkgo.label-filter=${LABEL}"
	else
		LABEL_STRING='-ginkgo.label-filter=common'
	fi
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
./cnf-certification-test.test \
	${FOCUS_STRING} \
	${SKIP_STRING} \
	"${LABEL_STRING}" \
	${GINKGO_ARGS} |& tee $OUTPUT_LOC/tnf-execution.log

# preserving the exit status
RESULT=$?

# revert to normal mode
set +o pipefail

# exit with retrieved exit status
exit $RESULT
