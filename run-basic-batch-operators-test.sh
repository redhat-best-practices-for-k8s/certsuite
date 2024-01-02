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

# VARIABLES

# Variable to add header only on the first run
addHeaders=-a

# Create report directory
mkdir "$REPORT_FOLDER"

cleanup() {
	# Workaround for cleaning operator leftovers, see https://access.redhat.com/solutions/6971276
	oc delete mutatingwebhookconfigurations controller.devfile.io || true
	oc delete validatingwebhookconfigurations controller.devfile.io || true

	# cleanup any leftovers
	# https://docs.openshift.com/container-platform/4.14/operators/admin/olm-deleting-operators-from-cluster.html
	oc get csv -n openshift-operators | grep -v packageserver | grep -v NAME | awk '{print " oc delete --wait=true csv " $2 " -n openshift-operators"}' | bash || true
	oc get csv -A | grep -v packageserver | grep -v NAME | awk '{print " oc delete --wait=true csv " $2 " -n " $1}' | bash || true
	oc get subscriptions -A | grep -v NAME | awk '{print " oc delete --wait=true subscription " $2 " -n " $1}' | bash || true
	oc get job,configmap -n openshift-marketplace | grep -v NAME | grep -v "configmap/kube-root-ca.crt" | grep -v "configmap/marketplace-operator-lock" | grep -v "configmap/marketplace-trusted-ca" | grep -v "configmap/openshift-service-ca.crt" | awk '{print " oc delete --wait=true " $1 " -n openshift-marketplace" }' | bash || true
}

waitDeleteNamespace() {
	local namespaceDeleting=$1
	# Wait for the namespace to be removed
	if [ "$namespaceDeleting" != "openshift-operators" ]; then

		echo "non openshift-operators namespace = $namespaceDeleting, deleting "
		withRetry 2 10 oc wait namespace "$namespaceDeleting" --for=delete --timeout=60s || true
		forceDeleteNamespaceIfPresent "$namespaceDeleting"
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

			echo "command: $*" >&2
			echo "stderr: $stderr" >&2
			echo "stdout: $stdout" >&2
			echo "status: $status" >&2
			return 0
		fi
		# If the command failed, increment the retry counter
		retries=$((retries + 1))
		# shellcheck disable=SC2031
		echo "command: $*" >&2
		# shellcheck disable=SC2031
		echo "stderr: $stderr" >&2
		# shellcheck disable=SC2031
		echo "stdout: $stdout" >&2
		# shellcheck disable=SC2031
		echo "status: $status" >&2
		echo "Retry $retries/$maxRetries: Waiting for a few seconds before the next attempt..." >&2
		sleep "$timeout"
	done

	echo "Maximum retries reached. Exiting with failure." >&2
	return 1
}

waitClusterOk() {
	local \
		startTime \
		status \
		timeoutSeconds=600

	startTime=$(date +%s 2>&1) || {
		echo >&2 "date failed with error $?: $startTime"
		return 0
	}
	while true; do
		status=0
		oc get nodes &>/dev/null &&
			return 0 ||
			echo >&2 "get nodes failed with error $?."
		currentTime=$(date +%s)
		elapsedTime=$((currentTime - startTime))
		# If elapsed time is greater than the timeout report failure
		if [ "$elapsedTime" -ge "$timeoutSeconds" ]; then
			echo "Timeout reached $timeoutSeconds seconds waiting for cluster to be reachable."
			return 1
		fi

		# Otherwise wait a bit
		echo "Waiting for cluster to be reachable..."
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
		csvs=$(oc get csv -n "$csvNamespace") || true
		if [ "$csvs" != "" ]; then
			# If any CSV is present, break
			break
		else
			currentTime=$(date +%s)
			elapsedTime=$((currentTime - startTime))
			# If elapsed time is greater than the timeout report failure
			if [ "$elapsedTime" -ge "$timeoutSeconds" ]; then
				echo "Timeout reached $timeoutSeconds seconds waiting for CSV."
				return 1
			fi

			# Otherwise wait a bit
			echo "Waiting for csv to be created in namespace $csvNamespace ..."
			sleep 5
		fi
	done

	# Label CSV with "test-network-function.com/operator=target"
	command=$(withRetry 180 10 oc get csv -n "$csvNamespace" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | grep -v openshift-operator-lifecycle-manager | sed '/^ *$/d' | awk '{print "  withRetry 180 10 oc label " $3  " -n " $2 " " $1  " test-network-function.com/operator=target "}')
	eval "$command"

	# Wait for the CSV to be succeeded
	withRetry 2 10 oc wait csv -l test-network-function.com/operator=target -n "$ns" --for=jsonpath=\{.status.phase\}=Succeeded --timeout=60s || status="$?"
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
	withRetry 2 10 oc wait namespace "$aNamespace" --for=delete --timeout=5s || true

	# If a namespace with this name does not exist, all is good, exit
	if ! oc get namespace "$aNamespace"; then
		return 0
	fi

	# Otherwise force delete namespace
	withRetry 180 10 oc get namespace "$aNamespace" -ojson | sed '/"kubernetes"/d' >temp.yaml
	withRetry 180 10 oc proxy &
	pid=$!
	echo "PID: $pid"
	sleep 5
	curl -H "Content-Type: application/yaml" -X PUT --data-binary @temp.yaml http://127.0.0.1:8001/api/v1/namespaces/"$aNamespace"/finalize
	kill -9 "$pid"
	withRetry 2 10 oc wait namespace "$aNamespace" --for=delete --timeout=60s
}

# Main

# Check if the number of parameters is correct
if [ "$#" -eq 1 ]; then
	OPERATOR_CATALOG=$1
	# Get all the packages present in the cluster catalog
	withRetry 180 10 oc get packagemanifest -o jsonpath='{range .items[*]}{.metadata.name}{","}{.status.catalogSource}{"\n"}{end}' | grep "$OPERATOR_CATALOG" | head -n -1 >"$OPERATOR_LIST_PATH"

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
	echo "Docker config is missing at $DOCKER_CONFIG"
	exit 1
fi

# Check KUBECONFIG
if [[ ! -v "KUBECONFIG" ]]; then
	echo "The environment variable KUBECONFIG is not set."
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

	# New line
	echo "<br>"
} >>"$BASE_DIR"/"$INDEX_FILE"

echo "$OPERATOR_PAGE" >>"$REPORT_FOLDER"/"$INDEX_FILE"

# Wait for the cluster to be reachable
waitClusterOk

cleanup

# For each operator in a provided catalog, this script will install the operator and run the CNF test suite.
while IFS=, read -r packageName catalog; do
	if [ "$packageName" = "" ]; then
		continue
	fi

	echo "package=$packageName catalog=$catalog"

	# Wait for the cluster to be reachable
	waitClusterOk

	status=0
	withRetry 180 10 tasty install "$packageName" --source "$catalog" --stdout &>/dev/null || status=$?

	# if tasty fails, skip this operator
	if [ "$status" != 0 ]; then
		# Add per operator links
		{
			# Add error message
			echo "Results for: <b>$packageName</b>, "'<span style="color: red;">Operator installation failed due to tasty internal error, skipping test</span>'

			# Add tnf_config link
			echo ", tnf_config: "
			echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$packageName"'/tnf_config.yml">'"link"'</a>'

			# New line
			echo "<br>"
		} >>"$REPORT_FOLDER"/"$INDEX_FILE"

		cleanup

		continue
	fi
	# Wait for the cluster to be reachable
	waitClusterOk

	namesCount=$(withRetry 180 10 tasty install "$packageName" --source "$catalog" --stdout | grep -c "name:")

	if [ "$namesCount" = "4" ]; then
		# Get namespace from tasty
		ns=$(withRetry 180 10 tasty install "$packageName" --source "$catalog" --stdout | grep "name:" | head -n1 | awk '{ print $2 }')
	elif [ "$namesCount" = "2" ]; then
		ns="test-operators"
	fi

	echo "namespace=$ns"

	# If a namespace is present, it is probably stuck deleting from previous runs. Force delete it.
	forceDeleteNamespaceIfPresent "$ns"

	# Wait for the cluster to be reachable
	waitClusterOk

	# Install the operator in a custom namespace
	withRetry 180 10 tasty install "$packageName" --source "$catalog" -w -n "$ns" --stdout >operator.yml
	if [ "$ns" = "test-operators" ]; then
		sed -i '/targetNamespaces:/ { N; /- test-operators/d }' operator.yml
	fi

	# apply namespace/operator group and subscription
	withRetry 180 10 oc apply -f operator.yml

	# Wait for the cluster to be reachable
	waitClusterOk

	# Setting report directory
	reportDir="$REPORT_FOLDER"/"$packageName"

	# Store the results of CNF test in a new directory
	mkdir -p "$reportDir"

	configYaml="$reportDir"/tnf_config.yml

	# Change the targetNameSpace in tnf_config file
	sed "s/\$ns/$ns/" "$CONFIG_YAML_TEMPLATE" >"$configYaml"
	status=0
	# Wait for the CSV to appear
	waitForCsvToAppearAndLabel "$ns" || status="$?"

	if [ "$status" != 0 ]; then
		# Add per operator links
		{
			# Add error message
			echo "Results for: <b>$packageName</b>, "'<span style="color: red;">Operator installation failed, skipping test</span>'

			# Add tnf_config link
			echo ", tnf_config: "
			echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$packageName"'/tnf_config.yml">'"link"'</a>'

			# New line
			echo "<br>"
		} >>"$REPORT_FOLDER"/"$INDEX_FILE"

		# Remove the operator
		withRetry 180 10 oc delete -f operator.yml

		cleanup
		waitDeleteNamespace "$ns"

		continue
	fi

	echo "operator $packageName installed"

	# Label deployments, statefulsets and pods with "test-network-function.com/generic=target"
	oc get deployment -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "  oc label " $3  " -n " $2 " " $1  " test-network-function.com/generic=target "}' | bash || true
	oc get statefulset -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "  oc label " $3  " -n " $2 " " $1  " test-network-function.com/generic=target "}' | bash || true
	oc get pods -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "  oc label " $3  " -n " $2 " " $1  " test-network-function.com/generic=target "}' | bash || true

	# Extra wait to ensure that all pods are running
	sleep 30

	# run tnf-container
	TNF_LOG_LEVEL=trace ./run-tnf-container.sh -k "$KUBECONFIG" -t "$reportDir" -o "$reportDir" -c "$DOCKER_CONFIG" -l all || true

	# Unlabel the operator
	oc get csv -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "  oc label " $3  " -n " $2 " " $1  " test-network-function.com/operator- "}' | bash || true

	# remove the operator
	withRetry 180 10 oc delete -f operator.yml --wait=false

	waitDeleteNamespace "$ns"
	cleanup

	# Check parsing claim file
	./tnf claim show csv -c "$reportDir"/claim.json -n "$packageName" -t "$CNF_TYPE" "$addHeaders" || {

		# if parsing claim file fails, skip this operator
		# Add per operator links
		{
			# Add error message
			echo "Results for: <b>$packageName</b>, "'<span style="color: red;">Operator installation failed due to claim parsing error, skipping test</span>'

			# Add tnf_config link
			echo ", tnf_config: "
			echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$packageName"'/tnf_config.yml">'"link"'</a>'

			# New line
			echo "<br>"
		} >>"$REPORT_FOLDER/$INDEX_FILE"

		cleanup

		continue
	}

	# merge claim.json from each operator to a single csv file
	./tnf claim show csv -c "$reportDir"/claim.json -n "$packageName" -t "$CNF_TYPE" "$addHeaders" >>"$REPORT_FOLDER"/results.csv

	# extract parser
	tar -xvf "$reportDir"/*.tar.gz -C "$reportDir" results.html

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

# Resetting project to default
withRetry 180 10 oc project default

# closing html file
echo '</body></html>' >>"$REPORT_FOLDER"/"$INDEX_FILE"
echo "DONE"
