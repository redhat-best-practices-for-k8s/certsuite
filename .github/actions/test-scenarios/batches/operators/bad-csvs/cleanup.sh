#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="operators-bad-smoke-ns"
RESULTS_DIR="${1:?Usage: cleanup.sh <results-dir>}"

# Deleting the cluster-scoped CRD first lets the next run apply a fresh
# version; the CSVs live inside the namespace and go with it.
kubectl delete crd badcrds.scenario-smoke.example.com --ignore-not-found --timeout=120s

# Delete without waiting for finalization: every entry owns its namespaces, so
# the runner reaps them in parallel at the end of the suite.
delete_namespaces() {
  kubectl delete namespace "$@" --ignore-not-found --wait=false
  if [[ -n "${SCENARIO_REAP_FILE:-}" ]]; then
    printf '%s\n' "$@" >>"${SCENARIO_REAP_FILE}"
  fi
}

rm -rf "$RESULTS_DIR"
delete_namespaces "$NAMESPACE"
