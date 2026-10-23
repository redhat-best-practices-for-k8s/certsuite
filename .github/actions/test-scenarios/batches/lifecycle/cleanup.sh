#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${SCENARIO_NAMESPACE:-lifecycle-smoke-ns}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Cluster-scoped, so it outlives the namespace: delete it first, before the
# shared workload cleanup removes the namespace and its CRs.
kubectl delete crd scenarioscalingcrs.scaling-smoke.example.com --ignore-not-found --timeout=120s

exec "${SCRIPT_DIR}/../../cleanup-workload.sh" "$NAMESPACE" "${1:?Usage: cleanup.sh <results-dir>}"
