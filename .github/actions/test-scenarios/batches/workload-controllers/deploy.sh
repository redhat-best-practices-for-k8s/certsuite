#!/usr/bin/env bash
set -euo pipefail

# Deploy every workload in a batch manifest and wait until all of them are
# ready. Pods that are not ready are left out of certsuite discovery, which
# would turn an expected failure into a pass.
# Usage: deploy.sh <manifest-file>

NAMESPACE="controllers-smoke-ns"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFEST_PATH="${SCRIPT_DIR}/manifests/${1:?Usage: deploy.sh <manifest-file>}"

kubectl create namespace "$NAMESPACE"
kubectl apply -n "$NAMESPACE" -f "$MANIFEST_PATH"
kubectl wait deployment -n "$NAMESPACE" --all --for=condition=Available --timeout=180s
kubectl wait pod -n "$NAMESPACE" -l scenario-controllers=target --for=condition=Ready --timeout=180s
