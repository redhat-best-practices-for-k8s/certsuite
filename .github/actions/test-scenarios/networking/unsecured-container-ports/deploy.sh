#!/usr/bin/env bash
set -euo pipefail

# Deploy unsecured-container-ports workloads spread across nodes.
# Usage: deploy.sh <manifest-file>
#
# The manifest may hold several Deployments; every one of them is required to
# span at least two nodes.

MANIFEST_FILE="${1:?Usage: deploy.sh <manifest-file>}"

NAMESPACE="${SCENARIO_NAMESPACE:-unsecured-ports-ns}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFESTS_DIR="${SCRIPT_DIR}/manifests"

kubectl create namespace "$NAMESPACE"

if [[ "$MANIFEST_FILE" == tls-* ]]; then
  openssl req -x509 -nodes -days 1 -newkey rsa:2048 \
    -keyout /tmp/tls.key -out /tmp/tls.crt \
    -subj "/CN=tls-test.${NAMESPACE}.svc"

  kubectl create secret tls tls-test-cert \
    --cert=/tmp/tls.crt --key=/tmp/tls.key \
    -n "$NAMESPACE"

  kubectl create configmap nginx-tls -n "$NAMESPACE" --from-literal=default.conf='
server {
    listen 8443 ssl;
    ssl_certificate /etc/tls/tls.crt;
    ssl_certificate_key /etc/tls/tls.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    location / { return 200 "tls"; }
}'
fi

kubectl apply -n "$NAMESPACE" -f "${MANIFESTS_DIR}/${MANIFEST_FILE}"
kubectl wait deployment -n "$NAMESPACE" --all --for=condition=Available --timeout=120s

# Each workload's replicas must land on separate nodes: the checks under test
# probe every pod, so a single-node placement would make the cross-node case
# trivially pass.
for deploy in $(kubectl get deployment -n "$NAMESPACE" \
  -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}'); do
  node_count=$(kubectl get pods -n "$NAMESPACE" -l "app=${deploy}" \
    -o jsonpath='{range .items[*]}{.spec.nodeName}{"\n"}{end}' | sort -u | grep -c . || true)

  if [[ "${node_count}" -lt 2 ]]; then
    echo "FAIL: expected ${deploy} pods on at least 2 nodes, got ${node_count}"
    kubectl get pods -n "$NAMESPACE" -o wide
    exit 1
  fi

  echo "Workload '${deploy}' deployed on ${node_count} nodes."
done
