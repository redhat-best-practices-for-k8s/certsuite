#!/usr/bin/env bash
set -euo pipefail

# The service-mesh check skips unless autodiscovery finds an installed mesh, so
# create the istio-system namespace and its istiod Deployment before the
# workload under test. A mesh that is already installed is left alone: applying
# over it would replace the real control plane with this stand-in. Both entries
# share the stand-in; cleanup removes it.
# Usage: deploy.sh <manifest-file>

NAMESPACE="${SCENARIO_NAMESPACE:-service-mesh-smoke-ns}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFEST_PATH="${SCRIPT_DIR}/manifests/${1:?Usage: deploy.sh <manifest-file>}"
STAND_IN_LABEL="$(kubectl get deployment istiod -n istio-system \
  -o jsonpath='{.metadata.labels.scenario-smoke}' 2>/dev/null || true)"

if kubectl get namespace istio-system >/dev/null 2>&1 && [ "$STAND_IN_LABEL" != "service-mesh" ]; then
  echo "istio-system already exists without the scenario stand-in; using the installed mesh" >&2
else
  kubectl apply -f "${SCRIPT_DIR}/manifests/istio.yaml"
  kubectl wait deployment/istiod -n istio-system --for=condition=Available --timeout=120s
fi

kubectl create namespace "$NAMESPACE"
kubectl apply -n "$NAMESPACE" -f "$MANIFEST_PATH"
kubectl wait pod -n "$NAMESPACE" -l scenario-service-mesh=target --for=condition=Ready --timeout=180s
