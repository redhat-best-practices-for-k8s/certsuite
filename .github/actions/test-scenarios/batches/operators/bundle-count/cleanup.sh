#!/usr/bin/env bash
set -euo pipefail

CATALOG_NAMESPACE="operators-big-catalog-smoke-ns"
OPERATOR_NAMESPACE="operators-big-smoke-ns"
RESULTS_DIR="${1:?Usage: cleanup.sh <results-dir>}"

# The operator scenarios span several namespaces, so they delete them directly
# instead of the shared single-namespace cleanup-workload.sh.
rm -rf "$RESULTS_DIR"
kubectl delete namespace "$CATALOG_NAMESPACE" "$OPERATOR_NAMESPACE" --ignore-not-found --timeout=120s
