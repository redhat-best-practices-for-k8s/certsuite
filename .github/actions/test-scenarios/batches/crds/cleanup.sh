#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="crds-smoke-ns"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Both manifests define the same CRD; deleting it here lets the next run apply
# its own version (with or without the status subresource).
kubectl delete crd scenariocrds.scenario-smoke.example.com --ignore-not-found --timeout=120s

exec "${SCRIPT_DIR}/../../cleanup-workload.sh" "$NAMESPACE" "${1:?Usage: cleanup.sh <results-dir>}"
