#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${1:?Usage: cleanup-workload.sh <namespace> <results-dir>}"
RESULTS_DIR="${2:?Usage: cleanup-workload.sh <namespace> <results-dir>}"

kubectl delete namespace "$NAMESPACE" --ignore-not-found --timeout=120s || true
rm -rf "$RESULTS_DIR"
