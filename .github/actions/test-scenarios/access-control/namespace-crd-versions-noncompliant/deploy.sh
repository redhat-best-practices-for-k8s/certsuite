#!/usr/bin/env bash
set -euo pipefail

# Deploy a CRD that serves v1 and retires v1beta1, then create a CR in a
# namespace outside the configured target namespaces. The check must still
# report that CR while ignoring the unserved version.
# Usage: deploy.sh <manifest-file>

NAMESPACE="${SCENARIO_NAMESPACE:-namespace-crd-noncompliant-ns}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFEST_PATH="${SCRIPT_DIR}/manifests/${1:?Usage: deploy.sh <manifest-file>}"

kubectl create namespace "$NAMESPACE"
kubectl apply -n "$NAMESPACE" -f "$MANIFEST_PATH"
kubectl wait --for=condition=Established crd/scenarioversionedforeigncrds.scenario-smoke.example.com --timeout=120s
kubectl apply -n default -f "${SCRIPT_DIR}/manifests/versioned-crd-instance.yaml"
