#!/usr/bin/env bash

if [ "$TNF_RUN_CFD_TEST" == "true" ]; then

	export TNF_IMAGE_NAME=cnf-tests
	export TNF_IMAGE_TAG="${TNF_CFD_IMAGE_TAG:-4.6}"
	export TNF_OFFICIAL_ORG=quay.io/openshift-kni/
	export TNF_OFFICIAL_IMAGE="${TNF_OFFICIAL_ORG}${TNF_IMAGE_NAME}:${TNF_IMAGE_TAG}"
	export TNF_CFD_SKIP="${TNF_CFD_SKIP:-performance|sriov|ptp|sctp|xt_u32|dpdk|ovn}"
	export TNF_CMD="/usr/bin/test-run.sh"
	export OUTPUT_ARG="--junit"
	export CONTAINER_NETWORK_MODE="host"

	if [[ -n "$KUBECONFIG" ]]; then
		export LOCAL_KUBECONFIG=$KUBECONFIG
	elif [[ -f "$HOME/.kube/config" ]]; then
		export LOCAL_KUBECONFIG=$HOME/.kube/config
	fi
	
	# For older verions of docker, dns server may need to be set explicitly, e.g.
	#
	# export DNS_ARG=172.0.0.53
	./run-container.sh -ginkgo.v -ginkgo.skip=${TNF_CFD_SKIP} 
else
	# removing report if not running, so the final claim won't include stale reports
	rm -f ${OUTPUT_LOC}/validation_junit.xml
fi