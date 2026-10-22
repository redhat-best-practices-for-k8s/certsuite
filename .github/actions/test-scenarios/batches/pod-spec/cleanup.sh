#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${SCENARIO_NAMESPACE:-pod-spec-smoke-ns}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Deleting the namespace does not remove the cluster-scoped RuntimeClass, the
# ClusterRoleBinding or the StorageClasses.
kubectl delete -f "${SCRIPT_DIR}/manifests/runtimeclass.yaml" --ignore-not-found
kubectl delete clusterrolebinding bad-crb-binding --ignore-not-found
kubectl delete storageclass good-nfs-sc bad-local-storage-sc bad-retain-sc --ignore-not-found

# Entries that bind the static hostPath PVs have to wait for the namespace to
# finish terminating: the PVs below stay wedged on the pvc-protection finalizer
# until their claims are gone. The other entries in this path use no volumes and
# can take the non-blocking path.
WAIT_ARG=""
if [[ -n "$(kubectl get pvc -n "$NAMESPACE" --no-headers --ignore-not-found 2>/dev/null)" ]]; then
  WAIT_ARG="--wait"
fi
"${SCRIPT_DIR}/../../cleanup-workload.sh" "$NAMESPACE" "${1:?Usage: cleanup.sh <results-dir>}" "$WAIT_ARG"

# The statically bound hostPath volumes survive namespace deletion in Released
# state (their claimRef still names the deleted claims) and can never rebind,
# so delete them explicitly once the claims are gone.
kubectl delete persistentvolume good-data-pv bad-storage-pv bad-retain-pv --ignore-not-found --timeout=120s
