#!/usr/bin/env bash
set -euo pipefail

# Deploy a fully compliant OLM operator from an in-cluster File-Based Catalog:
# an opm registry serves the FBC out of a ConfigMap, the CatalogSource points
# at it by address, and OLM installs the bundle inline (no bundle image). The
# Subscription is applied only after the address patch so the InstallPlan
# resolves on the first attempt.
# Usage: deploy.sh <manifest-file>

CATALOG_NAMESPACE="operators-golden-catalog-smoke-ns"
OPERATOR_NAMESPACE="operators-golden-smoke-ns"
TENANT_NAMESPACE="operators-golden-tenant-smoke-ns"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFEST_PATH="${SCRIPT_DIR}/manifests/${1:?Usage: deploy.sh <manifest-file>}"

kubectl create namespace "$CATALOG_NAMESPACE"
kubectl create namespace "$OPERATOR_NAMESPACE"
kubectl create namespace "$TENANT_NAMESPACE"
kubectl apply -f "$MANIFEST_PATH"

kubectl rollout status deployment/scenario-smoke-golden-opm -n "$CATALOG_NAMESPACE" --timeout=180s

# The CatalogSource ships with an empty address; now that the registry is up,
# point it at the opm Service's ClusterIP. Address-based sources run no
# registry pod of their own - the Deployment behind the Service is the one
# certsuite's bundle-count probe talks to.
ADDRESS="$(kubectl get service scenario-smoke-golden-opm -n "$CATALOG_NAMESPACE" -o jsonpath='{.spec.clusterIP}'):50051"
kubectl patch catalogsource scenario-smoke-golden-cs -n "$OPERATOR_NAMESPACE" --type merge -p "{\"spec\":{\"address\":\"${ADDRESS}\"}}"

kubectl apply -f "${SCRIPT_DIR}/manifests/subscription.yaml"
kubectl wait --for=jsonpath='{.status.installedCSV}'=scenario-smoke-golden-operator.v1.0.0 subscription/scenario-smoke-golden-subscription -n "$OPERATOR_NAMESPACE" --timeout=300s
kubectl wait --for=jsonpath='{.status.phase}'=Succeeded clusterserviceversion/scenario-smoke-golden-operator.v1.0.0 -n "$OPERATOR_NAMESPACE" --timeout=300s
kubectl wait --for=condition=Established crd/goldencrds.scenario-smoke.example.com --timeout=120s
