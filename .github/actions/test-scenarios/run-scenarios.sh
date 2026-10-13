#!/usr/bin/env bash
set -euo pipefail

# Run all test scenarios defined in scenarios.json.
# Each scenario deploys its workloads, runs certsuite over the test IDs it
# covers, validates the claim against the expected state, and cleans up.
#
# Usage: run-scenarios.sh
# Env:   SMOKE_TESTS_LOG_LEVEL (default: info)
#
# Requires: jq, kubectl, and a ./certsuite binary built from this checkout.
# See CONTRIBUTING.md for the scenario format and for debugging a failure.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCENARIOS_FILE="${SCRIPT_DIR}/scenarios.json"
LOG_LEVEL="${SMOKE_TESTS_LOG_LEVEL:-info}"
OVERALL_RC=0
CURRENT_SCENARIO_DIR=""
CURRENT_OUTPUT_DIR=""

cleanup_current_scenario() {
  local scenario_dir="$CURRENT_SCENARIO_DIR"
  local output_dir="$CURRENT_OUTPUT_DIR"
  CURRENT_SCENARIO_DIR=""
  CURRENT_OUTPUT_DIR=""

  if [[ -n "$scenario_dir" ]]; then
    "$scenario_dir/cleanup.sh" "$output_dir" || true
  fi
}

trap cleanup_current_scenario EXIT

claim_object_count() {
  local claim_file=$1 test_id=$2 list_name=$3 needle=$4 invert=${5:-false}
  jq -r --arg id "$test_id" --arg list "$list_name" --arg needle "$needle" --argjson invert "$invert" '
    .claim.results[$id].checkDetails // "{}"
    | fromjson?
    | [.[$list] // [] | .[]
       | (($needle == "") or ((.ObjectFieldsValues | join(" ")) | contains($needle))) as $hit
       | select($hit != $invert)]
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

# Check one test's result in the claim against its expectations (a JSON object
# with expected_result and optional min_noncompliant, min_compliant,
# compliant_contains and noncompliant_contains).
validate_test() {
  local claim_file=$1 test_id=$2 spec=$3
  local expected actual min_noncompliant min_compliant compliant_contains noncompliant_contains count foreign
  expected=$(jq -r '.expected_result' <<<"$spec")
  min_noncompliant=$(jq -r '.min_noncompliant // empty' <<<"$spec")
  min_compliant=$(jq -r '.min_compliant // empty' <<<"$spec")
  compliant_contains=$(jq -r '.compliant_contains // empty' <<<"$spec")
  noncompliant_contains=$(jq -r '.noncompliant_contains // empty' <<<"$spec")

  actual=$(jq -r --arg id "$test_id" '.claim.results[$id].state // empty' "$claim_file")
  if [[ -z "$actual" ]]; then
    echo "FAIL: ${test_id}: not found in claim.json"
    return 1
  fi
  if [[ "$actual" != "$expected" ]]; then
    echo "FAIL: ${test_id}: expected '${expected}', got '${actual}'"
    return 1
  fi
  echo "PASS: ${test_id}: test state '${actual}' matches expected '${expected}'"

  local rc=0
  if [[ -n "$noncompliant_contains" ]]; then
    # Every non-compliant object must belong to the expected workload, so a
    # batch cannot pass on another workload's violation.
    foreign=$(claim_object_count "$claim_file" "$test_id" "NonCompliantObjectsOut" "$noncompliant_contains" true)
    if [[ "$foreign" -ne 0 ]]; then
      echo "FAIL: ${test_id}: ${foreign} non-compliant objects do not contain '${noncompliant_contains}'"
      rc=1
    fi
    min_noncompliant=${min_noncompliant:-1}
  fi
  if [[ -n "$min_noncompliant" ]]; then
    count=$(claim_object_count "$claim_file" "$test_id" "NonCompliantObjectsOut" "$noncompliant_contains")
    require_min_objects "${test_id}: non-compliant" "$count" "$min_noncompliant" || rc=1
  fi
  if [[ -n "$min_compliant" ]]; then
    count=$(claim_object_count "$claim_file" "$test_id" "CompliantObjectsOut" "$compliant_contains")
    require_min_objects "${test_id}: compliant" "$count" "$min_compliant" || rc=1
  fi
  return $rc
}

scenario_count=$(jq 'length' "$SCENARIOS_FILE")
echo "=== Running ${scenario_count} test scenario(s) ==="

for i in $(seq 0 $((scenario_count - 1))); do
  NAME=$(jq -r ".[$i].name" "$SCENARIOS_FILE")
  SCENARIO_PATH=$(jq -r ".[$i].path" "$SCENARIOS_FILE")
  MANIFEST=$(jq -r ".[$i].manifest" "$SCENARIOS_FILE")
  OUTPUT_DIR=$(jq -r ".[$i].output_dir" "$SCENARIOS_FILE")
  # A scenario either checks one test (label_filter + expected_result) or a
  # batch of tests in one certsuite run (tests: {id: expectations}).
  TESTS=$(jq -c ".[$i] | .tests // {(.label_filter): del(.name, .description, .label_filter, .path, .manifest, .output_dir)}" "$SCENARIOS_FILE")
  LABEL_FILTER=$(jq -r 'keys_unsorted | join(",")' <<<"$TESTS")

  SCENARIO_DIR="${SCRIPT_DIR}/${SCENARIO_PATH}"
  CONFIG_FILE="${SCENARIO_DIR}/manifests/certsuite-config.yaml"
  CURRENT_SCENARIO_DIR="$SCENARIO_DIR"
  CURRENT_OUTPUT_DIR="$OUTPUT_DIR"

  echo ""
  echo "========================================"
  echo "Scenario: ${NAME}"
  echo "  Manifest: ${MANIFEST}"
  jq -r 'to_entries[] | "  Expect:   \(.key) -> \(.value.expected_result)"' <<<"$TESTS"
  echo "========================================"

  RC=0

  # Deploy
  echo "--- Deploy ---"
  if ! "${SCENARIO_DIR}/deploy.sh" "${MANIFEST}"; then
    echo "FAIL: deploy failed for scenario '${NAME}'"
    RC=1
  fi

  # Run certsuite (allow failure since some scenarios expect it).
  # Keep the probe daemonset between scenarios: deleting it waits for the
  # probe pods to terminate, which costs ~40s per scenario.
  if [[ $RC -eq 0 ]]; then
    echo "--- Run certsuite ---"
    ./certsuite run \
      --label-filter="${LABEL_FILTER}" \
      --config-file="${CONFIG_FILE}" \
      --output-dir="${OUTPUT_DIR}" \
      --log-level="${LOG_LEVEL}" \
      --cleanup-probe=false || true
  fi

  # Validate
  if [[ $RC -eq 0 ]]; then
    echo "--- Validate ---"
    CLAIM_FILE="${OUTPUT_DIR}/claim.json"
    if [[ ! -f "$CLAIM_FILE" ]]; then
      echo "FAIL: claim.json not found at ${CLAIM_FILE}"
      RC=1
    else
      while IFS= read -r test_id; do
        validate_test "$CLAIM_FILE" "$test_id" "$(jq -c --arg id "$test_id" '.[$id]' <<<"$TESTS")" || RC=1
      done < <(jq -r 'keys_unsorted[]' <<<"$TESTS")
    fi
  fi

  # Print debug log on failure
  if [[ $RC -ne 0 ]]; then
    echo "--- Scenario certsuite.log (debug) ---"
    if [[ -f "certsuite.log" ]]; then
      grep -iE "${LABEL_FILTER//,/|}|tls|tlsversion|Probing|exec.*fallback|probe.*pod|probePods|DaemonSet|daemonset|cnf-suite|unsecured|plaintext|openssl|TLS probe" certsuite.log || tail -n 50 certsuite.log
    else
      echo "(certsuite.log not found)"
    fi
  fi

  # Cleanup (always runs)
  echo "--- Cleanup ---"
  cleanup_current_scenario

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
