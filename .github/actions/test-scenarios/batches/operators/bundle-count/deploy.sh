#!/usr/bin/env bash
set -euo pipefail

# Deploy an operator from a catalog with 1002 bundles: one installable inline
# bundle for the operator itself plus 1001 filler bundles that exist only to
# push the bundle count past the 1000 limit. The CSV must install and the
# catalog must be synced so certsuite can resolve the CatalogSource through
# the package server before counting its bundles.
# Usage: deploy.sh <manifest-file>

CATALOG_NAMESPACE="operators-big-catalog-smoke-ns"
OPERATOR_NAMESPACE="operators-big-smoke-ns"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFEST_PATH="${SCRIPT_DIR}/manifests/${1:?Usage: deploy.sh <manifest-file>}"

kubectl create namespace "$CATALOG_NAMESPACE"
kubectl create namespace "$OPERATOR_NAMESPACE"

# kubectl create, not apply: apply records the whole object in the
# kubectl.kubernetes.io/last-applied-configuration annotation, and the
# 1002-bundle FBC (~350KB) exceeds the 262144-byte annotation limit. The
# ConfigMap itself is well under the 1MiB data limit. Deploy runs on a clean
# cluster (cleanup.sh removed the previous state), so create semantics are
# fine.
kubectl create -f "$MANIFEST_PATH"

kubectl rollout status deployment/scenario-smoke-big-opm -n "$CATALOG_NAMESPACE" --timeout=180s

ADDRESS="$(kubectl get service scenario-smoke-big-opm -n "$CATALOG_NAMESPACE" -o jsonpath='{.spec.clusterIP}'):50051"
kubectl patch catalogsource scenario-smoke-big-cs -n "$OPERATOR_NAMESPACE" --type merge -p "{\"spec\":{\"address\":\"${ADDRESS}\"}}"

kubectl apply -f "${SCRIPT_DIR}/manifests/subscription.yaml"
kubectl wait --for=jsonpath='{.status.installedCSV}'=scenario-smoke-big-operator.v1.0.0 subscription/scenario-smoke-big-subscription -n "$OPERATOR_NAMESPACE" --timeout=300s
kubectl wait --for=jsonpath='{.status.phase}'=Succeeded clusterserviceversion/scenario-smoke-big-operator.v1.0.0 -n "$OPERATOR_NAMESPACE" --timeout=300s
