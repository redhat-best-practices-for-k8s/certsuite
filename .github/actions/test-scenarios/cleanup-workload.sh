#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${1:?Usage: cleanup-workload.sh <namespace> <results-dir> [--wait]}"
RESULTS_DIR="${2:?Usage: cleanup-workload.sh <namespace> <results-dir> [--wait]}"
WAIT="${3:-}"

rm -rf "$RESULTS_DIR"

if [[ "$WAIT" == "--wait" ]]; then
  # Scenarios whose cluster-scoped objects depend on the namespace being gone
  # (the pod-spec volumes) still have to block on finalization.
  kubectl delete namespace "$NAMESPACE" --ignore-not-found --timeout=120s
else
  # Waiting for finalization is the largest single cost in the suite (~12s per
  # scenario). Delete in the background and let the runner reap every
  # namespace in parallel once the suite is done.
  kubectl delete namespace "$NAMESPACE" --ignore-not-found --wait=false
  if [[ -n "${SCENARIO_REAP_FILE:-}" ]]; then
    printf '%s\n' "$NAMESPACE" >>"${SCENARIO_REAP_FILE}"
  fi
fi
