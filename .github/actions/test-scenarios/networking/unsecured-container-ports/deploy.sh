#!/usr/bin/env bash
set -euo pipefail

# Deploy an unsecured-container-ports workload spread across nodes.
# Usage: deploy.sh <manifest-file>

MANIFEST_FILE="${1:?Usage: deploy.sh <manifest-file>}"
DEPLOY_NAME="${MANIFEST_FILE%.yaml}"

NAMESPACE="unsecured-ports-ns"
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
kubectl wait deployment -n "$NAMESPACE" "$DEPLOY_NAME" --for=condition=Available --timeout=120s

NODE_COUNT=$(kubectl get pods -n "$NAMESPACE" -l "app=${DEPLOY_NAME}" \
  -o jsonpath='{range .items[*]}{.spec.nodeName}{"\n"}{end}' | sort -u | grep -c . || true)

if [[ "${NODE_COUNT}" -lt 2 ]]; then
  echo "FAIL: expected pods on at least 2 nodes, got ${NODE_COUNT}"
  kubectl get pods -n "$NAMESPACE" -o wide
  exit 1
fi

echo "Workload '${DEPLOY_NAME}' deployed on ${NODE_COUNT} nodes."
