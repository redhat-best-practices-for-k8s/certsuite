#!/usr/bin/env bash
set -euo pipefail

CATALOG_NAMESPACE="operators-golden-catalog-smoke-ns"
OPERATOR_NAMESPACE="operators-golden-smoke-ns"
TENANT_NAMESPACE="operators-golden-tenant-smoke-ns"
RESULTS_DIR="${1:?Usage: cleanup.sh <results-dir>}"

# The operator scenarios span several namespaces, so they delete them directly
# instead of the shared single-namespace cleanup-workload.sh. Deleting the
# cluster-scoped CRD first lets the next run apply a fresh version.
kubectl delete crd goldencrds.scenario-smoke.example.com --ignore-not-found --timeout=120s

# Delete without waiting for finalization: every entry owns its namespaces, so
# the runner reaps them in parallel at the end of the suite.
delete_namespaces() {
  kubectl delete namespace "$@" --ignore-not-found --wait=false
  if [[ -n "${SCENARIO_REAP_FILE:-}" ]]; then
    printf '%s\n' "$@" >>"${SCENARIO_REAP_FILE}"
  fi
}

rm -rf "$RESULTS_DIR"
delete_namespaces "$CATALOG_NAMESPACE" "$OPERATOR_NAMESPACE" "$TENANT_NAMESPACE"
