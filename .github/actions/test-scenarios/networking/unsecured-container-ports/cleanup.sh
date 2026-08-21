#!/usr/bin/env bash
set -euo pipefail

# Clean up unsecured-container-ports scenario resources.

NAMESPACE="unsecured-ports-ns"
RESULTS_DIR="${1:-unsecured-ports-results}"

kubectl delete namespace "$NAMESPACE" --ignore-not-found
rm -rf "$RESULTS_DIR"

echo "Unsecured ports test resources cleaned up."
