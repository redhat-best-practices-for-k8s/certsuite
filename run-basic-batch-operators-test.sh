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

# Report folder
REPORT_FOLDER_RELATIVE="report_$TIMESTAMP"

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
	oc get csv -n openshift-operators | grep -v packageserver | grep -v NAME | awk '{print "oc delete --wait=true csv " $2 " -n openshift-operators"}' | bash || true
	oc get csv -A | grep -v packageserver | grep -v NAME | awk '{print "oc delete --wait=true csv " $2 " -n " $1}' | bash || true
	oc get subscriptions -A | grep -v NAME | awk '{print "oc delete --wait=true subscription " $2 " -n " $1}' | bash || true
	oc get job,configmap -n openshift-marketplace | grep -v NAME | grep -v "configmap/kube-root-ca.crt" | grep -v "configmap/marketplace-operator-lock" | grep -v "configmap/marketplace-trusted-ca" | grep -v "configmap/openshift-service-ca.crt" | awk '{print "oc delete --wait=true " $1 " -n openshift-marketplace" }' | bash || true
}

waitDeleteNamespace() {
	namespaceDeleting=$1
	# Wait for the CSV to be removed
	oc wait csv -l test-network-function.com/operator=target -n "$namespaceDeleting" --for=delete --timeout=300s || true

	# Wait for the namespace to be removed
	if [ "$namespaceDeleting" != "openshift-operators" ]; then

		echo "non openshift-operators namespace = $namespaceDeleting, deleting "
		oc wait namespace "$namespaceDeleting" --for=delete --timeout=300s || true
		forceDeleteNamespaceIfPresent "$namespaceDeleting"
	fi
}

waitForCsvToAppearAndLabel() {
	csvNamespace=$1
	timeoutSeconds=300
	startTime=$(date +%s)
	while true; do
		csvs=$(oc get csv -n "$csvNamespace")
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
	oc get csv -n "$csvNamespace" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | grep -v openshift-operator-lifecycle-manager | sed '/^ *$/d' | awk '{print "oc label " $3  " -n " $2 " " $1  " test-network-function.com/operator=target "}' | bash

	# Wait for the CSV to be succeeded
	status=0
	oc wait csv -l test-network-function.com/operator=target -n "$ns" --for=jsonpath=\{.status.phase\}=Succeeded --timeout=300s || status="$?"
	return $status
}

forceDeleteNamespaceIfPresent() {
	aNamespace=$1

	# Do not delete the redhat-operators namespace
	if [ "$aNamespace" = "openshift-operators" ]; then
		return 0
	fi
	# Delete namespace
	oc delete namespace "$aNamespace" --wait=false || true
	oc wait namespace "$aNamespace" --for=delete --timeout=30s || true

	# If a namespace with this name does not exist, all is good, exit
	if ! oc get namespace "$aNamespace"; then
		return 0
	fi

	# Otherwise force delete namespace
	oc get namespace "$aNamespace" -ojson | sed '/"kubernetes"/d' >temp.yaml
	oc proxy &
	pid=$!
	echo "PID: $pid"
	sleep 5
	curl -H "Content-Type: application/yaml" -X PUT --data-binary @temp.yaml http://127.0.0.1:8001/api/v1/namespaces/"$aNamespace"/finalize
	kill -9 "$pid"
	oc wait namespace "$aNamespace" --for=delete --timeout=300s || true
}

# Check if the number of parameters is correct
if [ "$#" -eq 1 ]; then
	OPERATOR_CATALOG=$1
	# Get all the packages present in the cluster catalog
	oc get packagemanifest -o jsonpath='{range .items[*]}{.metadata.name}{","}{.status.catalogSource}{"\n"}{end}' | grep "$OPERATOR_CATALOG" | head -n -1 >"$OPERATOR_LIST_PATH"

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

cleanup

# For each operator in a provided catalog, this script will install the operator and run the CNF test suite.
while IFS=, read -r package_name catalog; do
	if [ "$package_name" = "" ]; then
		continue
	fi

	echo "package=$package_name catalog=$catalog"

	status=0
	tasty install "$package_name" --source "$catalog" --stdout &>/dev/null || status=$?

	# if tasty fails, skip this operator
	if [ "$status" != 0 ]; then
		# Add per operator links
		{
			# Add error message
			echo "Results for: <b>$package_name</b>, "'<span style="color: red;">Operator installation failed due to tasty internal error, skipping test</span>'

			# Add tnf_config link
			echo ", tnf_config: "
			echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$package_name"'/tnf_config.yml">'"link"'</a>'

			# New line
			echo "<br>"
		} >>"$REPORT_FOLDER"/"$INDEX_FILE"

		cleanup

		continue
	fi

	namesCount=$(tasty install "$package_name" --source "$catalog" --stdout | grep -c "name:")

	if [ "$namesCount" = "4" ]; then
		# Get namespace from tasty
		ns=$(tasty install "$package_name" --source "$catalog" --stdout | grep "name:" | head -n1 | awk '{ print $2 }')
	elif [ "$namesCount" = "2" ]; then
		ns="openshift-operators"
	fi

	echo "namespace=$ns"

	# If a namespace is present, it is probably stuck deleting from previous runs. Force delete it.
	forceDeleteNamespaceIfPresent "$ns"

	# Install the operator in a custom namespace
	tasty install "$package_name" --source "$catalog" -w

	# Setting report directory
	reportDir="$REPORT_FOLDER"/"$package_name"

	# Store the results of CNF test in a new directory
	mkdir -p "$reportDir"

	configYaml="$reportDir"/tnf_config.yml

	# Change the targetNameSpace in tng_config file
	sed "s/\$ns/$ns/" "$CONFIG_YAML_TEMPLATE" >"$configYaml"
	status=0
	# Wait for the CSV to appear
	waitForCsvToAppearAndLabel "$ns" || status="$?"

	if [ "$status" != 0 ]; then
		# Add per operator links
		{
			# Add error message
			echo "Results for: <b>$package_name</b>, "'<span style="color: red;">Operator installation failed, skipping test</span>'

			# Add tnf_config link
			echo ", tnf_config: "
			echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$package_name"'/tnf_config.yml">'"link"'</a>'

			# New line
			echo "<br>"
		} >>"$REPORT_FOLDER"/"$INDEX_FILE"
		# Remove the operator
		tasty remove "$package_name"

		cleanup
		waitDeleteNamespace "$ns"

		continue
	fi

	echo "operator $package_name installed"

	# Label deployments, statefulsets and pods with "test-network-function.com/generic=target"
	oc get deployment -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "oc label " $3  " -n " $2 " " $1  " test-network-function.com/generic=target "}' | bash
	oc get statefulset -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "oc label " $3  " -n " $2 " " $1  " test-network-function.com/generic=target "}' | bash
	oc get pods -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "oc label " $3  " -n " $2 " " $1  " test-network-function.com/generic=target "}' | bash

	# run tnf-container
	./run-tnf-container.sh -k "$KUBECONFIG" -t "$reportDir" -o "$reportDir" -c "$DOCKER_CONFIG" -l all || true

	# Unlabel and uninstall the operator
	oc get csv -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "oc label " $3  " -n " $2 " " $1  " test-network-function.com/operator- "}' | bash

	# remove the operator
	tasty remove "$package_name"

	cleanup
	waitDeleteNamespace "$ns"

	# merge claim.json from each operator to a single csv file
	./tnf claim show csv -c "$reportDir"/claim.json -n "$package_name" -t "$CNF_TYPE" "$addHeaders" >>"$REPORT_FOLDER"/results.csv

	# Add per operator links
	{
		# Add parser link
		echo "Results for: <b>$package_name</b>,  parsed details:"
		echo '<a href="/results.html?claimfile=/'"$REPORT_FOLDER_RELATIVE"'/'"$package_name"'/claim.json">'"link"'</a>'

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
	addHeaders=""

done <"$OPERATOR_LIST_PATH"

# Resetting project to default
oc project default

# closing html file
echo '</body></html>' >>"$REPORT_FOLDER"/"$INDEX_FILE"
echo "DONE"
