#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${SCENARIO_NAMESPACE:-namespace-crd-noncompliant-ns}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# The CRD is cluster-scoped and outlives the namespace. Deleting it also
# removes the CR created in the default namespace.
kubectl delete crd scenarioversionedforeigncrds.scenario-smoke.example.com --ignore-not-found --timeout=120s

exec "${SCRIPT_DIR}/../../cleanup-workload.sh" "$NAMESPACE" "${1:?Usage: cleanup.sh <results-dir>}"
