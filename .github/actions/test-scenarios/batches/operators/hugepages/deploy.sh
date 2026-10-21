#!/usr/bin/env bash
set -euo pipefail

# Deploy an OLM operator whose pod requests hugepages-2Mi. The Kind workers
# expose no hugepages, so the pod is created but never schedules. That broken
# state is the point of the scenario: certsuite still associates the Pending
# pod with its CSV through the owner chain, and the hugepages check fails on
# it. We therefore wait for the pod to be unschedulable, not for it to run -
# the CSV never reaches Succeeded.
# Usage: deploy.sh <manifest-file>

CATALOG_NAMESPACE="operators-hugepages-catalog-smoke-ns"
OPERATOR_NAMESPACE="operators-hugepages-smoke-ns"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFEST_PATH="${SCRIPT_DIR}/manifests/${1:?Usage: deploy.sh <manifest-file>}"

kubectl create namespace "$CATALOG_NAMESPACE"
kubectl create namespace "$OPERATOR_NAMESPACE"
kubectl apply -f "$MANIFEST_PATH"

kubectl rollout status deployment/scenario-smoke-hugepages-opm -n "$CATALOG_NAMESPACE" --timeout=180s

ADDRESS="$(kubectl get service scenario-smoke-hugepages-opm -n "$CATALOG_NAMESPACE" -o jsonpath='{.spec.clusterIP}'):50051"
kubectl patch catalogsource scenario-smoke-hugepages-cs -n "$OPERATOR_NAMESPACE" --type merge -p "{\"spec\":{\"address\":\"${ADDRESS}\"}}"

kubectl apply -f "${SCRIPT_DIR}/manifests/subscription.yaml"

# Waiting on the unschedulable pod covers the whole chain: InstallPlan ->
# CSV -> Deployment -> ReplicaSet -> Pending pod. kubectl wait fails
# immediately on an empty selector match, so first wait for the pod to be
# created, then for the scheduler to have marked it unschedulable.
kubectl wait --for=create pod -l app=scenario-smoke-hugepages -n "$OPERATOR_NAMESPACE" --timeout=180s
kubectl wait --for=condition=PodScheduled=false pod -l app=scenario-smoke-hugepages -n "$OPERATOR_NAMESPACE" --timeout=180s
