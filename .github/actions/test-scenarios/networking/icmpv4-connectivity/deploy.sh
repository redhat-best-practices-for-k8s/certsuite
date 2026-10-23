#!/usr/bin/env bash
set -euo pipefail

# Deploy the workloads of an icmpv4-connectivity manifest and wait until all of
# them are ready. The shared deploy covers the Deployments; the extra wait
# covers the pods they create before the check looks for their addresses.
# Usage: deploy.sh <manifest-file>

NAMESPACE="${SCENARIO_NAMESPACE:-icmpv4-smoke-ns}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFEST_PATH="${SCRIPT_DIR}/manifests/${1:?Usage: deploy.sh <manifest-file>}"

"${SCRIPT_DIR}/../../deploy-workload.sh" "$NAMESPACE" "$MANIFEST_PATH"
kubectl wait pod -n "$NAMESPACE" -l scenario-icmpv4=target --for=condition=Ready --timeout=120s
