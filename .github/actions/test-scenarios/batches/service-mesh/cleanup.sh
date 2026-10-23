#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${SCENARIO_NAMESPACE:-service-mesh-smoke-ns}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STAND_IN_LABEL="$(kubectl get deployment istiod -n istio-system \
  -o jsonpath='{.metadata.labels.scenario-smoke}' 2>/dev/null || true)"

# Only remove a mesh this scenario created; a real Istio installation must
# survive. The next entry recreates the stand-in immediately, and a namespace
# still terminating would reject that create, so the delete waits. It is not
# fatal either way: the workload namespace has to be cleaned up regardless.
if [ "$STAND_IN_LABEL" = "service-mesh" ]; then
  kubectl delete namespace istio-system --timeout=120s ||
    echo "warning: istio-system is still terminating" >&2
fi

exec "${SCRIPT_DIR}/../../cleanup-workload.sh" "$NAMESPACE" "${1:?Usage: cleanup.sh <results-dir>}"
