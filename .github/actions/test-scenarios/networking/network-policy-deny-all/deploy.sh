#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="network-policy-smoke-ns"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

exec "${SCRIPT_DIR}/../../deploy-workload.sh" "$NAMESPACE" "${SCRIPT_DIR}/manifests/${1:?Usage: deploy.sh <manifest-file>}"
