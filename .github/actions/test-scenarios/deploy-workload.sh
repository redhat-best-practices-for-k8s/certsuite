#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${1:?Usage: deploy-workload.sh <namespace> <manifest-path>}"
MANIFEST_PATH="${2:?Usage: deploy-workload.sh <namespace> <manifest-path>}"

kubectl create namespace "$NAMESPACE"
kubectl apply -n "$NAMESPACE" -f "$MANIFEST_PATH"
# A manifest may hold several Deployments (batch scenarios do), and the
# namespace holds only this scenario's workloads, so wait on all of them.
kubectl wait deployment -n "$NAMESPACE" --all --for=condition=Available --timeout=120s
