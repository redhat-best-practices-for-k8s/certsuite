#!/usr/bin/env bash

# Requires
# - kubeconfig file mounted to /usr/tnf/kubeconfig/config
#   If more than one kubeconfig needs to be used, bind
#   additional volumes for each kubeconfig, e.g.
#     - /usr/tnf/kubeconfig/config
#     - /usr/tnf/kubeconfig/config.2
#     - /usr/tnf/kubeconfig/config.3
# - TNF config files mounted into /usr/tnf/config
# - A directory to output claim into mounted at /usr/tnf/claim
# - A $KUBECONFIG environment variable passed to the TNF container
#   containing all paths to kubeconfigs located in the container, e.g.
#   KUBECONFIG=/usr/tnf/kubeconfig/config:/usr/tnf/kubeconfig/config.2

export REQUIRED_NUM_OF_ARGS=5
export REQUIRED_VARS=('LOCAL_KUBECONFIG' 'LOCAL_TNF_CONFIG' 'OUTPUT_LOC')
export REQUIRED_VARS_ERROR_MESSAGES=(
	'KUBECONFIG is invalid or not given. Use the -k option to provide path to one or more kubeconfig files.'
	'TNFCONFIG is required. Use the -t option to specify the directory containing the TNF configuration files.'
	'OUTPUT_LOC is required. Use the -o option to specify the output location for the test results.'
)

export TNF_IMAGE_NAME=cnf-certification-test
export TNF_IMAGE_TAG=latest
export TNF_OFFICIAL_ORG=quay.io/testnetworkfunction/
export TNF_OFFICIAL_IMAGE="${TNF_OFFICIAL_ORG}${TNF_IMAGE_NAME}:${TNF_IMAGE_TAG}"
export TNF_BIN_DIR="cnf-certification-test"
export TNF_CMD="cnf-certification-test"
export OUTPUT_ARG="-o"
export CONTAINER_NETWORK_MODE='host'

usage() {
	# shellcheck disable=SC2162 # Read without -r will mangle backslashes.
	read -d '' usage_prompt <<-EOF
		Usage: $0 -t TNFCONFIG -o OUTPUT_LOC [-i IMAGE] [-k KUBECONFIG] [-n NETWORK_MODE] [-d DNS_RESOLVER_ADDRESS] [-l LABEL] [-c DOCKERCFG]

		Configure and run the containerised TNF test offering.

		Options (required)
		  -t: set the directory containing TNF config files set up for the test.
		  -o: set the output location for the test results.

		Options (optional)
		  -i: set the TNF container image. Supports local images, as well as images from external registries.
		  -k: set path to one or more local kubeconfigs, separated by a colon.
		      The -k option takes precedence, overwriting the results of local kubeconfig autodiscovery.
		      See the 'Kubeconfig lookup order' section below for more details.
		  -c: set path to one or more local dockercfgs, separated by a colon.
		      The -c option takes precedence, overwriting the results of local dockercfg autodiscovery.
		      See the 'DockerCfg lookup order' section below for more details.
		  -n: set the network mode of the container.
		  -d: set the DNS resolver address for the test containers started by docker, may be required with 
		      certain docker version if the kubeconfig contains host names
		  -l: Set the test labels that should be tested

		Kubeconfig lookup order
		  1. If -k is specified, use the paths provided with the -k option.
		  2. If -k is not specified, use paths defined in \$KUBECONFIG on the underlying host.
		  3. If no paths are defined, use the default kubeconfig file located in '\$HOME/.kube/config'
		     (currently: $HOME/.kube/config).

		Examples
		  $0 -t ~/tnf/config -o ~/tnf/output -f networking access-control -s access-control-host-resource-PRIVILEGED_POD

		  Because -k is omitted, $(basename "$0") will first try to autodiscover local kubeconfig files.
		  If it succeeds, the networking and access-control tests will be run using the autodiscovered configuration.
		  The test results will be saved to the '~/tnf/output' directory on the host.

		  $0 -k ~/.kube/ABC:~/.kube/DEF -t ~/tnf/config -o ~/tnf/output -l "access-control,networking"

		  The command will bind two kubeconfig files (~/.kube/ABC and ~/.kube/DEF) to the TNF container,
		  run the access-control and networking tests, and save the test results into the '~/tnf/output' directory
		  on the host.

		  Because -c is omitted, $(basename "$0") will first try to autodiscover local docker config files.
		  If it succeeds, the networking and access-control tests will be run using the autodiscovered configuration.
		  The test results will be saved to the '~/tnf/output' directory on the host.

		  $0 -c ~/.docker/conf1:~/.docker/conf2 -t ~/tnf/config -o ~/tnf/output -l "access-control,networking"

		  The command will bind two docker config files (~/.docker/conf1 and ~/.docker/conf2) to the TNF container,
		  run the access-control and networking tests, and save the test results into the '~/tnf/output' directory
		  on the host.

		  $0 -i custom-tnf-image:v1.2-dev -t ~/tnf/config -o ~/tnf/output -l "access-control,networking"

		  The command will run the access-control and networking tests as implemented in the custom-tnf-image:v1.2-dev
		  local image set by the -i parameter. The test results will be saved to the '~/tnf/output' directory.

		Test suites
		  Allowed tests are listed in the README.
		  Note: Tests must be specified after all other arguments!
	EOF

	echo -e "$usage_prompt"
}

usage_error() {
	usage
	exit 1
}

check_required_vars() {
	local var_missing=false

	for index in "${!REQUIRED_VARS[@]}"; do
		var=${REQUIRED_VARS[$index]}
		if [[ -z ${!var} ]]; then
			error_message=${REQUIRED_VARS_ERROR_MESSAGES[$index]}
			echo "$0: error: $error_message" 1>&2
			var_missing=true
		fi
	done

	if $var_missing; then
		echo ""
		usage_error
	fi
}

check_cli_required_num_of_args() {
	if (($# < REQUIRED_NUM_OF_ARGS)); then
		usage_error
	fi
}

perform_kubeconfig_autodiscovery() {
	if [[ -n "$KUBECONFIG" ]]; then
		export LOCAL_KUBECONFIG=$KUBECONFIG
		# shellcheck disable=SC2016 # Use double quotes.
		kubeconfig_autodiscovery_source='$KUBECONFIG'
	elif [[ -f "$HOME/.kube/config" ]]; then
		export LOCAL_KUBECONFIG=$HOME/.kube/config
		kubeconfig_autodiscovery_source="\$HOME/.kube/config ($HOME/.kube/config)"
	fi
}

perform_dockercfg_autodiscovery() {
	# As of now, the Docker Config is an optional variable to be supplied as the only test suite that
	# requires it is the preflight suite.
	# See the openshift-preflight documentation about environment variables
	# https://github.com/redhat-openshift-ecosystem/openshift-preflight/blob/main/docs/CONFIG.md#container-policy-configuration
	if [[ -n "$PFLT_DOCKERCONFIG" ]]; then
		export LOCAL_DOCKERCFG=$PFLT_DOCKERCONFIG

		# shellcheck disable=SC2016 # Don't expand in single quotes.
		dockercfg_autodiscovery_source='$PFLT_DOCKERCONFIG'
	elif [[ -f "$HOME/.docker/config.json" ]]; then
		export LOCAL_DOCKERCFG=$HOME/.docker/config.json
		dockercfg_autodiscovery_source="\$HOME/.docker/config.json ($HOME/.docker/config.json)"
	else
		# Default Case: Set the variable to NA (Not Applicable)
		export LOCAL_DOCKERCFG=NA
	fi
}

display_kubeconfig_autodiscovery_summary() {
	if [[ -n "$kubeconfig_autodiscovery_source" ]]; then
		echo "Kubeconfig Autodiscovery: configuration loaded from $kubeconfig_autodiscovery_source"
	fi
}

display_dockercfg_autodiscovery_summary() {
	if [[ -n "$dockercfg_autodiscovery_source" ]]; then
		echo "DockerCfg Autodiscovery: configuration loaded from $dockercfg_autodiscovery_source"
	fi

	# Let the user know that the LOCAL_DOCKERCFG is set to NA
	if [[ $LOCAL_DOCKERCFG == 'NA' ]]; then
		echo "LOCAL_DOCKERCFG being set to 'NA' because PFLT_DOCKERCONFIG is not set and HOME/.docker/config.json does not exist."
	fi
}

if [ -n "${REQUIRED_NUM_OF_ARGS}" ]; then
	check_cli_required_num_of_args "$@"
fi

echo "Performing KUBECONFIG autodiscovery"
perform_kubeconfig_autodiscovery
echo "Performing DOCKERCFG autodiscovery"
perform_dockercfg_autodiscovery

# Parse args beginning with -
while [[ $1 == -* ]]; do
	case "$1" in
	-h | --help | -\?)
		usage
		exit 0
		;;
	-k)
		if (($# > 1)); then
			export LOCAL_KUBECONFIG=$2
			unset kubeconfig_autodiscovery_source
			shift
		else
			echo "-k requires an argument" 1>&2
			exit 1
		fi
		echo "-k $LOCAL_KUBECONFIG"
		;;
	-c)
		if (($# > 1)); then
			export LOCAL_DOCKERCFG=$2
			unset dockercfg_autodiscovery_source
			shift
		else
			echo "-c requires an argument" 1>&2
			exit 1
		fi
		echo "-c $LOCAL_DOCKERCFG"
		;;
	-t)
		if (($# > 1)); then
			export LOCAL_TNF_CONFIG=$2
			shift
		else
			echo "-t requires an argument" 1>&2
			exit 1
		fi
		echo "-t $LOCAL_TNF_CONFIG"
		;;
	-b)
		if (($# > 1)); then
			export LOCAL_TNF_OFFLINE_DB=$2
			shift
		else
			echo "-b requires an argument" 1>&2
			exit 1
		fi
		echo "-b $LOCAL_TNF_OFFLINE_DB"
		;;
	-o)
		if (($# > 1)); then
			export OUTPUT_LOC=$2
			shift
		else
			echo "-o requires an argument" 1>&2
			exit 1
		fi
		echo "-o $OUTPUT_LOC"
		;;
	-i)
		if (($# > 1)); then
			export TNF_IMAGE=$2
			shift
		else
			echo "-i requires an argument" 1>&2
			exit 1
		fi
		echo "-i $TNF_IMAGE"
		;;
	-n)
		if (($# > 1)); then
			export CONTAINER_NETWORK_MODE=$2
			shift
		else
			echo "-n requires an argument" 1>&2
			exit 1
		fi
		echo "-n $CONTAINER_NETWORK_MODE"
		;;
	-d)
		if (($# > 1)); then
			export DNS_ARG=$2
			shift
		else
			echo "-d requires an argument" 1>&2
			exit 1
		fi
		echo "-d $DNS_ARG"
		;;
	-l)
		while (("$#" >= 2)) && ! [[ $2 = --* ]] && ! [[ $2 = -* ]]; do
			TNF_LABEL="$TNF_LABEL $2"
			shift
		done
		TNF_LABEL="$(echo -e "${TNF_LABEL}" | sed -e 's/^[[:space:]]*//')" # strip the leading whitespace
		export TNF_LABEL
		echo "-l $TNF_LABEL"
		;;
	--)
		shift
		break
		;;
	-*)
		echo "invalid option: $1" 1>&2
		usage_error
		;;
	esac
	shift
done

display_kubeconfig_autodiscovery_summary
display_dockercfg_autodiscovery_summary
check_required_vars
cd script || exit 1
./run-container.sh "$@"
