#!/usr/bin/env bash
set -euo pipefail

# Apply hand-written CSVs in a namespace with no OperatorGroup: OLM reconciles
# them to phase=Failed within seconds but never runs their install strategy,
# so their specs stay exactly as written for certsuite to inspect. The bad CRD
# carries a served version named "v0", which the apiserver accepts (valid
# DNS1035 label) but certsuite's version regex rejects.
# Usage: deploy.sh <manifest-file>

NAMESPACE="operators-bad-smoke-ns"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFEST_PATH="${SCRIPT_DIR}/manifests/${1:?Usage: deploy.sh <manifest-file>}"

kubectl create namespace "$NAMESPACE"
kubectl apply -n "$NAMESPACE" -f "$MANIFEST_PATH"
kubectl wait --for=condition=Established crd/badcrds.scenario-smoke.example.com --timeout=120s

for csv in scenario-smoke-bad-scc.v1.0.0 scenario-smoke-dup-operator.v1.0.0 scenario-smoke-dup-operator.v2.0.0; do
	kubectl wait --for=jsonpath='{.status.phase}'=Failed clusterserviceversion/"$csv" -n "$NAMESPACE" --timeout=120s
done
