#!/usr/bin/env bash
set -x
# defaults
export OUTPUT_LOC="$PWD/cnf-certification-test"

usage() {
	echo "$0 [-o OUTPUT_LOC] [-f SUITE...] -s [SUITE...] [-l LABEL...]"
	echo "Call the script and list the test suites to run"
	echo "  e.g."
	echo "    $0 [ARGS] -f access-control lifecycle"
	echo "  will run the access-control and lifecycle suites"
	echo ""
	echo "Allowed suites are listed in the README."
}

usage_error() {
	usage
	exit 1
}

FOCUS=""
SKIP=""
LABEL=""
BASEDIR=$(dirname $(realpath $0))
# Parge args beginning with "-"
while [[ $1 == -* ]]; do
	case "$1" in
		-h|--help|-\?) usage; exit 0;;
		-o) if (($# > 1)); then
				OUTPUT_LOC=$2; shift
			else
				echo "-o requires an argument" 1>&2
				exit 1
			fi ;;
		-s|--skip)
			while (( "$#" >= 2 )) && ! [[ $2 = --* ]] && ! [[ $2 = -* ]] ; do
				SKIP="$2|$SKIP"
				shift
			done;;
		-f|--focus)
			while (( "$#" >= 2 )) && ! [[ $2 = --* ]]  && ! [[ $2 = -* ]] ; do
				FOCUS="$2|$FOCUS"
				shift
			done;;
		-l|--label)
			while (( "$#" >= 2 )) && ! [[ $2 = --* ]]  && ! [[ $2 = -* ]] ; do
				LABEL="$2|$LABEL"
				shift
			done;;
		-*) echo "invalid option: $1" 1>&2; usage_error;;
	esac
	shift
done
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


# If no focus is set then display usage and quit with a non-zero exit code.
[ -z "$FOCUS" ] && echo "no focus found" && usage_error

FOCUS=${FOCUS%?}  # strip the trailing "|" from the concatenation
SKIP=${SKIP%?} # strip the trailing "|" from the concatenation
LABEL=${LABEL%?} # strip the trailing "|" from the concatenation

res=`oc version | grep  Server`
if [ -z "$res" ]
then
   echo "Minikube or similar detected"
   export TNF_NON_OCP_CLUSTER=true
fi
# Run cnf-feature-deploy test container if not running inside a container
# cgroup file doesn't exist on MacOS. Consider that as not running in container as well
if [[ ! -f "/proc/1/cgroup" ]] || grep -q init\.scope /proc/1/cgroup; then
	cd script
	./run-cfd-container.sh
	cd ..
fi

if [[ -z "${TNF_PARTNER_SRC_DIR}" ]]; then
	echo "env var \"TNF_PARTNER_SRC_DIR\" not set, running the script without updating infra"
else
	make -C $TNF_PARTNER_SRC_DIR install-partner-pods
fi

echo "Running with focus '$FOCUS'"
echo "Running with skip  '$SKIP'"
echo "Running with label filter '$LABEL'"
echo "Report will be output to '$OUTPUT_LOC'"
echo "ginkgo arguments '${GINKGO_ARGS}'"
SKIP_STRING=""
LABEL_STRING=""
if [ -n "$SKIP" ]; then
	SKIP_STRING=-ginkgo.skip="$SKIP"
fi
if [ -n "$LABEL" ]; then
    LABEL_STRING=-ginkgo.label-filter="$LABEL"
fi

cd ./cnf-certification-test && ./cnf-certification-test.test -ginkgo.focus="$FOCUS" $SKIP_STRING $LABEL_STRING ${GINKGO_ARGS}
