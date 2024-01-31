#!/usr/bin/env bash

# configure_tnf_container_client configures the underlying container virtualization client.  If the user sets the
# TNF_CONTAINER_CLIENT environment variable, then that value is utilized.  Otherwise, "podman" is used by default.
# This is particularly useful for Operating Systems which do not readily support "podman", and "docker" must be used.
function configure_tnf_container_client() {
	PODMAN_EXECUTABLE="podman"
	DEFAULT_CONTAINER_EXECUTABLE="${PODMAN_EXECUTABLE}"

	if [ "" == "${TNF_CONTAINER_CLIENT}" ]; then
		echo "The \$TNF_CONTAINER_CLIENT environment variable is not set; defaulting to use: ${DEFAULT_CONTAINER_EXECUTABLE}"
		export TNF_CONTAINER_CLIENT="${DEFAULT_CONTAINER_EXECUTABLE}"
	else
		echo "\$TNF_CONTAINER_CLIENT is set;  using: ${TNF_CONTAINER_CLIENT}"
	fi
}

# call the function to configure "podman" or something else if specified by TNF_CONTAINER_CLIENT
configure_tnf_container_client

CONTAINER_TNF_DIR=/usr/tnf
CONTAINER_TNF_OFFLINE_DB_DIR=/usr/offline-db
CONTAINER_TNF_KUBECONFIG_FILE_BASE_PATH="$CONTAINER_TNF_DIR/kubeconfig/config"
CONTAINER_TNF_DOCKERCFG_FILE_BASE_PATH="$CONTAINER_TNF_DIR/dockercfg/config.json"
CONTAINER_DEFAULT_NETWORK_MODE=bridge
CONTAINER_DEFAULT_TNF_NON_INTRUSIVE_ONLY=false
CONTAINER_DEFAULT_TNF_ALLOW_PREFLIGHT_INSECURE=false
CONTAINER_DEFAULT_TNF_DISABLE_CONFIG_AUTODISCOVER=false
TNF_LOG_LEVEL_DEFAULT=info
TNF_OMIT_ARTIFACTS_ZIP_FILE_DEFAULT=false
TNF_INCLUDE_WEB_FILES_IN_OUTPUT_FOLDER_DEFAULT=false
ON_DEMAND_DEBUG_PODS_DEFAULT=false
TNF_ENABLE_DATA_COLLECTION_DEFAULT=false
TNF_ENABLE_XML_CREATION_DEFAULT=false

get_container_tnf_kubeconfig_path_from_index() {
	local local_path_index="$1"
	kubeconfig_path=$CONTAINER_TNF_KUBECONFIG_FILE_BASE_PATH

	# To maintain backward compatibility with the TNF container image,
	# indexing of kubeconfigs starts from the second file.
	# For example:
	# - /usr/tnf/kubeconfig/config
	# - /usr/tnf/kubeconfig/config.2
	# - /usr/tnf/kubeconfig/config.3
	if ((local_path_index > 0)); then
		kubeconfig_index=$((local_path_index + 1))
		kubeconfig_path="$kubeconfig_path.$kubeconfig_index"
	fi
	echo $kubeconfig_path
}

get_container_tnf_dockercfg_path_from_index() {
	local local_path_index="$1"
	dockercfg_path=$CONTAINER_TNF_DOCKERCFG_FILE_BASE_PATH
	if ((local_path_index > 0)); then
		dockercfg_index=$((local_path_index + 1))
		dockercfg_path="$dockercfg_path.$dockercfg_index"
	fi
	echo $dockercfg_path
}

display_config_summary() {
	printf "Mounting %d kubeconfig volume(s):\n" "${#container_tnf_kubeconfig_volume_bindings[@]}"
	printf -- "-v %s\n" "${container_tnf_kubeconfig_volume_bindings[@]}"

	# Only display this -v line if the variable is set.
	if [ -n "${container_tnf_dockercfg_volume_bindings}" ]; then
		printf "Mounting %d dockercfg volume(s):\n" "${#container_tnf_dockercfg_volume_bindings[@]}"
		printf -- "-v %s\n" "${container_tnf_dockercfg_volume_bindings[@]}"
	fi

	# Checks whether a prefix of the selected image path matches the address of the official TNF repository
	if [[ "$TNF_IMAGE" != $TNF_OFFICIAL_ORG* ]]; then
		printf "Warning: Could not verify whether '%s' is an official TNF image.\n" "$TNF_IMAGE"
		printf "\t Official TNF images can be pulled directly from '%s'.\n" "$TNF_OFFICIAL_ORG"
	fi
}

join_paths() {
	local IFS=:
	echo "$*"
}

# Explode loaded KUBECONFIG (e.g. /kubeconfig/path1:/kubeconfig/path2:...)
# into an array of individual paths to local kubeconfigs.
# shellcheck disable=SC2162 # Read without -r will mangle backslashes.
IFS=: read -a local_kubeconfig_paths <<<"$LOCAL_KUBECONFIG"

declare -a container_tnf_kubeconfig_paths
declare -a container_tnf_kubeconfig_volume_bindings

# Assign a file in the TNF container for each provided local kubeconfig
for local_path_index in "${!local_kubeconfig_paths[@]}"; do
	local_path=${local_kubeconfig_paths[$local_path_index]}
	container_path=$(get_container_tnf_kubeconfig_path_from_index "$local_path_index")

	container_tnf_kubeconfig_paths+=("$container_path")
	container_tnf_kubeconfig_volume_bindings+=("$local_path:$container_path:Z")
done

# Explode loaded DOCKERCFG
# shellcheck disable=SC2162 # Read without -r will mangle backslashes.
if [[ $LOCAL_DOCKERCFG != NA ]]; then
	IFS=: read -a local_dockercfg_paths <<<"$LOCAL_DOCKERCFG"

	declare -a container_tnf_dockercfg_paths
	declare -a container_tnf_dockercfg_volume_bindings

	# Assign a file in the TNF container for each provided local dockercfg
	for local_path_index in "${!local_dockercfg_paths[@]}"; do
		local_path=${local_dockercfg_paths[$local_path_index]}
		container_path=$(get_container_tnf_dockercfg_path_from_index "$local_path_index")

		container_tnf_dockercfg_paths+=("$container_path")
		container_tnf_dockercfg_volume_bindings+=("$local_path:$container_path:Z")
	done
fi

TNF_IMAGE="${TNF_IMAGE:-$TNF_OFFICIAL_IMAGE}"
CONTAINER_NETWORK_MODE="${CONTAINER_NETWORK_MODE:-$CONTAINER_DEFAULT_NETWORK_MODE}"
CONTAINER_TNF_NON_INTRUSIVE_ONLY="${TNF_NON_INTRUSIVE_ONLY:-$CONTAINER_DEFAULT_TNF_NON_INTRUSIVE_ONLY}"
CONTAINER_TNF_ALLOW_PREFLIGHT_INSECURE="${TNF_ALLOW_PREFLIGHT_INSECURE:-$CONTAINER_DEFAULT_TNF_ALLOW_PREFLIGHT_INSECURE}"
CONTAINER_TNF_DISABLE_CONFIG_AUTODISCOVER="${TNF_DISABLE_CONFIG_AUTODISCOVER:-$CONTAINER_DEFAULT_TNF_DISABLE_CONFIG_AUTODISCOVER}"
TNF_LOG_LEVEL="${TNF_LOG_LEVEL:-$TNF_LOG_LEVEL_DEFAULT}"
ON_DEMAND_DEBUG_PODS="${ON_DEMAND_DEBUG_PODS:-$ON_DEMAND_DEBUG_PODS_DEFAULT}"
TNF_OMIT_ARTIFACTS_ZIP_FILE="${TNF_OMIT_ARTIFACTS_ZIP_FILE:-$TNF_OMIT_ARTIFACTS_ZIP_FILE_DEFAULT}"
TNF_INCLUDE_WEB_FILES_IN_OUTPUT_FOLDER="${TNF_INCLUDE_WEB_FILES_IN_OUTPUT_FOLDER:-$TNF_INCLUDE_WEB_FILES_IN_OUTPUT_FOLDER_DEFAULT}"
TNF_ENABLE_DATA_COLLECTION="${TNF_ENABLE_DATA_COLLECTION:-$TNF_ENABLE_DATA_COLLECTION_DEFAULT}"
TNF_ENABLE_XML_CREATION="${TNF_ENABLE_XML_CREATION:-$TNF_ENABLE_XML_CREATION_DEFAULT}"
display_config_summary

# Construct new $KUBECONFIG env variable containing all paths to kubeconfigs mounted to the container.
# This environment variable is passed to the TNF container and is made available for use by oc/kubectl.
CONTAINER_TNF_KUBECONFIG=$(join_paths "${container_tnf_kubeconfig_paths[@]}")
container_tnf_kubeconfig_volumes_cmd_args=$(printf -- "-v %s " "${container_tnf_kubeconfig_volume_bindings[@]}")

# Construct new $DOCKERCFG env variable containing all paths to dockercfgs mounted to the container.
# This environment variable is passed to the TNF container
if [[ $LOCAL_DOCKERCFG != NA ]]; then
	CONTAINER_TNF_DOCKERCFG=$(join_paths "${container_tnf_dockercfg_paths[@]}")
	container_tnf_dockercfg_volumes_cmd_args=$(printf -- "-v %s " "${container_tnf_dockercfg_volume_bindings[@]}")
fi

# ${variable:-value} uses new value if undefined or null. ${variable:+value}
# is opposite of the above. See:
#  https://www.grymoire.com/Unix/Sh.html#uh-36
ADD_HOST_ARG="${TNF_ENABLE_CRC_TESTING:+--add-host api.crc.testing:host-gateway}"
CONTAINER_TNF_DOCKERCFG="${CONTAINER_TNF_DOCKERCFG:-NA}"
DNS_ARG="${DNS_ARG:+--dns $DNS_ARG}"
TNF_OFFLINE_DB_MOUNT_ARG="${LOCAL_TNF_OFFLINE_DB:+-v $LOCAL_TNF_OFFLINE_DB:/usr/offline-db-ext:Z}"

set -x
# shellcheck disable=SC2068,SC2086 # Double quote array expansions.
${TNF_CONTAINER_CLIENT} run --rm $DNS_ARG \
	--network $CONTAINER_NETWORK_MODE \
	${container_tnf_kubeconfig_volumes_cmd_args[@]} \
	${container_tnf_dockercfg_volumes_cmd_args[@]} \
	$CONFIG_VOLUME_MOUNT_ARG \
	$TNF_OFFLINE_DB_MOUNT_ARG \
	$ADD_HOST_ARG \
	-v $OUTPUT_LOC:$CONTAINER_TNF_DIR/claim:Z \
	-e KUBECONFIG=$CONTAINER_TNF_KUBECONFIG \
	-e PFLT_DOCKERCONFIG=$CONTAINER_TNF_DOCKERCFG \
	-e TNF_OFFLINE_DB=$CONTAINER_TNF_OFFLINE_DB_DIR \
	-e TNF_NON_INTRUSIVE_ONLY=$CONTAINER_TNF_NON_INTRUSIVE_ONLY \
	-e TNF_ALLOW_PREFLIGHT_INSECURE=$CONTAINER_TNF_ALLOW_PREFLIGHT_INSECURE \
	-e TNF_DISABLE_CONFIG_AUTODISCOVER=$CONTAINER_TNF_DISABLE_CONFIG_AUTODISCOVER \
	-e TNF_PARTNER_REPO=$TNF_PARTNER_REPO \
	-e SUPPORT_IMAGE=$SUPPORT_IMAGE \
	-e TNF_LOG_LEVEL=$TNF_LOG_LEVEL \
	-e ON_DEMAND_DEBUG_PODS=$ON_DEMAND_DEBUG_PODS \
	-e TNF_OMIT_ARTIFACTS_ZIP_FILE=$TNF_OMIT_ARTIFACTS_ZIP_FILE \
	-e TNF_INCLUDE_WEB_FILES_IN_OUTPUT_FOLDER=$TNF_INCLUDE_WEB_FILES_IN_OUTPUT_FOLDER \
	-e TNF_ENABLE_DATA_COLLECTION=$TNF_ENABLE_DATA_COLLECTION \
	-e TNF_ENABLE_XML_CREATION=$TNF_ENABLE_XML_CREATION \
	-e PATH=/usr/bin:/usr/local/oc/bin \
	$TNF_IMAGE \
	$TNF_CMD $OUTPUT_ARG $CONTAINER_TNF_DIR/claim $LABEL_ARG $TNF_LABEL "$@"
