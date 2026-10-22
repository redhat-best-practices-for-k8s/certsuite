#!/usr/bin/env bash
set -euo pipefail

# Clean up unsecured-container-ports scenario resources.
# Usage: cleanup.sh <results-dir>

NAMESPACE="${SCENARIO_NAMESPACE:-unsecured-ports-ns}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

"${SCRIPT_DIR}/../../cleanup-workload.sh" "$NAMESPACE" "${1:?Usage: cleanup.sh <results-dir>}"

echo "Unsecured ports test resources cleaned up."
