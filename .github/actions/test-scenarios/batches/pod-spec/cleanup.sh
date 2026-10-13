#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="pod-spec-smoke-ns"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Deleting the namespace does not remove the cluster-scoped RuntimeClass.
kubectl delete -f "${SCRIPT_DIR}/manifests/runtimeclass.yaml" --ignore-not-found
exec "${SCRIPT_DIR}/../../cleanup-workload.sh" "$NAMESPACE" "${1:?Usage: cleanup.sh <results-dir>}"
