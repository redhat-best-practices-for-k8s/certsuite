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

rm -rf "$RESULTS_DIR"
kubectl delete namespace "$CATALOG_NAMESPACE" "$OPERATOR_NAMESPACE" "$TENANT_NAMESPACE" --ignore-not-found --timeout=120s
