#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="operators-bad-smoke-ns"
RESULTS_DIR="${1:?Usage: cleanup.sh <results-dir>}"

# Deleting the cluster-scoped CRD first lets the next run apply a fresh
# version; the CSVs live inside the namespace and go with it.
kubectl delete crd badcrds.scenario-smoke.example.com --ignore-not-found --timeout=120s

rm -rf "$RESULTS_DIR"
kubectl delete namespace "$NAMESPACE" --ignore-not-found --timeout=120s
