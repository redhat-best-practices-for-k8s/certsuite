#!/usr/bin/env bash
set -euo pipefail

# Clean up TLS scenario test resources.
# Usage: cleanup.sh <results-dir>

NAMESPACE="${SCENARIO_NAMESPACE:-tls-test-ns}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

"${SCRIPT_DIR}/../../cleanup-workload.sh" "$NAMESPACE" "${1:?Usage: cleanup.sh <results-dir>}"

echo "TLS test resources cleaned up."
