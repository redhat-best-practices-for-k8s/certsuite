#!/bin/sh

# Test run timestamp
TIMESTAMP=$(date | sed 's/[ :]/_/g')

# base folder
BASE_DIR=/var/www/html

# INPUTS

# config.yaml template file path
CONFIG_YAML_TEMPLATE=$(pwd)/tnf_config.yml.template

# Docker config used to pull operator images
DOCKER_CONFIG=config.json

# Location of telco/non-telco classification file
CNF_TYPE=cmd/tnf/claim/show/csv/cnf-type.json

# operator bundle list from user
ORIG_BUNDLE_PATH=$1

# OUTPUTS

# report folder
REPORT_FOLDER_RELATIVE="report_$TIMESTAMP"

# Report results folder
REPORT_FOLDER="$BASE_DIR"/"$REPORT_FOLDER_RELATIVE"

# bundle file name
BUNDLE_FILENAME=bundelist.txt

# bundle path in the report
BUNDLE_PATH="$REPORT_FOLDER"/"$BUNDLE_FILENAME"

# VARIABLES

# variable to add header only on the first run
addHeaders=-a

getLatestBundle() {
	PACKAGE_NAME=$1
	jsonData=$(grpcurl -plaintext localhost:50051 api.Registry.ListBundles | jq --arg packagename "$PACKAGE_NAME" '. | select(.packageName == $packagename) | del(.csvJson) | del(.object)')

	latest=$(echo "$jsonData" | jq -r '.version' | sort -V | tail -n1)

	# retrieve the bundlePath for the latest version
	bundlePath=$(echo "$jsonData" | jq -r --arg latest "$latest" '. | select(.version == $latest) | .bundlePath' | tail -n1)

	echo "$PACKAGE_NAME", "$bundlePath"
}

# create report directory
mkdir "$REPORT_FOLDER"

# Check if the number of parameters is correct
if [ "$#" -eq 0 ]; then
	echo "missing operator bundle list file, getting redhat operators from catalog"
	echo "results will be in $BASE_DIR directory"
	oc -n openshift-marketplace port-forward service/redhat-operators 50051:50051 &
	background_pid=$!
	echo "Background PID: $background_pid"

	# get all the packages present in the cluster catalog
	oc get packagemanifest | grep "Red Hat Operators" | awk '{print $1}' >"$REPORT_FOLDER"/redhat-operators.txt

	# get BundleImage for each of the packages
	while read -r i; do getLatestBundle "$i"; done <"$REPORT_FOLDER"/redhat-operators.txt >"$BUNDLE_PATH"

	# kill the port forwarding
	kill -9 "$background_pid"
	ORIG_BUNDLE_PATH=redhat-operators
else
	# if user provided file is present, copy it in the report
	cp "$ORIG_BUNDLE_PATH" "$BUNDLE_PATH"
fi

# check for docker config file
if [ ! -e "$DOCKER_CONFIG" ]; then
	echo "Docker config is missing at $DOCKER_CONFIG"
	exit 1
fi

# check KUBECONFIG
if [ -z "$KUBECONFIG" ]; then
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

# add per test run links
{
	# add per operator details link
	echo "Time: <b>$TIMESTAMP</b>, file: <b>$ORIG_BUNDLE_PATH</b>"

	#add detailed results
	echo ", detailed results: "'<a href="/'"$REPORT_FOLDER_RELATIVE"'/index.html">'"link"'</a>'

	# add CSV file link
	echo ", CSV: "
	echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/results.csv">'"link"'</a>'

	# add operator bundle list link
	echo ", bundle list: "
	echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$BUNDLE_FILENAME"'">'"link"'</a>'

	# new line
	echo "<br>"
} >>"$BASE_DIR"/index.html

echo "$OPERATOR_PAGE" >>"$REPORT_FOLDER"/index.html
# For each bundle in a provided catalog, this script will install the operator and run the CNF test suite.
while IFS=, read -r package_name bundle_image; do
	# Workaround for cleaning operator leftovers, see https://access.redhat.com/solutions/6971276
	oc delete mutatingwebhookconfigurations controller.devfile.io
	oc delete validatingwebhookconfigurations controller.devfile.io

	# read package name and bundle image from the text file
	package_name=$(echo "$package_name" | awk '{$1=$1};1')
	bundle_image=$(echo "$bundle_image" | awk '{$1=$1};1')

	# create a new namespace for each operator install
	ns=test-"$package_name"
	oc new-project "$ns"

	# use operator-sdk binary to install the operator in a custom namespace
	operator-sdk run bundle "$bundle_image"

	reportDir="$REPORT_FOLDER"/"$package_name"

	# store the results of CNF test in a new directory
	mkdir -p "$reportDir"

	configYaml="$reportDir"/tnf_config.yml

	# change the targetNameSpace in tng_config file
	sed "s/\$ns/$ns/" "$CONFIG_YAML_TEMPLATE" >"$configYaml"

	# label everything
	oc get deployment -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "oc label " $3  " -n " $2 " " $1  " test-network-function.com/generic=target "}' | bash
	oc get statefulset -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "oc label " $3  " -n " $2 " " $1  " test-network-function.com/generic=target "}' | bash
	oc get pods -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "oc label " $3  " -n " $2 " " $1  " test-network-function.com/generic=target "}' | bash
	oc get csv -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "oc label " $3  " -n " $2 " " $1  " test-network-function.com/operator=target "}' | bash
	oc get operator -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "oc label " $3  " " $1  " test-network-function.com/operator=target "}' | bash

	# run tnf-container
	./run-tnf-container.sh -k "$KUBECONFIG" -t "$reportDir" -o "$reportDir" -c "$DOCKER_CONFIG" -l all

	# unlabel and uninstall the operator
	oc get csv -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "oc label " $3  " -n " $2 " " $1  " test-network-function.com/operator- "}' | bash

	operator-sdk cleanup "$package_name"

	# delete the namespace
	oc delete ns "$ns"

	# merge claim.json from each operator to a single csv file
	./tnf claim show csv -c "$reportDir"/claim.json -n "$package_name" -t "$CNF_TYPE" "$addHeaders" >>"$REPORT_FOLDER"/results.csv

	# add per operator links
	{
		# add parser link
		echo "Results for: <b>$package_name</b>,  parsed details:"
		echo '<a href="/results.html?claimfile=/'"$REPORT_FOLDER_RELATIVE"'/'"$package_name"'/claim.json">'"link"'</a>'

		# add log link
		echo ", log: "
		echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$package_name"'/tnf-execution.log">'"link"'</a>'

		# add tnf_config link
		echo ", tnf_config: "
		echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$package_name"'/tnf_config.yml">'"link"'</a>'

		# new line
		echo "<br>"
	} >>"$REPORT_FOLDER"/index.html

	# Only print headers once
	addHeaders=""

done <"$BUNDLE_PATH"

# Workaround for cleaning operator leftovers, see https://access.redhat.com/solutions/6971276
oc delete mutatingwebhookconfigurations controller.devfile.io
oc delete validatingwebhookconfigurations controller.devfile.io

# Resetting project to default
oc project default

# closing html file
echo '</body></html>' >>"$REPORT_FOLDER"/index.html
