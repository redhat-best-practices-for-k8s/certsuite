#!/usr/bin/env bash
set -euo pipefail

# Deploy the workloads of an icmpv4-connectivity manifest and wait until all of
# them are ready. deploy-workload.sh cannot be used here: it waits on the one
# deployment named after the manifest, and these manifests carry two.
# Usage: deploy.sh <manifest-file>

NAMESPACE="icmpv4-smoke-ns"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFEST_PATH="${SCRIPT_DIR}/manifests/${1:?Usage: deploy.sh <manifest-file>}"

kubectl create namespace "$NAMESPACE"
kubectl apply -n "$NAMESPACE" -f "$MANIFEST_PATH"
kubectl wait deployment -n "$NAMESPACE" --all --for=condition=Available --timeout=120s
kubectl wait pod -n "$NAMESPACE" -l scenario-icmpv4=target --for=condition=Ready --timeout=120s
