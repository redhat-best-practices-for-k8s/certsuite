#!/usr/bin/env bash
set -euo pipefail

# Deploy a CRD that serves v1 and retires v1beta1, plus one CR in the scenario
# namespace. The CRD must be Established before the CR can be created, and
# before certsuite autodiscovers it.
# Usage: deploy.sh <manifest-file>

NAMESPACE="${SCENARIO_NAMESPACE:-namespace-crd-versions-smoke-ns}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFEST_PATH="${SCRIPT_DIR}/manifests/${1:?Usage: deploy.sh <manifest-file>}"

kubectl create namespace "$NAMESPACE"
kubectl apply -n "$NAMESPACE" -f "$MANIFEST_PATH"
kubectl wait --for=condition=Established crd/scenarioversionedcrds.scenario-smoke.example.com --timeout=120s
kubectl apply -n "$NAMESPACE" -f "${SCRIPT_DIR}/manifests/versioned-crd-instance.yaml"
