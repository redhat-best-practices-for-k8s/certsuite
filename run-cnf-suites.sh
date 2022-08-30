#!/usr/bin/env bash

# [debug] uncomment line below to print out the statements as they are being executed
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

LABEL=""
LIST=false
BASEDIR=$(dirname $(realpath $0))
# Parse args beginning with "-"
while [[ $1 == -* ]]; do
	case "$1" in
		-h|--help|-\?) usage; exit 0;;
		-L|--list) LIST=true;;
		-o) if (($# > 1)); then
				OUTPUT_LOC=$2; shift
			else
				echo "-o requires an argument" 1>&2
				exit 1
			fi ;;
		-l|--label)
			while (( "$#" >= 2 )) && ! [[ $2 = --* ]]  && ! [[ $2 = -* ]] ; do
				# LABEL="$2|$LABEL"
				LABEL="$LABEL $2"
				shift
			done;;
		-*) echo "invalid option: $1" 1>&2; usage_error;;
	esac
	shift
done

# Check if the test cluster runs OCP
res=`oc version | grep  Server`
if [ -z "$res" ]
then
   echo "Standard Kubernetes cluster detected (not OCP)"
   export TNF_NON_OCP_CLUSTER=true
fi

# List the specs (filtering by suite)
if [ "$LIST" = true ] ; then
	LABEL="$(echo -e "${LABEL}" | sed -e 's/^[[:space:]]*//')" # strip the leading whitespace
	cd $BASEDIR/cnf-certification-test
	./cnf-certification-test.test --ginkgo.dry-run --ginkgo.v --ginkgo.label-filter="$LABEL"
	cd ..
	exit 0;
fi

# specify Junit report file name.
GINKGO_ARGS="-junit $OUTPUT_LOC -claimloc $OUTPUT_LOC --ginkgo.junit-report $OUTPUT_LOC/cnf-certification-tests_junit.xml -ginkgo.v -test.v"

# Make sure the HTML output is copied to the output directory,
# even in case of a test failure
function html_output() {
    if [ -f ${OUTPUT_LOC}/claim.json ]; then
        echo -n "var initialjson=" > ${OUTPUT_LOC}/claimjson.js
        cat ${OUTPUT_LOC}/claim.json >>  ${OUTPUT_LOC}/claimjson.js
    fi
    cp ${BASEDIR}/script/results.html ${OUTPUT_LOC}
}
trap html_output EXIT

LABEL="$(echo -e "${LABEL}" | sed -e 's/^[[:space:]]*//')" # strip the leading whitespace

# Run cnf-feature-deploy test container if not running inside a container
# cgroup file doesn't exist on MacOS. Consider that as not running in container as well
if [[ ! -f "/proc/1/cgroup" ]] || grep -q init\.scope /proc/1/cgroup; then
	cd script
	./run-cfd-container.sh
	cd ..
fi

echo "Running with label filter '$LABEL'"
echo "Report will be output to '$OUTPUT_LOC'"
echo "ginkgo arguments '${GINKGO_ARGS}'"
LABEL_STRING=""
if [ -n "$LABEL" ]; then
    LABEL_STRING="-ginkgo.label-filter=${LABEL}"
else
	echo "No test label (-l) was set, so only diagnostic functions will run.".
fi

cd ./cnf-certification-test && ./cnf-certification-test.test "${LABEL_STRING}" ${GINKGO_ARGS}

# if [[ ! -z "${TNF_PARTNER_SRC_DIR}" ]]; then
# 	echo "attempting to delete litmus"
# 	make -C $TNF_PARTNER_SRC_DIR delete-litmus
# fi
