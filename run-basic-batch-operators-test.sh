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

# operator catalog from user
OPERATOR_CATALOG=""

# operator from user
OPERATORS_UNDER_TEST=""

# OUTPUTS

# report folder
REPORT_FOLDER_RELATIVE="report_$TIMESTAMP"

# Report results folder
REPORT_FOLDER="$BASE_DIR"/"$REPORT_FOLDER_RELATIVE"

# operator  file name
OPERATOR_LIST_FILENAME=operator-list.txt

# operator list path in the report
OPERATOR_LIST_PATH="$REPORT_FOLDER"/"$OPERATOR_LIST_FILENAME"

# VARIABLES

# variable to add header only on the first run
addHeaders=-a

# create report directory
mkdir "$REPORT_FOLDER"

cleanup() {
	# cleanup any leftovers
	# https://docs.openshift.com/container-platform/4.14/operators/admin/olm-deleting-operators-from-cluster.html
	oc get csv -n openshift-operators | grep -v packageserver | grep -v NAME | awk '{print "oc delete --wait=true csv " $2 " -n openshift-operators"}' | bash
	oc get csv -A | grep -v packageserver | grep -v NAME | awk '{print "oc delete --wait=true csv " $2 " -n " $1}' | bash
	oc get subscriptions -A | grep -v NAME | awk '{print "oc delete --wait=true subscription " $2 " -n " $1}' | bash
	oc get job,configmap -n openshift-marketplace | grep -v NAME | grep -v "configmap/kube-root-ca.crt" | grep -v "configmap/marketplace-operator-lock" | grep -v "configmap/marketplace-trusted-ca" | grep -v "configmap/openshift-service-ca.crt" | awk '{print "oc delete --wait=true " $1 " -n openshift-marketplace" }' | bash
}

waitDeleteNamespace() {
	namespaceDeleting=$1
	# wait for the CSV to be removed
	oc wait csv -l test-network-function.com/operator=target -n "$namespaceDeleting" --for=delete --timeout=300s

	# wait for the namespace to be removed
	if [ "$namespaceDeleting" != "openshift-operators" ]; then

		echo "non openshift-operators namespace = $namespaceDeleting, deleting "
		oc wait namespace "$namespaceDeleting" --for=delete --timeout=300s
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
			# if any CSV is present, break
			break
		else
			currentTime=$(date +%s)
			elapsedTime=$((currentTime - startTime))
			# if elapsed time is greater than the timeout report failure
			if [ "$elapsedTime" -ge "$timeoutSeconds" ]; then
				echo "Timeout reached $timeoutSeconds seconds waiting for CSV."
				return 1
			fi

			# otherwise wait a bit
			echo "Waiting for csv to be created in namespace $csvNamespace ..."
			sleep 5
		fi
	done

	# label CSV
	oc get csv -n "$csvNamespace" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | grep -v openshift-operator-lifecycle-manager | sed '/^ *$/d' | awk '{print "oc label " $3  " -n " $2 " " $1  " test-network-function.com/operator=target "}' | bash

	# wait for the CSV to be succeeded
	oc wait csv -l test-network-function.com/operator=target -n "$ns" --for=jsonpath=\{.status.phase\}=Succeeded --timeout=300s
}

forceDeleteNamespaceIfPresent() {
	aNamespace=$1

	# do not delete the redhat-operators namespace
	if [ "$aNamespace" = "openshift-operators" ]; then
		return 0
	fi
	# delete namespace
	oc delete namespace "$aNamespace" --wait=false
	oc wait namespace "$aNamespace" --for=delete --timeout=30s

	# if a namespace with this name does not exist, all is good, exit
	if ! oc get namespace "$aNamespace"; then
		return 0
	fi

	# otherwise force delete namespace
	oc get namespace "$aNamespace" -ojson | sed '/"kubernetes"/d' >temp.yaml
	oc proxy &
	pid=$!
	echo "PID: $pid"
	sleep 5
	curl -H "Content-Type: application/yaml" -X PUT --data-binary @temp.yaml http://127.0.0.1:8001/api/v1/namespaces/"$aNamespace"/finalize
	kill -9 "$pid"
	oc wait namespace "$aNamespace" --for=delete --timeout=300s
}

# Check if the number of parameters is correct
if [ "$#" -eq 0 ]; then
	OPERATOR_CATALOG=redhat-operators
elif [ "$#" -eq 1 ]; then
	OPERATOR_CATALOG=$1
	# get all the packages present in the cluster catalog
	oc get packagemanifest -o jsonpath='{range .items[*]}{.metadata.name}{","}{.status.catalogSource}{"\n"}{end}' | grep "$OPERATOR_CATALOG" | head -n -1 >"$OPERATOR_LIST_PATH"

elif [ "$#" -eq 2 ]; then
	OPERATOR_CATALOG=$1
	OPERATORS_UNDER_TEST=$2
	echo "$OPERATORS_UNDER_TEST " | sed 's/ /,'"$OPERATOR_CATALOG"'\n/g' >"$OPERATOR_LIST_PATH"
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
	echo "Time: <b>$TIMESTAMP</b>, file: <b>$OPERATOR_CATALOG</b>"

	#add detailed results
	echo ", detailed results: "'<a href="/'"$REPORT_FOLDER_RELATIVE"'/index.html">'"link"'</a>'

	# add CSV file link
	echo ", CSV: "
	echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/results.csv">'"link"'</a>'

	# add operator list link
	echo ", operator list: "
	echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$OPERATOR_LIST_FILENAME"'">'"link"'</a>'

	# new line
	echo "<br>"
} >>"$BASE_DIR"/index.html

echo "$OPERATOR_PAGE" >>"$REPORT_FOLDER"/index.html

cleanup

# For each operator in a provided catalog, this script will install the operator and run the CNF test suite.
while IFS=, read -r package_name catalog; do
	if [ "$package_name" = "" ]; then
		continue
	fi
	# Workaround for cleaning operator leftovers, see https://access.redhat.com/solutions/6971276
	oc delete mutatingwebhookconfigurations controller.devfile.io
	oc delete validatingwebhookconfigurations controller.devfile.io

	echo "package=$package_name catalog=$catalog"

	namesCount=$(tasty install "$package_name" --source "$catalog" --stdout | grep -c "name:")

	if [ "$namesCount" = "4" ]; then
		# get namespace from tasty
		ns=$(tasty install "$package_name" --source "$catalog" --stdout | grep "name:" | head -n1 | awk '{ print $2 }')
	elif [ "$namesCount" = "2" ]; then
		ns="openshift-operators"
	fi

	echo "namespace=$ns"

	# if a namespace is present, it is probably stuck deleting from previous runs. Force delete it.
	forceDeleteNamespaceIfPresent "$ns"

	# install the operator in a custom namespace
	tasty install "$package_name" --source "$catalog" -w

	# setting report directory
	reportDir="$REPORT_FOLDER"/"$package_name"

	# store the results of CNF test in a new directory
	mkdir -p "$reportDir"

	configYaml="$reportDir"/tnf_config.yml

	# change the targetNameSpace in tng_config file
	sed "s/\$ns/$ns/" "$CONFIG_YAML_TEMPLATE" >"$configYaml"

	# wait for the CSV to appear
	waitForCsvToAppearAndLabel "$ns"

	if [ "$?" = 1 ]; then
		# add per operator links
		{
			# add error message
			echo "Results for: <b>$package_name</b>, "'<span style="color: red;">Operator installation failed, skipping test</span>'

			# add tnf_config link
			echo ", tnf_config: "
			echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$package_name"'/tnf_config.yml">'"link"'</a>'

			# new line
			echo "<br>"
		} >>"$REPORT_FOLDER"/index.html
		# remove the operator
		tasty remove "$package_name"

		cleanup
		waitDeleteNamespace "$ns"

		continue
	fi

	echo "operator $package_name installed"

	# label everything
	oc get deployment -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "oc label " $3  " -n " $2 " " $1  " test-network-function.com/generic=target "}' | bash
	oc get statefulset -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "oc label " $3  " -n " $2 " " $1  " test-network-function.com/generic=target "}' | bash
	oc get pods -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "oc label " $3  " -n " $2 " " $1  " test-network-function.com/generic=target "}' | bash

	# run tnf-container
	./run-tnf-container.sh -k "$KUBECONFIG" -t "$reportDir" -o "$reportDir" -c "$DOCKER_CONFIG" -l all

	# unlabel and uninstall the operator
	oc get csv -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "oc label " $3  " -n " $2 " " $1  " test-network-function.com/operator- "}' | bash

	# remove the operator
	tasty remove "$package_name"

	cleanup
	waitDeleteNamespace "$ns"

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

done <"$OPERATOR_LIST_PATH"

# Workaround for cleaning operator leftovers, see https://access.redhat.com/solutions/6971276
oc delete mutatingwebhookconfigurations controller.devfile.io
oc delete validatingwebhookconfigurations controller.devfile.io

# Resetting project to default
oc project default

# closing html file
echo '</body></html>' >>"$REPORT_FOLDER"/index.html
