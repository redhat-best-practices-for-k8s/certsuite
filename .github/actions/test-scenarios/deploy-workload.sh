#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${1:?Usage: deploy-workload.sh <namespace> <manifest-path>}"
MANIFEST_PATH="${2:?Usage: deploy-workload.sh <namespace> <manifest-path>}"
DEPLOY_NAME="$(basename "${MANIFEST_PATH%.yaml}")"

kubectl create namespace "$NAMESPACE"
kubectl apply -n "$NAMESPACE" -f "$MANIFEST_PATH"
kubectl wait deployment -n "$NAMESPACE" "$DEPLOY_NAME" --for=condition=Available --timeout=120s
