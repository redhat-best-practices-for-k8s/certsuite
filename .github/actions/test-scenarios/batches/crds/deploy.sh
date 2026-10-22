#!/usr/bin/env bash
set -euo pipefail

# Deploy the CRD and Role of a batch manifest. There are no pods to wait for;
# the CRD must be Established before certsuite can discover it.
# Usage: deploy.sh <manifest-file>

NAMESPACE="${SCENARIO_NAMESPACE:-crds-smoke-ns}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFEST_PATH="${SCRIPT_DIR}/manifests/${1:?Usage: deploy.sh <manifest-file>}"

kubectl create namespace "$NAMESPACE"
kubectl apply -n "$NAMESPACE" -f "$MANIFEST_PATH"
kubectl wait --for=condition=Established crd/scenariocrds.scenario-smoke.example.com --timeout=120s
