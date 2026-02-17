#!/usr/bin/env bash
set -euo pipefail

# Deploy a single TLS test workload into tls-test-ns.
# Usage: deploy.sh <manifest-file>
#
# The manifest file is looked up under manifests/ and must define a Deployment
# whose name matches the filename without the .yaml extension.

MANIFEST_FILE="${1:?Usage: deploy.sh <manifest-file>}"
DEPLOY_NAME="${MANIFEST_FILE%.yaml}"

NAMESPACE="tls-test-ns"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFESTS_DIR="${SCRIPT_DIR}/manifests"

kubectl create namespace "$NAMESPACE"

# Generate self-signed TLS certificate
openssl req -x509 -nodes -days 1 -newkey rsa:2048 \
  -keyout /tmp/tls.key -out /tmp/tls.crt \
  -subj "/CN=tls-test.${NAMESPACE}.svc"

kubectl create secret tls tls-test-cert \
  --cert=/tmp/tls.crt --key=/tmp/tls.key \
  -n "$NAMESPACE"

# nginx config: TLS 1.3 only
kubectl create configmap nginx-tls13-only -n "$NAMESPACE" --from-literal=default.conf='
server {
    listen 8443 ssl;
    ssl_certificate /etc/tls/tls.crt;
    ssl_certificate_key /etc/tls/tls.key;
    ssl_protocols TLSv1.3;
    location / { return 200 "tls13-only"; }
}'

# nginx config: TLS 1.2+1.3 with Intermediate-profile ciphers only
kubectl create configmap nginx-tls12-allowed -n "$NAMESPACE" --from-literal=default.conf='
server {
    listen 8443 ssl;
    ssl_certificate /etc/tls/tls.crt;
    ssl_certificate_key /etc/tls/tls.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    location / { return 200 "tls12-allowed"; }
}'

# Deploy the single workload
kubectl apply -n "$NAMESPACE" -f "${MANIFESTS_DIR}/${MANIFEST_FILE}"
kubectl wait deployment -n "$NAMESPACE" "$DEPLOY_NAME" --for=condition=Available --timeout=120s

echo "Workload '${DEPLOY_NAME}' deployed successfully."
