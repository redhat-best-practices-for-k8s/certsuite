#!/usr/bin/env bash
set -euo pipefail

# Run all test scenarios defined in scenarios.json.
# Each scenario deploys a single workload, runs certsuite, validates the
# result against the expected state, and cleans up.
#
# Usage: run-scenarios.sh [--log-level LEVEL]
#
# Requires: jq, kubectl, ./certsuite binary

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCENARIOS_FILE="${SCRIPT_DIR}/scenarios.json"
LOG_LEVEL="${SMOKE_TESTS_LOG_LEVEL:-info}"
OVERALL_RC=0

claim_object_count() {
  local claim_file=$1 test_id=$2 list_name=$3 needle=$4
  jq -r --arg id "$test_id" --arg list "$list_name" --arg needle "$needle" '
    .claim.results[$id].checkDetails // "{}"
    | fromjson?
    | [.[$list] // [] | .[] | select(($needle == "") or ((.ObjectFieldsValues | join(" ")) | contains($needle)))]
    | length
  ' "$claim_file"
}

require_min_objects() {
  local label=$1 count=$2 min=$3
  if [[ "$count" -lt "$min" ]]; then
    echo "FAIL: expected at least ${min} ${label} objects, got ${count}"
    return 1
  fi
  echo "PASS: ${label} object count ${count} >= ${min}"
}

scenario_count=$(jq 'length' "$SCENARIOS_FILE")
echo "=== Running ${scenario_count} test scenario(s) ==="

for i in $(seq 0 $((scenario_count - 1))); do
  NAME=$(jq -r ".[$i].name" "$SCENARIOS_FILE")
  LABEL_FILTER=$(jq -r ".[$i].label_filter" "$SCENARIOS_FILE")
  SCENARIO_PATH=$(jq -r ".[$i].path" "$SCENARIOS_FILE")
  MANIFEST=$(jq -r ".[$i].manifest" "$SCENARIOS_FILE")
  OUTPUT_DIR=$(jq -r ".[$i].output_dir" "$SCENARIOS_FILE")
  EXPECTED_RESULT=$(jq -r ".[$i].expected_result" "$SCENARIOS_FILE")
  MIN_NONCOMPLIANT=$(jq -r ".[$i].min_noncompliant // empty" "$SCENARIOS_FILE")
  MIN_COMPLIANT=$(jq -r ".[$i].min_compliant // empty" "$SCENARIOS_FILE")
  COMPLIANT_CONTAINS=$(jq -r ".[$i].compliant_contains // empty" "$SCENARIOS_FILE")

  SCENARIO_DIR="${SCRIPT_DIR}/${SCENARIO_PATH}"
  CONFIG_FILE="${SCENARIO_DIR}/manifests/certsuite-config.yaml"

  echo ""
  echo "========================================"
  echo "Scenario: ${NAME}"
  echo "  Label filter:    ${LABEL_FILTER}"
  echo "  Manifest:        ${MANIFEST}"
  echo "  Expected result: ${EXPECTED_RESULT}"
  echo "========================================"

  RC=0

  # Deploy
  echo "--- Deploy ---"
  if ! "${SCENARIO_DIR}/deploy.sh" "${MANIFEST}"; then
    echo "FAIL: deploy failed for scenario '${NAME}'"
    RC=1
  fi

  # Run certsuite (allow failure since some scenarios expect it)
  if [[ $RC -eq 0 ]]; then
    echo "--- Run certsuite ---"
    ./certsuite run \
      --label-filter="${LABEL_FILTER}" \
      --config-file="${CONFIG_FILE}" \
      --output-dir="${OUTPUT_DIR}" \
      --log-level="${LOG_LEVEL}" || true
  fi

  # Validate
  if [[ $RC -eq 0 ]]; then
    echo "--- Validate ---"
    CLAIM_FILE="${OUTPUT_DIR}/claim.json"
    if [[ ! -f "$CLAIM_FILE" ]]; then
      echo "FAIL: claim.json not found at ${CLAIM_FILE}"
      RC=1
    else
      ACTUAL_STATE=$(jq -r --arg id "$LABEL_FILTER" '.claim.results[$id].state // empty' "$CLAIM_FILE")
      if [[ -z "$ACTUAL_STATE" ]]; then
        echo "FAIL: test ${LABEL_FILTER} not found in claim.json"
        RC=1
      elif [[ "$ACTUAL_STATE" != "$EXPECTED_RESULT" ]]; then
        echo "FAIL: expected '${EXPECTED_RESULT}', got '${ACTUAL_STATE}'"
        RC=1
      else
        echo "PASS: test state '${ACTUAL_STATE}' matches expected '${EXPECTED_RESULT}'"
        if [[ -n "${MIN_NONCOMPLIANT}" ]]; then
          NONCOMPLIANT_COUNT=$(claim_object_count "$CLAIM_FILE" "$LABEL_FILTER" "NonCompliantObjectsOut" "")
          require_min_objects "non-compliant" "$NONCOMPLIANT_COUNT" "$MIN_NONCOMPLIANT" || RC=1
        fi
        if [[ -n "${MIN_COMPLIANT}" ]]; then
          COMPLIANT_COUNT=$(claim_object_count "$CLAIM_FILE" "$LABEL_FILTER" "CompliantObjectsOut" "$COMPLIANT_CONTAINS")
          require_min_objects "compliant" "$COMPLIANT_COUNT" "$MIN_COMPLIANT" || RC=1
        fi
      fi
    fi
  fi

  # Print debug log on failure
  if [[ $RC -ne 0 ]]; then
    echo "--- Scenario certsuite.log (debug) ---"
    if [[ -f "certsuite.log" ]]; then
      grep -iE "tls|tlsversion|Probing|exec.*fallback|probe.*pod|probePods|DaemonSet|daemonset|cnf-suite|unsecured|plaintext|openssl|TLS probe" certsuite.log || echo "(no matching log lines)"
    else
      echo "(certsuite.log not found)"
    fi
  fi

  # Cleanup (always runs)
  echo "--- Cleanup ---"
  "${SCENARIO_DIR}/cleanup.sh" "${OUTPUT_DIR}" || true

  if [[ $RC -ne 0 ]]; then
    OVERALL_RC=1
  else
    echo "Scenario '${NAME}' passed."
  fi
done

echo ""
if [[ $OVERALL_RC -ne 0 ]]; then
  echo "=== One or more scenarios FAILED ==="
  exit 1
fi
echo "=== All scenarios passed ==="
