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

echoColor() {
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
addHeaders=-a

# Create report directory
mkdir "$REPORT_FOLDER"

cleanup() {
	# Workaround for cleaning operator leftovers, see https://access.redhat.com/solutions/6971276
	oc delete mutatingwebhookconfigurations controller.devfile.io || true
	oc delete validatingwebhookconfigurations controller.devfile.io || true

	# Leftovers specific to certain operators
	oc delete ValidatingWebhookConfiguration sriov-operator-webhook-config || true
	oc delete MutatingWebhookConfiguration sriov-operator-webhook-config || true
}

waitDeleteNamespace() {
	local namespaceDeleting=$1
	# Wait for the namespace to be removed
	if [ "$namespaceDeleting" != "openshift-operators" ]; then

		echoColor "$BLUE" "non openshift-operators namespace = $namespaceDeleting, deleting "
		withRetry 2 0 oc wait namespace "$namespaceDeleting" --for=delete --timeout=60s || true

		forceDeleteNamespaceIfPresent "$namespaceDeleting" >>"$LOG_FILE_PATH" 2>&1 || true
	fi
}

# Executes command with retry
withRetry() {
	local \
		maxRetries=$1 \
		timeout=$2 \
		retries=0 \
		status=0 \
		stderr='' \
		stdout=''
	shift 2

	until [ "$retries" -ge "$maxRetries" ]; do
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

		echoColor "$GREY" "Retry $retries/$maxRetries: Waiting for a few seconds before the next attempt..."
		sleep "$timeout"
	done
	echoColor "$GREY" "Maximum retries reached."
	return 1
}

waitClusterOk() {
	local \
		startTime \
		status \
		timeoutSeconds=600

	startTime=$(date +%s 2>&1) || {
		echo "date failed with error $?: $startTime" >>"$LOG_FILE_PATH"
		return 0
	}
	while true; do
		status=0
		oc get nodes &>/dev/null &&
			return 0 ||
			echo "get nodes failed with error $?." >>"$LOG_FILE_PATH"
		currentTime=$(date +%s)
		elapsedTime=$((currentTime - startTime))
		# If elapsed time is greater than the timeout report failure
		if [ "$elapsedTime" -ge "$timeoutSeconds" ]; then
			echoColor "$BLUE" "Timeout reached $timeoutSeconds seconds waiting for cluster to be reachable."
			return 1
		fi

		# Otherwise wait a bit
		echoColor "$BLUE" "Waiting for cluster to be reachable..."
		sleep 5
	done
}

waitForCsvToAppearAndLabel() {
	local csvNamespace=$1
	local timeoutSeconds=100
	local startTime=0
	local currentTime=0
	local elapsedTime=0
	local command=""
	local status=0

	startTime=$(date +%s)
	while true; do
		csvs=$(oc get csv -n "$csvNamespace" 2>>"$LOG_FILE_PATH") || true
		if [ "$csvs" != "" ]; then
			# If any CSV is present, break
			break
		else
			currentTime=$(date +%s)
			elapsedTime=$((currentTime - startTime))
			# If elapsed time is greater than the timeout report failure
			if [ "$elapsedTime" -ge "$timeoutSeconds" ]; then
				echoColor "$BLUE" "Timeout reached $timeoutSeconds seconds waiting for CSV."
				return 1
			fi

			# Otherwise wait a bit
			echoColor "$BLUE" "Waiting for csv to be created in namespace $csvNamespace ..."
			sleep 5
		fi
	done

	# Label CSV with "test-network-function.com/operator=target"
	command=$(withRetry 5 10 oc get csv -n "$csvNamespace" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | grep -v openshift-operator-lifecycle-manager | sed '/^ *$/d' | awk '{print "  withRetry 5 10 oc label " $3  " -n " $2 " " $1  " test-network-function.com/operator=target "}')
	eval "$command"

	# Wait for the CSV to be succeeded
	echoColor "$BLUE" "Wait for CSV to be succeeded"
	withRetry 30 0 oc wait csv -l test-network-function.com/operator=target -n "$ns" --for=jsonpath=\{.status.phase\}=Succeeded --timeout=5s || status="$?"
	return $status
}

forceDeleteNamespaceIfPresent() {
	local aNamespace=$1
	local pid=0

	# Do not delete the redhat-operators namespace
	if [ "$aNamespace" = "openshift-operators" ]; then
		return 0
	fi

	# Delete namespace
	oc delete namespace "$aNamespace" --wait=false || true
	withRetry 2 0 oc wait namespace "$aNamespace" --for=delete --timeout=5s || true

	# If a namespace with this name does not exist, all is good, exit
	if ! oc get namespace "$aNamespace"; then
		return 0
	fi
	echoColor "$RED" "Namespace cannot be deleted normally, force deleting"
	# Otherwise force delete namespace
	withRetry 5 10 oc get namespace "$aNamespace" -ojson | sed '/"kubernetes"/d' >temp.yaml
	# Kill previous oc proxy command in the background
	killall "oc"
	# Start a new proxy
	oc proxy &
	pid=$!
	echo "PID: $pid"
	sleep 5
	curl -H "Content-Type: application/yaml" -X PUT --data-binary @temp.yaml http://127.0.0.1:8001/api/v1/namespaces/"$aNamespace"/finalize >>"$LOG_FILE_PATH"
	kill -9 "$pid"
	withRetry 2 0 oc wait namespace "$aNamespace" --for=delete --timeout=60s
}

reportFailure() {
	local status=$1
	local ns=$2
	local packageName=$3
	local message=$4

	withRetry 3 5 oc operator uninstall -X "$packageName" -n "$ns" || true
	# Add per operator links
	{
		# Add error message
		echo "Results for: <b>$packageName</b>, "'<span style="color: red;">'"$message"'</span>'

		# Add tnf_config link
		echo ", tnf_config: "
		echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$packageName"'/tnf_config.yml">'"link"'</a>'

		# New line
		echo "<br>"
	} >>"$REPORT_FOLDER"/"$INDEX_FILE"

	waitDeleteNamespace "$ns"
}

getSuggestedNamespace() {
	local packageName=$1

	oc get packagemanifests -n openshift-marketplace "$packageName" -ojson | jq -r '.status.channels[].currentCSVDesc.annotations."operatorframework.io/suggested-namespace"' 2>/dev/null | grep -v "null" | sed 's/\n//g' | head -1 || true
}

# Main

# Check if the number of parameters is correct
if [ "$#" -eq 1 ]; then
	OPERATOR_CATALOG=$1
	# Get all the packages present in the cluster catalog
	withRetry 5 10 oc get packagemanifest -o jsonpath='{range .items[*]}{.metadata.name}{","}{.status.catalogSource}{"\n"}{end}' | grep "$OPERATOR_CATALOG" | head -n -1 | sort >"$OPERATOR_LIST_PATH"

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
	echoColor "$RED" "Docker config is missing at $DOCKER_CONFIG"
	exit 1
fi

# Check KUBECONFIG
if [[ ! -v "KUBECONFIG" ]]; then
	echoColor "$RED" "The environment variable KUBECONFIG is not set."
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
echoColor "$BLUE" "Wait for cluster to be reachable"
waitClusterOk

echoColor "$BLUE" "Starting to install and test operators"
# For each operator in a provided catalog, this script will install the operator and run the CNF test suite.
while IFS=, read -r packageName catalog; do
	if [ "$packageName" = "" ]; then
		continue
	fi

	echoColor "$GREY" "********* package= $packageName catalog= $catalog **********"

	# Wait for the cluster to be reachable
	echoColor "$BLUE" "Wait for cluster to be reachable"
	waitClusterOk

	# Variable to hold return status
	status=0

	ns=$(getSuggestedNamespace "$packageName")
	if [ "$ns" = "" ] || [ "$ns" = "openshift-operators" ]; then
		echoColor "$BLUE" "no suggested namespace for $packageName, using: test-operator"
		ns="test-operator"
	else
		echoColor "$BLUE" "using suggested namespace for $packageName: $ns "
	fi
	echoColor "$GREY" "namespace= $ns"

	echoColor "$BLUE" "Cluster cleanup"
	cleanup >>"$LOG_FILE_PATH" 2>&1 || status="$?"
	if [ "$status" != 0 ]; then
		echoColor "$RED" "Warning, cluster cleanup failed"
	fi

	# If a namespace is present, it is probably stuck deleting from previous runs. Force delete it.
	echoColor "$BLUE" "Remove namespace if present"
	forceDeleteNamespaceIfPresent "$ns" >>"$LOG_FILE_PATH" 2>&1 || status="$?"
	if [ "$status" != 0 ]; then
		echoColor "$RED" "Error, force deleting namespace failed"
	fi

	oc create namespace "$ns" || status="$?"
	if [ "$status" != 0 ]; then
		echoColor "$RED" "Error, creating namespace $ns failed"
	fi

	# Install the operator in a custom namespace
	echoColor "$BLUE" "install operator"
	oc operator install --create-operator-group "$packageName" -n "$ns" || status=$?
	if [ "$status" != 0 ]; then
		echoColor "$RED" "Operator installation failed but will still waiting for CSV"
	fi

	# Setting report directory
	reportDir="$REPORT_FOLDER"/"$packageName"

	# Store the results of CNF test in a new directory
	mkdir -p "$reportDir" || status="$?"
	if [ "$status" != 0 ]; then
		echoColor "$RED" "Error, creating report dir failed"
	fi

	configYaml="$reportDir"/tnf_config.yml

	# Change the targetNameSpace in tnf_config file
	sed "s/\$ns/$ns/" "$CONFIG_YAML_TEMPLATE" >"$configYaml"

	# Wait for the CSV to appear
	echoColor "$BLUE" "Wait for CSV to appear and label resources unde test"
	waitForCsvToAppearAndLabel "$ns" || status="$?"
	if [ "$status" != 0 ]; then
		echoColor "$RED" "Operator failed to install, continue"
		reportFailure "$status" "$ns" "$packageName" "Operator installation failed, skipping test"
		continue
	fi

	echoColor "$BLUE" "operator $packageName installed"

	# Label deployments, statefulsets and pods with "test-network-function.com/generic=target"
	{
		oc get deployment -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "  oc label " $3  " -n " $2 " " $1  " test-network-function.com/generic=target "}' | bash || true
		oc get statefulset -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "  oc label " $3  " -n " $2 " " $1  " test-network-function.com/generic=target "}' | bash || true
		oc get pods -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "  oc label " $3  " -n " $2 " " $1  " test-network-function.com/generic=target "}' | bash || true
	} >>"$LOG_FILE_PATH" 2>&1

	echoColor "$BLUE" "Wait to ensure all pods are running"
	# Extra wait to ensure that all pods are running
	sleep 30

	# run tnf-container
	echoColor "$BLUE" "run CNF suite"
	TNF_LOG_LEVEL=trace ./run-tnf-container.sh -k "$KUBECONFIG" -t "$reportDir" -o "$reportDir" -c "$DOCKER_CONFIG" -l all >>"$LOG_FILE_PATH" 2>&1 || {
		reportFailure "$status" "$ns" "$packageName" "CNF suite exited with errors"
		continue
	}

	echoColor "$BLUE" "unlabel operator"
	# Unlabel the operator
	oc get csv -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "  oc label " $3  " -n " $2 " " $1  " test-network-function.com/operator- "}' | bash || status="$?"
	if [ "$status" != 0 ]; then
		echoColor "$RED" "Error, failed to unlabel the operator"
	fi

	# remove the operator
	echoColor "$BLUE" "Remove operator"
	oc operator uninstall -X "$packageName" -n "$ns" || status="$?"
	if [ "$status" != 0 ]; then
		echoColor "$RED" "Operator failed to un-install, continue"
	fi

	# Delete the namespace
	oc delete namespace "$ns" --wait=false || status="$?"
	if [ "$status" != 0 ]; then
		echoColor "$RED" "Error, failed to delete namespace: $ns"
	fi

	echoColor "$BLUE" "Wait for cleanup to finish"
	waitDeleteNamespace "$ns" || status="$?"
	if [ "$status" != 0 ]; then
		echoColor "$RED" "Error, fail to wait for the namespace to be deleted"
	fi

	# Check parsing claim file
	echoColor "$BLUE" "Parse claim file"

	# merge claim.json from each operator to a single csv file
	echoColor "$BLUE" "add claim.json from this operator to the csv file"
	./tnf claim show csv -c "$reportDir"/claim.json -n "$packageName" -t "$CNF_TYPE" "$addHeaders" >>"$REPORT_FOLDER"/results.csv || status="$?"
	if [ "$status" != 0 ]; then
		echoColor "$RED" "failed to parse claim file"
	fi

	# extract parser
	echoColor "$BLUE" "extract parser from report"
	withRetry 2 10 tar -xvf "$reportDir"/*.tar.gz -C "$reportDir" results.html || status="$?"
	if [ "$status" != 0 ]; then
		echoColor "$RED" "Failed get result.html from report"
	fi

	# Add per operator links
	{
		# Add parser link
		echo "Results for: <b>$packageName</b>,  parsed details:"
		echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$packageName"'/results.html?claimfile=/'"$REPORT_FOLDER_RELATIVE"'/'"$packageName"'/claim.json">'"link"'</a>'

		# Add log link
		echo ", log: "
		echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$packageName"'/tnf-execution.log">'"link"'</a>'

		# Add tnf_config link
		echo ", tnf_config: "
		echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$packageName"'/tnf_config.yml">'"link"'</a>'

		# new line
		echo "<br>"
	} >>"$REPORT_FOLDER"/"$INDEX_FILE"

	# Only print headers once
	addHeaders=""

done <"$OPERATOR_LIST_PATH"

# closing html file
echo '</body></html>' >>"$REPORT_FOLDER"/"$INDEX_FILE"
echoColor "$GREEN" DONE
