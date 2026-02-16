#!/bin/bash
set -o nounset -o pipefail

# Test run timestamp
TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S_%Z")

# Base folder
BASE_DIR=/var/www/html

# index.html
INDEX_FILE=index2.html

# INPUTS

# certsuite_config.yaml template file path
CONFIG_YAML_TEMPLATE="$(pwd)"/certsuite_config.yml.template

# CatalogSource.yaml template file path
CATALOG_SOURCE_TEMPLATE="$(pwd)"/CatalogSource.yaml.template

# Docker config used to pull operator images
DOCKER_CONFIG=config.json

# Location of telco/non-telco classification file
CNF_TYPE_DIR="$(pwd)"/cmd/certsuite/claim/show/csv

# Operator catalog name
OPERATOR_CATALOG_NAME="operator-catalog"

# Operator catalog namespace
OPERATOR_CATALOG_NAMESPACE="openshift-marketplace"

# Operator from user
OPERATORS_UNDER_TEST=""

# Certsuite container image
CERTSUITE_IMAGE_NAME=quay.io/redhat-best-practices-for-k8s/certsuite
CERTSUITE_IMAGE_TAG=unstable

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
	printf "$format\n" "$@" >>"$LOG_FILE_PATH"
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

	# Remove all test labels from all namespaces
	echo_color "$BLUE" "Removing test labels from all resources in all namespaces"

	# Remove operator labels from CSVs in all namespaces
	oc get csv --all-namespaces -o json 2>/dev/null |
		jq -r '.items[] | select(.metadata.labels."redhat-best-practices-for-k8s.com/operator" != null) | .metadata.namespace + " " + .metadata.name' 2>/dev/null |
		while read -r ns name; do
			[ -n "$ns" ] && [ -n "$name" ] && oc label csv -n "$ns" "$name" redhat-best-practices-for-k8s.com/operator- 2>/dev/null || true
		done

	# Remove generic labels from deployments in all namespaces
	oc get deployment --all-namespaces -o json 2>/dev/null |
		jq -r '.items[] | select(.metadata.labels."redhat-best-practices-for-k8s.com/generic" != null) | .metadata.namespace + " " + .metadata.name' 2>/dev/null |
		while read -r ns name; do
			[ -n "$ns" ] && [ -n "$name" ] && oc label deployment -n "$ns" "$name" redhat-best-practices-for-k8s.com/generic- 2>/dev/null || true
		done

	# Remove generic labels from statefulsets in all namespaces
	oc get statefulset --all-namespaces -o json 2>/dev/null |
		jq -r '.items[] | select(.metadata.labels."redhat-best-practices-for-k8s.com/generic" != null) | .metadata.namespace + " " + .metadata.name' 2>/dev/null |
		while read -r ns name; do
			[ -n "$ns" ] && [ -n "$name" ] && oc label statefulset -n "$ns" "$name" redhat-best-practices-for-k8s.com/generic- 2>/dev/null || true
		done

	# Remove generic labels from pods in all namespaces
	oc get pods --all-namespaces -o json 2>/dev/null |
		jq -r '.items[] | select(.metadata.labels."redhat-best-practices-for-k8s.com/generic" != null) | .metadata.namespace + " " + .metadata.name' 2>/dev/null |
		while read -r ns name; do
			[ -n "$ns" ] && [ -n "$name" ] && oc label pod -n "$ns" "$name" redhat-best-practices-for-k8s.com/generic- 2>/dev/null || true
		done
}

wait_delete_namespace() {
	local namespace_deleting=$1
	# Wait for the namespace to be removed
	if [ "$namespace_deleting" != "openshift-operators" ] && [ "$namespace_deleting" != "openshift-storage" ]; then

		echo_color "$BLUE" "non openshift-operators namespace = $namespace_deleting, deleting "
		with_retry 2 0 oc wait namespace "$namespace_deleting" --for=delete --timeout=60s || true

		force_delete_namespace_if_present "$namespace_deleting" >>"$LOG_FILE_PATH" 2>&1 || true
	else
		if [ "$namespace_deleting" = "openshift-storage" ]; then
			echo_color "$BLUE" "Skipping deletion of openshift-storage namespace"
		fi
	fi
}

# Executes command with retry
with_retry() {
	local \
		max_retries=$1 \
		interval_sec=$2 \
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
		sleep "$interval_sec"
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

wait_package_ok() {
	local \
		package_name=$1 \
		start_time \
		timeout_seconds=600

	start_time=$(date +%s 2>&1) || {
		echo "date failed with error $?: $start_time" >>"$LOG_FILE_PATH"
		return 0
	}

	while true; do
		(oc get packagemanifests | grep "$package_name") &>/dev/null &&
			return 0 ||
			echo "get packagemanifest $package_name failed with error $?." >>"$LOG_FILE_PATH"

		current_time=$(date +%s)
		elapsed_time=$((current_time - start_time))
		# If elapsed time is greater than the timeout report failure
		if [ "$elapsed_time" -ge "$timeout_seconds" ]; then
			echo_color "$BLUE" "Timeout reached $timeout_seconds seconds waiting for packagemanifest $package_name to be reachable."
			return 1
		fi

		# Otherwise wait a bit
		echo_color "$BLUE" "Waiting for package $package_name to be reachable..."
		sleep 5
	done
}

wait_all_packages_ok() {
	local \
		start_time \
		prev_count \
		curr_count \
		elapsed_time \
		timeout_seconds=600

	prev_count="$(get_packages)"
	start_time="$(date +%s 2>&1)" || {
		echo "date failed with error $?: $start_time" >>"$LOG_FILE_PATH"
		return 0
	}

	# wait until package number is stable
	while true; do
		curr_count=$(get_packages)
		if [ "${curr_count}" -ne "${prev_count}" ] || [ "${curr_count}" -eq 0 ]; then
			prev_count="${curr_count}"
		else
			return 0
		fi

		curr_time="$(date +%s)"
		elapsed_time="$((curr_time - start_time))"
		# If elapsed time is greater than the timeout report failure
		if [ "$elapsed_time" -ge "$timeout_seconds" ]; then
			echo_color "$RED" "Timeout reached $timeout_seconds seconds waiting for packagemanifests to be reachable."
			return 1
		fi

		# Otherwise wait a bit
		echo_color "$BLUE" "Waiting for packages to be reachable..."
		sleep 5
	done
}

get_packages() {
	oc get packagemanifest \
		-n ${OPERATOR_CATALOG_NAMESPACE} -o json |
		jq -r '.items[] | select(.status.catalogSource == "'${OPERATOR_CATALOG_NAME}'") | .metadata.name' |
		wc -w
}

# CSV-based labeling: labels only the specific operator's CSV (not all CSVs in namespace)
wait_for_csv_to_appear_and_label() {
	local csv_namespace=$1
	local operator_package=$2
	local timeout_seconds=100
	local start_time=0
	local current_time=0
	local elapsed_time=0
	local csv_name=""
	local status=0

	start_time=$(date +%s)
	while true; do
		# Wait for the specific operator's CSV (name starts with package name)
		# Try to get CSV name from subscription status first
		csv_name=$(oc get subscription "$operator_package" -n "$csv_namespace" -o jsonpath='{.status.installedCSV}' 2>/dev/null)
		
		# If subscription.status.installedCSV is empty (e.g., for + operators with conflicts),
		# fall back to querying CSV directly by pattern matching
		if [ -z "$csv_name" ] || [ "$csv_name" = "<none>" ]; then
			csv_name=$(oc get csv -n "$csv_namespace" -o custom-columns=':.metadata.name' --no-headers 2>/dev/null | grep -i "$operator_package" | head -1)
		fi
		if [ -n "$csv_name" ]; then
			# Found the CSV for this operator
			break
		else
			current_time=$(date +%s)
			elapsed_time=$((current_time - start_time))
			# If elapsed time is greater than the timeout report failure
			if [ "$elapsed_time" -ge "$timeout_seconds" ]; then
				echo_color "$BLUE" "Timeout reached $timeout_seconds seconds waiting for CSV for $operator_package."
				return 1
			fi

			# Otherwise wait a bit
			echo_color "$BLUE" "Waiting for csv for $operator_package to be created in namespace $csv_namespace ..."
			sleep 5
		fi
	done

	# Label only the specific CSV for this operator with "redhat-best-practices-for-k8s.com/operator=target"
	echo_color "$GREY" "Labeling CSV: $csv_name"
	with_retry 5 10 oc label csv -n "$csv_namespace" "$csv_name" redhat-best-practices-for-k8s.com/operator=target --overwrite 2>>"$LOG_FILE_PATH" || true

	# Wait for the CSV to be succeeded
	echo_color "$BLUE" "Wait for CSV to be succeeded"
	with_retry 60 5 oc wait csv "$csv_name" -n "$csv_namespace" --for=jsonpath=\{.status.phase\}=Succeeded --timeout=10s || status="$?"
	return $status
}

force_delete_namespace_if_present() {
	local a_namespace=$1
	local pid=0

	# Do not delete the openshift-operators or openshift-storage namespaces
	if [ "$a_namespace" = "openshift-operators" ] || [ "$a_namespace" = "openshift-storage" ]; then
		return 0
	fi

	# If a namespace with this name does not exist, all is good, exit
	if ! oc get namespace "$a_namespace"; then
		return 0
	fi

	# Remove finalizers
	remove_all_finalizers "$a_namespace" "namespace" "" || true

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
	local skip_cleanup=$4
	local message=$5

	# Skip uninstall for operators with + suffix or legacy lvms/odf operators
	if [ "$skip_cleanup" = true ] || [ "$package_name" = "lvms-operator" ] || [ "$package_name" = "odf-operator" ]; then
		echo_color "$BLUE" "Skipping uninstall and namespace deletion for $package_name"
	else
		with_retry 3 5 oc operator uninstall -X "$package_name" -n "$ns" || true
		wait_delete_namespace "$ns"
	fi

	# Add per operator links
	{
		# Add error message
		echo "Results for: <b>$package_name</b>, "'<span style="color: red;">'"$message"'</span>'

		# Add certsuite_config link
		echo ", certsuite_config: "
		echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$package_name"'/certsuite_config.yml">'"link"'</a>'

		# New line
		echo "<br>"
	} >>"$REPORT_FOLDER"/"$INDEX_FILE"
}

get_suggested_namespace() {
	local package_name=$1

	oc get packagemanifests -n openshift-marketplace "$package_name" -ojson | jq -r '.status.channels[].currentCSVDesc.annotations."operatorframework.io/suggested-namespace"' 2>/dev/null | grep -v "null" | sed 's/\n//g' | head -1 || true
}

create_catalog() {
	catalog_source_yaml=catalogSource.yml
	sed "s|\$CATALOG_INDEX|$CATALOG_INDEX|" "$CATALOG_SOURCE_TEMPLATE" >"$catalog_source_yaml"

	oc apply -f $catalog_source_yaml
	wait_pods_ok
}

wait_pods_ok() {
	local \
		start_time \
		timeout_seconds=100
	start_time=$(date +%s 2>&1) || {
		echo "date failed with error $?: $start_time" >>"$LOG_FILE_PATH"
		return 0
	}

	while true; do
		pods=$(oc get pods --no-headers -n openshift-marketplace -o custom-columns=":metadata.name,:status.phase" | grep ^"$OPERATOR_CATALOG_NAME"-)
		all_pods_running=true

		# Iterate over all necessary pods and check their phases.
		while IFS= read -r pod; do
			pod_status=$(echo "$pod" | awk '{print $2}')
			if [[ $pod_status != "Running" ]]; then
				all_pods_running=false
			fi
		done <<<"$pods"

		if [ "$all_pods_running" = true ]; then
			break
		fi

		current_time=$(date +%s)
		elapsed_time=$((current_time - start_time))
		# If elapsed time is greater than the timeout report failure
		if [ "$elapsed_time" -ge "$timeout_seconds" ]; then
			echo_color "$BLUE" "Timeout reached $timeout_seconds seconds waiting for packagemanifest $package_name to be reachable."
			return 1
		fi

		echo_color "$BLUE" "Waiting for necessary pods to be created and reach running state..."
		sleep 5
	done
}

remove_all_finalizers() {
	local resource_name=$1
	local resource_type=$2
	local namespace=$3

	echo "Removing finalizers from $resource_type/$resource_name..."

	if [ "$resource_type" == "namespace" ]; then
		# For namespaces, do not use the namespace argument
		if oc get "$resource_type" "$resource_name" -o json |
			jq 'del(.metadata.finalizers)' |
			oc apply -f -; then
			echo "Successfully removed finalizers from $resource_type/$resource_name."
		else
			echo "Failed to remove finalizers from $resource_type/$resource_name."
			return 1
		fi
	else
		# For other resource types, include the namespace argument
		if oc get "$resource_type" "$resource_name" -n "$namespace" -o json |
			jq 'del(.metadata.finalizers)' |
			oc apply -f -; then
			echo "Successfully removed finalizers from $resource_type/$resource_name."
		else
			echo "Failed to remove finalizers from $resource_type/$resource_name."
			return 1
		fi
	fi
}

# Main

# Writing CatalogSource template
cat <<EOF >"$CATALOG_SOURCE_TEMPLATE"
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: $OPERATOR_CATALOG_NAME
  namespace: $OPERATOR_CATALOG_NAMESPACE
spec:
  sourceType: grpc
  image: \$CATALOG_INDEX
  displayName: Operator Catalog
  publisher: Redhat
EOF

# Check if the number of parameters is correct
if [ "$#" -eq 1 ]; then
	CATALOG_INDEX=$1
	echo_color "$BLUE" "Creating Catalog Source"
	create_catalog
	# Get all the packages present in the cluster catalog
	wait_all_packages_ok
	with_retry 5 10 oc get packagemanifest -o jsonpath='{range .items[*]}{.metadata.name}{",'"$CATALOG_INDEX"'\n"}{end}' | head -n -1 | sort >"$OPERATOR_LIST_PATH"

elif [ "$#" -eq 2 ]; then
	CATALOG_INDEX=$1
	echo_color "$BLUE" "Creating Catalog Source"
	create_catalog
	OPERATORS_UNDER_TEST=$2
	echo "$OPERATORS_UNDER_TEST " | sed 's| |,'"$CATALOG_INDEX"'\n|g' >"$OPERATOR_LIST_PATH"
else
	echo 'Wrong parameter count.
  Usage: ./run-basic-batch-operators-test.sh <catalog-index> ["<operator-name 1> <operator-name 2> ... <operator-name N>"]
  Examples:
  ./run-basic-batch-operators-test.sh registry.redhat.io/redhat-operators
  ./run-basic-batch-operators-test.sh registry.redhat.io/redhat-operators "file-integrity-operator kiali-ossm"'
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
  - "redhat-best-practices-for-k8s.com/generic: target"
operatorsUnderTestLabels:
  - "redhat-best-practices-for-k8s.com/operator: target"
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
	echo "Time: <b>$TIMESTAMP</b>, Catalog index: <b>$CATALOG_INDEX</b>"

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
while IFS=, read -r package_name catalog_index; do
	if [ "$package_name" = "" ]; then
		continue
	fi

	echo_color "$GREY" "********* package= $package_name catalog index= $catalog_index **********"

	# Check for suffix indicators (+ or -)
	skip_cleanup=false
	force_test_namespace=false
	actual_package_name="$package_name"

	if [[ "$package_name" == *+ ]]; then
		# + suffix means skip uninstall and namespace deletion
		skip_cleanup=true
		actual_package_name="${package_name%+}"
		echo_color "$BLUE" "Package has + suffix: will skip cleanup for $actual_package_name"
	elif [[ "$package_name" == *- ]]; then
		# - suffix means use test-<packagename> namespace with normal processing
		force_test_namespace=true
		actual_package_name="${package_name%-}"
		echo_color "$BLUE" "Package has - suffix: will use test-$actual_package_name namespace"
	fi

	# Wait for the cluster to be reachable
	echo_color "$BLUE" "Wait for cluster to be reachable"
	wait_cluster_ok

	# Wait for package to be reachable
	echo_color "$BLUE" "Wait for package $actual_package_name to be reachable"
	wait_package_ok "$actual_package_name"

	# Variable to hold return status
	status=0

	# Determine namespace based on suffix and suggested namespace
	# Special case: odf-csi-addons-operator and mcg-operator always use openshift-storage
	if [ "$actual_package_name" = "odf-csi-addons-operator" ] || [ "$actual_package_name" = "mcg-operator" ]; then
		ns="openshift-storage"
		echo_color "$BLUE" "using openshift-storage namespace for $actual_package_name"
	elif [ "$force_test_namespace" = true ]; then
		# - suffix: always use test-<packagename>
		ns="test-$actual_package_name"
		echo_color "$BLUE" "using forced test namespace for $actual_package_name: $ns"
	else
		# Normal logic or + suffix
		ns=$(get_suggested_namespace "$actual_package_name")
		if [ "$ns" = "" ] || [ "$ns" = "openshift-operators" ]; then
			if [ "$skip_cleanup" = true ]; then
				# + suffix with no suggested namespace: use test-<packagename>
				ns="test-$actual_package_name"
				echo_color "$BLUE" "no suggested namespace for $actual_package_name, using: $ns"
			else
				# No suffix with no suggested namespace: use test-operator
				ns="test-operator"
				echo_color "$BLUE" "no suggested namespace for $actual_package_name, using: test-operator"
			fi
		else
			echo_color "$BLUE" "using suggested namespace for $actual_package_name: $ns"
		fi
	fi
	echo_color "$GREY" "namespace= $ns"

	echo_color "$BLUE" "Cluster cleanup"
	if ! cleanup >>"$LOG_FILE_PATH" 2>&1; then
		echo_color "$RED" "Warning, cluster cleanup failed"
	fi

	# Handle openshift-storage namespace - create if it doesn't exist, but never delete it
	if [ "$ns" = "openshift-storage" ]; then
		if oc get namespace "$ns" &>/dev/null; then
			echo_color "$BLUE" "Namespace openshift-storage already exists, using it"
		else
			echo_color "$BLUE" "Creating namespace openshift-storage"
			if ! oc create namespace "$ns"; then
				echo_color "$RED" "Error, creating namespace openshift-storage failed"
			fi
		fi
	elif [ "$skip_cleanup" = true ]; then
		# For operators with + suffix, preserve existing namespace if present
		if oc get namespace "$ns" &>/dev/null; then
			echo_color "$BLUE" "Namespace $ns already exists, preserving it (+ suffix)"
		else
			echo_color "$BLUE" "Creating namespace $ns"
			if ! oc create namespace "$ns"; then
				echo_color "$RED" "Error, creating namespace $ns failed"
			fi
		fi
	else
		# Normal processing: force delete namespace if present, then create it
		echo_color "$BLUE" "Remove namespace if present"
		if ! force_delete_namespace_if_present "$ns" >>"$LOG_FILE_PATH" 2>&1; then
			echo_color "$RED" "Error, force deleting namespace failed"
		fi

		if ! oc create namespace "$ns"; then
			echo_color "$RED" "Error, creating namespace $ns failed"
		fi
	fi

	# Install the operator in a custom namespace
	echo_color "$BLUE" "install operator"
	install_status=0
	
	# For + suffix operators, check if subscription already exists before attempting install
	if [ "$skip_cleanup" = true ]; then
		existing_subscription=$(oc get subscription "$actual_package_name" -n "$ns" -o name 2>/dev/null || true)
		
		if [ -n "$existing_subscription" ]; then
			echo_color "$BLUE" "Operator subscription already exists for + suffix operator, skipping oc operator install"
			echo_color "$BLUE" "Existing subscription: $existing_subscription"
			install_output="Subscription already exists, skipped oc operator install"
		else
			# Subscription does not exist, proceed with install
			install_output=$(oc operator install --create-operator-group "$actual_package_name" -n "$ns" 2>&1) || install_status=$?
		fi
	else
		# Normal operator (no + suffix), always run install
		install_output=$(oc operator install --create-operator-group "$actual_package_name" -n "$ns" 2>&1) || install_status=$?
	fi

	if [ "$install_status" -ne 0 ]; then
		# Check if it's a "already exists" error and we're using + suffix
		if [ "$skip_cleanup" = true ] && echo "$install_output" | grep -q "already exists"; then
			echo_color "$BLUE" "Operator subscription already exists (expected with + suffix), continuing..."
		else
			echo_color "$RED" "Operator installation failed but will still waiting for CSV"
			echo "$install_output" >>"$LOG_FILE_PATH"
		fi
	else
		echo "$install_output"
	fi
	# Setting report directory
	report_dir="$REPORT_FOLDER"/"$actual_package_name"

	# Store the results of CNF test in a new directory
	if ! mkdir -p "$report_dir"; then
		echo_color "$RED" "Error, creating report dir failed"
	fi

	config_yaml="$report_dir"/certsuite_config.yml

	# Change the target_name_space in certsuite_config file
	sed "s/\$ns/$ns/" "$CONFIG_YAML_TEMPLATE" >"$config_yaml"

	# Wait for the CSV to appear and label only the specific operator's CSV
	echo_color "$BLUE" "Wait for CSV to appear and label resources under test"
	if ! wait_for_csv_to_appear_and_label "$ns" "$actual_package_name"; then
		echo_color "$RED" "Operator failed to install, continue"
		report_failure "$status" "$ns" "$actual_package_name" "$skip_cleanup" "Operator installation failed, skipping test"
		continue
	fi

	echo_color "$BLUE" "operator $actual_package_name installed"

	# Special handling for multicluster-engine operator with + suffix
	if [ "$actual_package_name" = "multicluster-engine" ] && [ "$skip_cleanup" = true ]; then
		echo_color "$BLUE" "Creating MultiClusterEngine custom resource"
		if cat <<EOF | oc apply -f - >>"$LOG_FILE_PATH" 2>&1; then
apiVersion: multicluster.openshift.io/v1
kind: MultiClusterEngine
metadata:
  name: multiclusterengine
spec: {}
EOF
			echo_color "$BLUE" "MultiClusterEngine custom resource created successfully"
			echo_color "$BLUE" "Waiting for MultiClusterEngine to be ready..."
			sleep 30
		else
			echo_color "$RED" "Failed to create MultiClusterEngine CR"
		fi
	fi

	echo_color "$BLUE" "Wait to ensure all pods are running"
	# Extra wait to ensure that all pods are running
	sleep 30

	# CSV-based labeling: label only operator-specific deployments, statefulsets, and pods
	echo_color "$BLUE" "Label operator-specific deployments, statefulsets, and pods"
	# Get the CSV name for the operator under test (already labeled earlier)
	csv_name=$(oc get csv -n "$ns" -l redhat-best-practices-for-k8s.com/operator=target -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)

	if [ -n "$csv_name" ]; then
		# Get the CSV UID for owner reference matching
		csv_uid=$(oc get csv "$csv_name" -n "$ns" -o jsonpath='{.metadata.uid}' 2>/dev/null)

		# Extract deployment names from CSV's spec.install.spec.deployments
		deployment_names=$(oc get csv "$csv_name" -n "$ns" -o jsonpath='{.spec.install.spec.deployments[*].name}' 2>/dev/null)

		# Label only the deployments defined in the CSV
		for dep_name in $deployment_names; do
			echo_color "$GREY" "Labeling deployment: $dep_name"
			oc label deployment -n "$ns" "$dep_name" redhat-best-practices-for-k8s.com/generic=target --overwrite 2>>"$LOG_FILE_PATH" || true

			# Label pods belonging to this deployment (using deployment's selector)
			selector=$(oc get deployment "$dep_name" -n "$ns" -o jsonpath='{.spec.selector.matchLabels}' 2>/dev/null | jq -r 'to_entries | map("\(.key)=\(.value)") | join(",")' 2>/dev/null)
			if [ -n "$selector" ]; then
				oc label pods -n "$ns" -l "$selector" redhat-best-practices-for-k8s.com/generic=target --overwrite 2>>"$LOG_FILE_PATH" || true
			fi
		done

		# Find and label statefulsets owned by the CSV (via ownerReferences)
		if [ -n "$csv_uid" ]; then
			echo_color "$GREY" "Looking for statefulsets owned by CSV: $csv_name"
			statefulset_names=$(oc get statefulset -n "$ns" -o json 2>/dev/null | jq -r --arg uid "$csv_uid" '.items[] | select(.metadata.ownerReferences[]?.uid == $uid) | .metadata.name' 2>/dev/null)

			for sts_name in $statefulset_names; do
				echo_color "$GREY" "Labeling statefulset: $sts_name"
				oc label statefulset -n "$ns" "$sts_name" redhat-best-practices-for-k8s.com/generic=target --overwrite 2>>"$LOG_FILE_PATH" || true

				# Label pods belonging to this statefulset (using statefulset's selector)
				selector=$(oc get statefulset "$sts_name" -n "$ns" -o jsonpath='{.spec.selector.matchLabels}' 2>/dev/null | jq -r 'to_entries | map("\(.key)=\(.value)") | join(",")' 2>/dev/null)
				if [ -n "$selector" ]; then
					oc label pods -n "$ns" -l "$selector" redhat-best-practices-for-k8s.com/generic=target --overwrite 2>>"$LOG_FILE_PATH" || true
				fi
			done
		fi
	else
		echo_color "$RED" "Warning: Could not find labeled CSV, falling back to namespace-wide labeling"
		# Fallback to original behavior
		oc get deployment -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "  oc label " $3  " -n " $2 " " $1  " redhat-best-practices-for-k8s.com/generic=target "}' | bash || true
		oc get statefulset -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "  oc label " $3  " -n " $2 " " $1  " redhat-best-practices-for-k8s.com/generic=target "}' | bash || true
		oc get pods -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "  oc label " $3  " -n " $2 " " $1  " redhat-best-practices-for-k8s.com/generic=target "}' | bash || true
	fi

	# Run certsuite container
	echo_color "$BLUE" "run CNF suite"

	config_dir="$(pwd)"/config
	mkdir -p "$config_dir"
	cp "$KUBECONFIG" "$config_dir"/kubeconfig
	cp "$DOCKER_CONFIG" "$config_dir"/dockerconfig
	cp "$config_yaml" "$config_dir"/certsuite_config.yaml

	podman run --rm \
		--network=host \
		-v "${config_dir}:/config:Z" \
		-v "${report_dir}:/reports:Z" \
		"${CERTSUITE_IMAGE_NAME}:${CERTSUITE_IMAGE_TAG}" \
		/usr/local/bin/certsuite run \
		--kubeconfig=/config/kubeconfig \
		--preflight-dockerconfig=/config/dockerconfig \
		--config-file=/config/certsuite_config.yaml \
		--output-dir=/reports \
		--label-filter=all >>"$LOG_FILE_PATH" 2>&1 || {
		report_failure "$status" "$ns" "$actual_package_name" "$skip_cleanup" "CNF suite exited with errors"
		continue
	}

	echo_color "$BLUE" "unlabel operator"
	# Unlabel the operator
	if ! oc get csv -n "$ns" -o custom-columns=':.metadata.name,:.metadata.namespace,:.kind' | sed '/^ *$/d' | awk '{print "  oc label " $3  " -n " $2 " " $1  " redhat-best-practices-for-k8s.com/operator- "}' | bash; then
		echo_color "$RED" "Error, failed to unlabel the operator"
	fi

	# Check if cleanup should be skipped (+ suffix or legacy lvms/odf operators)
	if [ "$skip_cleanup" = true ] || [ "$actual_package_name" = "lvms-operator" ] || [ "$actual_package_name" = "odf-operator" ]; then
		echo_color "$BLUE" "Skipping uninstall and namespace deletion for $actual_package_name"
	else
		# remove the operator
		echo_color "$BLUE" "Remove operator"
		if ! oc operator uninstall -X "$actual_package_name" -n "$ns"; then
			echo_color "$RED" "Operator failed to un-install, continue"
		fi

		# Skip namespace deletion for openshift-storage
		if [ "$ns" = "openshift-storage" ]; then
			echo_color "$BLUE" "Skipping namespace deletion for openshift-storage"
		else
			# Delete the namespace
			if ! oc delete namespace "$ns" --wait=false; then
				echo_color "$RED" "Error, failed to delete namespace: $ns"
			fi

			echo_color "$BLUE" "Wait for cleanup to finish"
			if ! wait_delete_namespace "$ns"; then
				echo_color "$RED" "Error, fail to wait for the namespace to be deleted"
			fi
		fi
	fi

	# Check parsing claim file
	echo_color "$BLUE" "Parse claim file"

	# merge claim.json from each operator to a single csv file
	echo_color "$BLUE" "add claim.json from this operator to the csv file"
	if ! podman run --rm \
		-v "${report_dir}:/reports:Z" \
		-v "${CNF_TYPE_DIR}:/cnftype:Z" \
		"${CERTSUITE_IMAGE_NAME}:${CERTSUITE_IMAGE_TAG}" \
		/usr/local/bin/certsuite claim \
		show csv -t /cnftype/cnf-type.json -c /reports/claim.json -n "$actual_package_name" "$add_headers" >>"$REPORT_FOLDER"/results.csv; then
		echo_color "$RED" "failed to parse claim file"
	fi

	# extract parser
	echo_color "$BLUE" "extract parser from report"
	if ! with_retry 2 10 tar -xvf "$report_dir"/*.tar.gz -C "$report_dir" results.html; then
		echo_color "$RED" "Failed get result.html from report"
	fi

	# Add per operator links
	{
		# Add parser link
		echo "Results for: <b>$actual_package_name</b>,  parsed details:"
		echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$actual_package_name"'/results.html?claimfile=/'"$REPORT_FOLDER_RELATIVE"'/'"$actual_package_name"'/claim.json">'"link"'</a>'

		# Add log link
		echo ", log: "
		echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$actual_package_name"'/certsuite.log">'"link"'</a>'

		# Add certsuite_config link
		echo ", certsuite_config: "
		echo '<a href="/'"$REPORT_FOLDER_RELATIVE"'/'"$actual_package_name"'/certsuite_config.yml">'"link"'</a>'

		# new line
		echo "<br>"
	} >>"$REPORT_FOLDER"/"$INDEX_FILE"

	# Only print headers once
	add_headers=""

done <"$OPERATOR_LIST_PATH"

# Delete the catalog
echo_color "$BLUE" "Remove Catalog"
if ! oc delete catalogsources -n "$OPERATOR_CATALOG_NAMESPACE" "$OPERATOR_CATALOG_NAME"; then
	echo_color "$RED" "Error, failed to delete catalog: $OPERATOR_CATALOG_NAME"
fi

# closing html file
echo '</body></html>' >>"$REPORT_FOLDER"/"$INDEX_FILE"
echo_color "$GREEN" DONE
