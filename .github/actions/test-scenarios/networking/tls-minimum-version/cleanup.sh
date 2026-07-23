#!/usr/bin/env bash
set -euo pipefail

# Clean up TLS scenario test resources.

NAMESPACE="tls-test-ns"
RESULTS_DIR="${1:-tls-test-results}"

kubectl delete namespace "$NAMESPACE" --ignore-not-found
rm -rf "$RESULTS_DIR"

echo "TLS test resources cleaned up."
