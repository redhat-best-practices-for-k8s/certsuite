#!/usr/bin/env bash
set -euo pipefail

# Deploy a lifecycle batch manifest. The CRD is applied for both entries: a CR
# cannot be created before its CRD is served, and the violations run creates no
# CR, so the CRD stays out of scale-object discovery there.
# Usage: deploy.sh <manifest-file>

NAMESPACE="${SCENARIO_NAMESPACE:-lifecycle-smoke-ns}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFEST_PATH="${SCRIPT_DIR}/manifests/${1:?Usage: deploy.sh <manifest-file>}"

kubectl apply -f "${SCRIPT_DIR}/manifests/scaling-crd.yaml"
kubectl wait --for=condition=Established crd/scenarioscalingcrs.scaling-smoke.example.com --timeout=120s
# The shared deploy creates the namespace, applies the manifest and waits for
# the Deployments. The golden manifest also has a bare helper Pod, which the
# Deployment wait does not cover; once the Deployments are Available their pods
# exist, so this second wait has something to match.
"${SCRIPT_DIR}/../../deploy-workload.sh" "$NAMESPACE" "$MANIFEST_PATH"
kubectl wait --for=condition=Ready pod --all -n "$NAMESPACE" --timeout=180s
