#!/bin/bash
set -o errexit -o nounset -o pipefail

# Test run timestamp
TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S_%Z")

# Base folder
BASE_DIR=/var/www/html

# index.html
INDEX_FILE=index2.html

# INPUTS

# tnf_config.yaml template file path
CONFIG_YAML_TEMPLATE="$(pwd)"/tnf_config.yml.template

# Docker config used to pull operator images
DOCKER_CONFIG=config.json

# Location of telco/non-telco classification file
CNF_TYPE=cmd/tnf/claim/show/csv/cnf-type.json

# Operator catalog from user
OPERATOR_CATALOG=""

# Operator from user
OPERATORS_UNDER_TEST=""

# OUTPUTS

# Colors
readonly \
	RED="\033[31m" \
	GREEN="\033[32m" \
	BLUE="\033[36m" \
	GREY="\033[90m" \
	ENDCOLOR="\033[0m"

# Check if DEBUG mode
if [ -n "${DEBUG_RUN+any}" ]; then
	echo "DEBUG_RUN is set. Running in debug mode"
	# Debug folder
	REPORT_FOLDER_RELATIVE="debug_$TIMESTAMP"
else
	echo "DEBUG_RUN is not set. Running in non-debug mode"
	# Report folder
	REPORT_FOLDER_RELATIVE="report_$TIMESTAMP"
fi

# Report results folder
REPORT_FOLDER="$BASE_DIR"/"$REPORT_FOLDER_RELATIVE"

# Operator  file name
OPERATOR_LIST_FILENAME=operator-list.txt

# Operator list path in the report
OPERATOR_LIST_PATH="$REPORT_FOLDER"/"$OPERATOR_LIST_FILENAME"

# Log file
LOG_FILENAME="output_$TIMESTAMP.log"

# Operator list path in the report
LOG_FILE_PATH="$REPORT_FOLDER"/"$LOG_FILENAME"

echo_color() {
	local color=$1
	local format=$2
	shift 2
	# shellcheck disable=SC2059
	printf "$color$format$ENDCOLOR\n" "$@"
	# shellcheck disable=SC2059
	printf "$format" "$@" >>"$LOG_FILE_PATH"
}

# VARIABLES

# Variable to add header only on the first run
add_headers=-a

# Create report directory
mkdir "$REPORT_FOLDER"

cleanup() {
	# Workaround for cleaning operator leftovers, see https://access.redhat.com/solutions/6971276
	oc delete mutatingwebhookconfigurations controller.devfile.io || true
	oc delete validatingwebhookconfigurations controller.devfile.io || true

	# Leftovers specific to certain operators
	oc delete Validating_webhook_configuration sriov-operator-webhook-config || true
	oc delete Mutating_webhook_configuration sriov-operator-webhook-config || true
}

wait_delete_namespace() {
	local namespace_deleting=$1
	# Wait for the namespace to be removed
	if [ "$namespace_deleting" != "openshift-operators" ]; then

		echo_color "$BLUE" "non openshift-operators namespace = $namespace_deleting, deleting "
		with_retry 2 0 oc wait namespace "$namespace_deleting" --for=delete --timeout=60s || true

		force_delete_namespace_if_present "$namespace_deleting" >>"$LOG_FILE_PATH" 2>&1 || true
	fi
}

# Executes command with retry
with_retry() {
	local \
		max_retries=$1 \
		timeout=$2 \
		retries=0 \
		status=0 \
		stderr='' \
		stdout=''
	shift 2

	until [ "$retries" -ge "$max_retries" ]; do
		# Execute oc command saving stdout, stderr and exit status
		# see: https://stackoverflow.com/questions/11027679/capture-stdout-and-stderr-into-different-variables/41069638#41069638
		unset stdout stderr status
		eval "$(
			(
				#oc command
				$"$@"
			) \
				2> >(
					# shellcheck disable=SC2030
					stderr=$(cat)
					typeset -p stderr
				) \
				> >(
					# shellcheck disable=SC2030
					stdout=$(cat)
					typeset -p stdout
				)
			# shellcheck disable=SC2030
			status=$?
			typeset -p status
		)"
		# shellcheck disable=SC2031
		if [ "$status" -eq 0 ]; then
			# If the command succeeded, break out of the loop
			# shellcheck disable=SC2031
			echo "$stdout"
			{
				echo "command: $*"
				echo "stderr: $stderr"
				echo "stdout: $stdout"
				echo "status: $status"
			} >>"$LOG_FILE_PATH"
			return 0
		fi
		# If the command failed, increment the retry counter
		retries=$((retries + 1))
		# shellcheck disable=SC2031
		{
			echo "command: $*"
			echo "stderr: $stderr"
			echo "stdout: $stdout"
			echo "status: $status"
		} >>"$LOG_FILE_PATH"

		echo_color "$GREY" "Retry $retries/$max_retries: Waiting for a few seconds before the next attempt..."
		sleep "$timeout"
	done
	echo_color "$GREY" "Maximum retries reached."
	return 1
}

wait_cluster_ok() {
	local \
		start_time \
		status \
		timeout_seconds=600

	start_time=$(date +%s 2>&1) || {
		echo "date failed with error $?: $start_time" >>"$LOG_FILE_PATH"
		return 0
	}
	while true; do
		status=0
		oc get nodes &>/dev/null &&
			return 0 ||
			echo "get nodes failed with error $?." >>"$LOG_FILE_PATH"
		current_time=$(date +%s)
		elapsed_time=$((current_time - start_time))
		# If elapsed time is greater than the timeout report failure
		if [ "$elapsed_time" -ge "$timeout_seconds" ]; then
			echo_color "$BLUE" "Timeout reached $timeout_seconds seconds waiting for cluster to be reachable."
			return 1
		fi

		# Otherwise wait a bit
		echo_color "$BLUE" "Waiting for cluster to be reachable..."
		sleep 5
	done
}

wait_for_csv_to_appear_and_label() {
	local csv_namespace=$1
	local timeout_seconds=100
	local start_time=0
	local current_time=0
	local elapsed_time=0
	local command=""
	local status=0

	start_time=$(date +%s)
	while true; do
		csvs=$(oc get csv -n "$csv_namespace" 2>>"$LOG_FILE_PATH") || true
		if [ "$csvs" != "" ]; then
			# If any CSV is present, break
			break
		else
			current_time=$(date +%s)
			elapsed_time=$((current_time - start_time))
			# If elapsed time is greater than the timeout report failure
			if [ "$elapsed_time" -ge "$timeout_seconds" ]; then
				echo_color "$BLUE" "Timeout reached $timeout_seconds seconds waiting for CSV."
				return 1
			fi

			# Otherwise wait a bit
			echo_color "$BLUE" "Waiting for csv to be created in namespace $csv_namespace ..."
			sleep 5
		fi
	done

	# Label CSV with "test-network-function.com/operator=target"
	command=$(with_retry 5 10 oc get csv -n "$csv_namespace" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | grep -v openshift-operator-lifecycle-manager | sed '/^ *$/d' | awk '{print "  with_retry 5 10 oc label " $3  " -n " $2 " " $1  " test-network-function.com/operator=target "}')
	eval "$command"

	# Wait for the CSV to be succeeded
	echo_color "$BLUE" "Wait for CSV to be succeeded"
	with_retry 30 0 oc wait csv -l test-network-function.com/operator=target -n "$ns" --for=jsonpath=\{.status.phase\}=Succeeded --timeout=5s || status="$?"
	return $status
}

force_delete_namespace_if_present() {
	local a_namespace=$1
	local pid=0

	# Do not delete the redhat-operators namespace
	if [ "$a_namespace" = "openshift-operators" ]; then
		return 0
	fi

	# Delete namespace
	oc delete namespace "$a_namespace" --wait=false || true
	with_retry 2 0 oc wait namespace "$a_namespace" --for=delete --timeout=5s || true

	# If a namespace with this name does not exist, all is good, exit
	if ! oc get namespace "$a_namespace"; then
		return 0
	fi
	echo_color "$RED" "Namespace cannot be deleted normally, force deleting"
	# Otherwise force delete namespace
	with_retry 5 10 oc get namespace "$a_namespace" -ojson | sed '/"kubernetes"/d' >temp.yaml
	# Kill previous oc proxy command in the background
	killall "oc"
	# Start a new proxy
	oc proxy &
	pid=$!
	echo "PID: $pid"
	sleep 5
	curl -H "Content-Type: application/yaml" -X PUT --data-binary @temp.yaml http://127.0.0.1:8001/api/v1/namespaces/"$a_namespace"/finalize >>"$LOG_FILE_PATH"
	kill -9 "$pid"
	with_retry 2 0 oc wait namespace "$a_namespace" --for=delete --timeout=60s
}

report_failure() {
	local status=$1
	local ns=$2
	local package_name=$3
	local message=$4

	with_retry 3 5 oc operator uninstall -X "$package_name" -n "$ns" || true
	# Add per operator links
	{
		# Add error message
		echo "Results for: <b>$package_name</b>, "'<span style="color: red;">'"$message"'</span>'

		# Add tnf_config link
		echo ", tnf_config: "
		echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$package_name"'/tnf_config.yml">'"link"'</a>'

		# New line
		echo "<br>"
	} >>"$REPORT_FOLDER"/"$INDEX_FILE"

	wait_delete_namespace "$ns"
}

get_suggested_namespace() {
	local package_name=$1

	oc get packagemanifests -n openshift-marketplace "$package_name" -ojson | jq -r '.status.channels[].currentCSVDesc.annotations."operatorframework.io/suggested-namespace"' 2>/dev/null | grep -v "null" | sed 's/\n//g' | head -1 || true
}

# Main

# Check if the number of parameters is correct
if [ "$#" -eq 1 ]; then
	OPERATOR_CATALOG=$1
	# Get all the packages present in the cluster catalog
	with_retry 5 10 oc get packagemanifest -o jsonpath='{range .items[*]}{.metadata.name}{","}{.status.catalogSource}{"\n"}{end}' | grep "$OPERATOR_CATALOG" | head -n -1 | sort >"$OPERATOR_LIST_PATH"

elif [ "$#" -eq 2 ]; then
	OPERATOR_CATALOG=$1
	OPERATORS_UNDER_TEST=$2
	echo "$OPERATORS_UNDER_TEST " | sed 's/ /,'"$OPERATOR_CATALOG"'\n/g' >"$OPERATOR_LIST_PATH"
else
	echo 'Wrong parameter count.
  Usage: ./run-basic-batch-operators-test.sh <catalog> ["<operator-name 1> <operator-name 2> ... <operator-name N>]
  Examples:
  ./run-basic-batch-operators-test.sh redhat-operators
  ./run-basic-batch-operators-test.sh redhat-operators "file-integrity-operator kiali-ossm"'
	exit 1
fi

# Check for docker config file
if [ ! -e "$DOCKER_CONFIG" ]; then
	echo_color "$RED" "Docker config is missing at $DOCKER_CONFIG"
	exit 1
fi

# Check KUBECONFIG
if [[ ! -v "KUBECONFIG" ]]; then
	echo_color "$RED" "The environment variable KUBECONFIG is not set."
	exit 1
fi

# Write config file template
cat <<EOF >"$CONFIG_YAML_TEMPLATE"
targetNameSpaces:
  - name: \$ns
podsUnderTestLabels:
  - "test-network-function.com/generic: target"
operatorsUnderTestLabels:
  - "test-network-function.com/operator: target" 
EOF

OPERATOR_PAGE='<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>HTTP Link Example</title>'

# Add per test run links
{
	# Add per operator details link
	echo "Time: <b>$TIMESTAMP</b>, catalog: <b>$OPERATOR_CATALOG</b>"

	#Add detailed results
	echo ", detailed results: "'<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$INDEX_FILE"'">'"link"'</a>'

	# Add CSV file link
	echo ", CSV: "
	echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/results.csv">'"link"'</a>'

	# Add operator list link
	echo ", operator list: "
	echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$OPERATOR_LIST_FILENAME"'">'"link"'</a>'

	# Add log link
	echo ", log: "
	echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$LOG_FILENAME"'">'"link"'</a>'

	# New line
	echo "<br>"
} >>"$BASE_DIR"/"$INDEX_FILE"

echo "$OPERATOR_PAGE" >>"$REPORT_FOLDER"/"$INDEX_FILE"

# Wait for the cluster to be reachable
echo_color "$BLUE" "Wait for cluster to be reachable"
wait_cluster_ok

echo_color "$BLUE" "Starting to install and test operators"
# For each operator in a provided catalog, this script will install the operator and run the CNF test suite.
while IFS=, read -r package_name catalog; do
	if [ "$package_name" = "" ]; then
		continue
	fi

	echo_color "$GREY" "********* package= $package_name catalog= $catalog **********"

	# Wait for the cluster to be reachable
	echo_color "$BLUE" "Wait for cluster to be reachable"
	wait_cluster_ok

	# Variable to hold return status
	status=0

	ns=$(get_suggested_namespace "$package_name")
	if [ "$ns" = "" ] || [ "$ns" = "openshift-operators" ]; then
		echo_color "$BLUE" "no suggested namespace for $package_name, using: test-operator"
		ns="test-operator"
	else
		echo_color "$BLUE" "using suggested namespace for $package_name: $ns "
	fi
	echo_color "$GREY" "namespace= $ns"

	echo_color "$BLUE" "Cluster cleanup"
	cleanup >>"$LOG_FILE_PATH" 2>&1 || status="$?"
	if [ "$status" != 0 ]; then
		echo_color "$RED" "Warning, cluster cleanup failed"
	fi

	# If a namespace is present, it is probably stuck deleting from previous runs. Force delete it.
	echo_color "$BLUE" "Remove namespace if present"
	force_delete_namespace_if_present "$ns" >>"$LOG_FILE_PATH" 2>&1 || status="$?"
	if [ "$status" != 0 ]; then
		echo_color "$RED" "Error, force deleting namespace failed"
	fi

	oc create namespace "$ns" || status="$?"
	if [ "$status" != 0 ]; then
		echo_color "$RED" "Error, creating namespace $ns failed"
	fi

	# Install the operator in a custom namespace
	echo_color "$BLUE" "install operator"
	oc operator install --create-operator-group "$package_name" -n "$ns" || status=$?
	if [ "$status" != 0 ]; then
		echo_color "$RED" "Operator installation failed but will still waiting for CSV"
	fi

	# Setting report directory
	report_dir="$REPORT_FOLDER"/"$package_name"

	# Store the results of CNF test in a new directory
	mkdir -p "$report_dir" || status="$?"
	if [ "$status" != 0 ]; then
		echo_color "$RED" "Error, creating report dir failed"
	fi

	config_yaml="$report_dir"/tnf_config.yml

	# Change the target_name_space in tnf_config file
	sed "s/\$ns/$ns/" "$CONFIG_YAML_TEMPLATE" >"$config_yaml"

	# Wait for the CSV to appear
	echo_color "$BLUE" "Wait for CSV to appear and label resources unde test"
	wait_for_csv_to_appear_and_label "$ns" || status="$?"
	if [ "$status" != 0 ]; then
		echo_color "$RED" "Operator failed to install, continue"
		report_failure "$status" "$ns" "$package_name" "Operator installation failed, skipping test"
		continue
	fi

	echo_color "$BLUE" "operator $package_name installed"

	# Label deployments, statefulsets and pods with "test-network-function.com/generic=target"
	{
		oc get deployment -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "  oc label " $3  " -n " $2 " " $1  " test-network-function.com/generic=target "}' | bash || true
		oc get statefulset -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "  oc label " $3  " -n " $2 " " $1  " test-network-function.com/generic=target "}' | bash || true
		oc get pods -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "  oc label " $3  " -n " $2 " " $1  " test-network-function.com/generic=target "}' | bash || true
	} >>"$LOG_FILE_PATH" 2>&1

	echo_color "$BLUE" "Wait to ensure all pods are running"
	# Extra wait to ensure that all pods are running
	sleep 30

	# run tnf-container
	echo_color "$BLUE" "run CNF suite"
	TNF_LOG_LEVEL=trace ./run-tnf-container.sh -k "$KUBECONFIG" -t "$report_dir" -o "$report_dir" -c "$DOCKER_CONFIG" -l all >>"$LOG_FILE_PATH" 2>&1 || {
		report_failure "$status" "$ns" "$package_name" "CNF suite exited with errors"
		continue
	}

	echo_color "$BLUE" "unlabel operator"
	# Unlabel the operator
	oc get csv -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "  oc label " $3  " -n " $2 " " $1  " test-network-function.com/operator- "}' | bash || status="$?"
	if [ "$status" != 0 ]; then
		echo_color "$RED" "Error, failed to unlabel the operator"
	fi

	# remove the operator
	echo_color "$BLUE" "Remove operator"
	oc operator uninstall -X "$package_name" -n "$ns" || status="$?"
	if [ "$status" != 0 ]; then
		echo_color "$RED" "Operator failed to un-install, continue"
	fi

	# Delete the namespace
	oc delete namespace "$ns" --wait=false || status="$?"
	if [ "$status" != 0 ]; then
		echo_color "$RED" "Error, failed to delete namespace: $ns"
	fi

	echo_color "$BLUE" "Wait for cleanup to finish"
	wait_delete_namespace "$ns" || status="$?"
	if [ "$status" != 0 ]; then
		echo_color "$RED" "Error, fail to wait for the namespace to be deleted"
	fi

	# Check parsing claim file
	echo_color "$BLUE" "Parse claim file"

	# merge claim.json from each operator to a single csv file
	echo_color "$BLUE" "add claim.json from this operator to the csv file"
	./tnf claim show csv -c "$report_dir"/claim.json -n "$package_name" -t "$CNF_TYPE" "$add_headers" >>"$REPORT_FOLDER"/results.csv || status="$?"
	if [ "$status" != 0 ]; then
		echo_color "$RED" "failed to parse claim file"
	fi

	# extract parser
	echo_color "$BLUE" "extract parser from report"
	with_retry 2 10 tar -xvf "$report_dir"/*.tar.gz -C "$report_dir" results.html || status="$?"
	if [ "$status" != 0 ]; then
		echo_color "$RED" "Failed get result.html from report"
	fi

	# Add per operator links
	{
		# Add parser link
		echo "Results for: <b>$package_name</b>,  parsed details:"
		echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$package_name"'/results.html?claimfile=/'"$REPORT_FOLDER_RELATIVE"'/'"$package_name"'/claim.json">'"link"'</a>'

		# Add log link
		echo ", log: "
		echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$package_name"'/tnf-execution.log">'"link"'</a>'

		# Add tnf_config link
		echo ", tnf_config: "
		echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$package_name"'/tnf_config.yml">'"link"'</a>'

		# new line
		echo "<br>"
	} >>"$REPORT_FOLDER"/"$INDEX_FILE"

	# Only print headers once
	add_headers=""

done <"$OPERATOR_LIST_PATH"

# closing html file
echo '</body></html>' >>"$REPORT_FOLDER"/"$INDEX_FILE"
echo_color "$GREEN" DONE
