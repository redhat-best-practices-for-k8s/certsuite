#!/usr/bin/env bash
set -euo pipefail

# Run all test scenarios defined in scenarios.json.
# Each scenario deploys its workloads, runs certsuite over the test IDs it
# covers, validates the claim against the expected state, and cleans up.
#
# Usage: run-scenarios.sh [--only <substring>] [--timings-file <path>] [--jobs <n>]
#
#   --only           Run only the scenarios whose name or path contains the
#                    substring. Handy while iterating on one scenario.
#   --timings-file   Where to write the per-scenario timing table
#                    (default: ${TMPDIR:-/tmp}/scenario-timings.tsv).
#   --jobs           How many lanes to run at once (default: 1). Entries in a
#                    lane always run serially; the lane field in scenarios.json
#                    decides which entries can share the cluster. See
#                    CONTRIBUTING.md for the lane rules.
#
# Env:   SMOKE_TESTS_LOG_LEVEL (default: info)
#        SMOKE_TESTS_JOBS      (default: 1, overridden by --jobs)
#
# Requires: jq, kubectl, and a ./certsuite binary built from this checkout.
# See CONTRIBUTING.md for the scenario format and for debugging a failure.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCENARIOS_FILE="${SCRIPT_DIR}/scenarios.json"
JUNIT_FILE_NAME="certsuite-tests_junit.xml"
LOG_LEVEL="${SMOKE_TESTS_LOG_LEVEL:-info}"
OVERALL_RC=0
CURRENT_SCENARIO_DIR=""
CURRENT_OUTPUT_DIR=""
ONLY_FILTER=""
JOBS="${SMOKE_TESTS_JOBS:-1}"
TIMINGS_FILE="${TMPDIR:-/tmp}/scenario-timings.tsv"
CONFIG_DIR="${TMPDIR:-/tmp}/certsuite-smoke-configs"
SCENARIO_REAP_FILE="${TMPDIR:-/tmp}/certsuite-smoke-reap-$$.txt"
LANE_PIDS=()
LANE_NAMES=()
export SCENARIO_REAP_FILE
: >"$SCENARIO_REAP_FILE"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --only)
      ONLY_FILTER="${2:?--only needs a value}"
      shift 2
      ;;
    --timings-file)
      TIMINGS_FILE="${2:?--timings-file needs a value}"
      shift 2
      ;;
    --jobs)
      JOBS="${2:?--jobs needs a value}"
      shift 2
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 2
      ;;
  esac
done

if ! [[ "$JOBS" =~ ^[0-9]+$ ]] || [[ "$JOBS" -lt 1 ]]; then
  echo "--jobs must be a positive integer, got '${JOBS}'" >&2
  exit 2
fi

# Fold each phase into a collapsible group in the GitHub Actions log.
group_start() {
  if [[ "${GITHUB_ACTIONS:-}" == "true" ]]; then
    echo "::group::$1"
  fi
}

group_end() {
  if [[ "${GITHUB_ACTIONS:-}" == "true" ]]; then
    echo "::endgroup::"
  fi
}

cleanup_current_scenario() {
  local scenario_dir="$CURRENT_SCENARIO_DIR"
  local output_dir="$CURRENT_OUTPUT_DIR"
  CURRENT_SCENARIO_DIR=""
  CURRENT_OUTPUT_DIR=""

  if [[ -n "$scenario_dir" ]]; then
    "$scenario_dir/cleanup.sh" "$output_dir" || true
  fi
}

# The base namespace a scenario group declares in its certsuite-config.yaml.
base_namespace() {
  awk '/^targetNameSpaces:/{found=1; next} found && /^  - name: /{print $3; exit}' "$1"
}

# Every entry gets its own namespace. Entries that share a path share the base
# namespace name, so a per-entry suffix keeps them apart; without it the next
# entry would try to create a namespace that is still terminating. The base is
# preserved as a prefix because some checks match on it (the reserved istio-
# prefix, for example).
# Keep this in sync with scenario_namespace() in generate-readme.py.
namespace_for() {
  local base=$1 output_dir=$2 suffix
  suffix=$(LC_ALL=C printf '%s' "${output_dir%-results}" | tr '[:upper:]' '[:lower:]' | tr -cs 'a-z0-9-' '-' | sed -e 's/^-//' -e 's/-$//' | cut -c1-30)
  suffix=${suffix%-}
  printf '%.63s' "${base}-${suffix}" | sed -e 's/-$//'
}

# Namespaces are deleted without waiting, so the suite reaps every one of them
# in parallel at the end instead of blocking for up to 120s per scenario.
reap_namespaces() {
  local namespace count
  [[ -s "$SCENARIO_REAP_FILE" ]] || return 0
  count=$(wc -l <"$SCENARIO_REAP_FILE" | tr -d '[:space:]')
  echo ""
  echo "--- Reaping ${count} namespace(s) ---"
  while IFS= read -r namespace; do
    kubectl wait --for=delete "namespace/${namespace}" --timeout=180s >/dev/null 2>&1 &
  done <"$SCENARIO_REAP_FILE"
  wait || true
  rm -f "$SCENARIO_REAP_FILE"
  echo "--- Namespaces reaped ---"
}

on_exit() {
  stop_lanes
  cleanup_current_scenario
  reap_namespaces
}

trap on_exit EXIT

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

# Errored results must carry a known errorType and an errorReason (not a
# skipReason); every other result must carry neither. The JUnit file must
# have exactly one <error> element per errored result. Keep the error types in
# sync with the ErrorType constants in pkg/checksdb/check.go.
validate_error_fields() {
  local claim_file=$1 junit_file=$2
  local output violations claim_errors junit_errors

  # The first output line is the errored-result count; any further lines are
  # violations.
  if ! output=$(jq -r '
    ["probe-exec-failure", "check-panic", "check-error", "lifecycle-panic", "lifecycle-error", "parallel-panic"] as $types
    | .claim.results
    | ([.[] | select(.state == "error")] | length),
      (to_entries[]
      | .key as $id
      | .value
      | if .state == "error" then
          (select((.errorType // "") | IN($types[]) | not) | "\($id): error result has unknown errorType \(.errorType // "" | @json)"),
          (select((.errorReason // "") == "") | "\($id): error result has no errorReason"),
          (select((.skipReason // "") != "") | "\($id): error result also sets skipReason")
        else
          select(has("errorType") or has("errorReason")) | "\($id): \(.state) result carries errorType/errorReason"
        end)
  ' "$claim_file"); then
    echo "FAIL: could not parse ${claim_file}"
    return 1
  fi
  claim_errors=$(head -n1 <<<"$output")
  violations=$(tail -n +2 <<<"$output")
  if [[ -n "$violations" ]]; then
    echo "FAIL: claim error fields are inconsistent:"
    echo "$violations"
    return 1
  fi

  if [[ ! -f "$junit_file" ]]; then
    echo "FAIL: JUnit file not found at ${junit_file}"
    return 1
  fi
  junit_errors=$(grep -cE '<error[ >]' "$junit_file" || true)
  if [[ "$junit_errors" -ne "$claim_errors" ]]; then
    echo "FAIL: JUnit has ${junit_errors} <error> element(s), claim has ${claim_errors} errored result(s)"
    return 1
  fi
  echo "PASS: error fields consistent across claim and JUnit (${claim_errors} errored result(s))"
}

SCENARIO_INDICES=$(jq -r --arg only "$ONLY_FILTER" '
  to_entries[]
  | select($only == "" or (.value.name | contains($only)) or (.value.path | contains($only)))
  | .key' "$SCENARIOS_FILE")

if [[ -z "$SCENARIO_INDICES" ]]; then
  echo "No scenario matches '${ONLY_FILTER}'" >&2
  exit 1
fi

scenario_count=$(wc -w <<<"$SCENARIO_INDICES" | tr -d '[:space:]')
echo "=== Running ${scenario_count} test scenario(s) with ${JOBS} lane(s) ==="
printf 'scenario\tdeploy\trun\tvalidate\tcleanup\ttotal\n' >"$TIMINGS_FILE"

# The lane an entry belongs to decides what it may run beside. Entries without
# a lane are namespace-scoped and can run beside anything; see CONTRIBUTING.md.
lane_for() {
  jq -r ".[$1].lane // \"namespaced\"" "$SCENARIOS_FILE"
}

# Deploy, run, validate and clean up one entry. Called directly in the serial
# path and from a per-lane subshell otherwise, so it must not depend on state
# another scenario leaves behind. Writes its timings to $TIMINGS_FILE, which a
# lane points at its own file.
run_scenario() {
  local i=$1
  local NAME SCENARIO_PATH MANIFEST OUTPUT_DIR TESTS LABEL_FILTER
  local SCENARIO_DIR CONFIG_FILE RENDERED_CONFIG CLAIM_FILE TEST_ID
  local RC=0 SCENARIO_START PHASE_START
  local DEPLOY_SECONDS=0 RUN_SECONDS=0 VALIDATE_SECONDS=0 CLEANUP_SECONDS=0

  NAME=$(jq -r ".[$i].name" "$SCENARIOS_FILE")
  SCENARIO_PATH=$(jq -r ".[$i].path" "$SCENARIOS_FILE")
  MANIFEST=$(jq -r ".[$i].manifest" "$SCENARIOS_FILE")
  OUTPUT_DIR=$(jq -r ".[$i].output_dir" "$SCENARIOS_FILE")
  # A scenario either checks one test (label_filter + expected_result) or a
  # batch of tests in one certsuite run (tests: {id: expectations}).
  TESTS=$(jq -c ".[$i] | .tests // {(.label_filter): del(.name, .description, .label_filter, .path, .manifest, .output_dir, .lane)}" "$SCENARIOS_FILE")
  LABEL_FILTER=$(jq -r 'keys_unsorted | join(",")' <<<"$TESTS")

  SCENARIO_DIR="${SCRIPT_DIR}/${SCENARIO_PATH}"
  CONFIG_FILE="${SCENARIO_DIR}/manifests/certsuite-config.yaml"
  CURRENT_SCENARIO_DIR="$SCENARIO_DIR"
  CURRENT_OUTPUT_DIR="$OUTPUT_DIR"

  # Give the entry its own namespace and point the config at it. Scenarios that
  # manage several namespaces of their own (the operator batches) declare
  # fixed_namespace and keep their config and scripts as written.
  if [[ "$(jq -r ".[$i].fixed_namespace // false" "$SCENARIOS_FILE")" == "true" ]]; then
    RENDERED_CONFIG="$CONFIG_FILE"
    unset SCENARIO_NAMESPACE
  else
    SCENARIO_NAMESPACE=$(namespace_for "$(base_namespace "$CONFIG_FILE")" "$OUTPUT_DIR")
    export SCENARIO_NAMESPACE
    mkdir -p "$CONFIG_DIR"
    RENDERED_CONFIG="${CONFIG_DIR}/${OUTPUT_DIR}.yaml"
    # Replace the whole line: sed's 0-address range only matches the prefix, and
    # BSD sed does not support the 0 address anyway.
    awk -v ns="$SCENARIO_NAMESPACE" '
      !done && /^  - name: / { print "  - name: " ns; done = 1; next }
      { print }
    ' "$CONFIG_FILE" >"$RENDERED_CONFIG"
  fi

  echo ""
  echo "========================================"
  echo "Scenario: ${NAME}"
  echo "  Manifest: ${MANIFEST}"
  if [[ -n "${SCENARIO_NAMESPACE:-}" ]]; then
    echo "  Namespace: ${SCENARIO_NAMESPACE}"
  fi
  jq -r 'to_entries[] | "  Expect:   \(.key) -> \(.value.expected_result)"' <<<"$TESTS"
  echo "========================================"

  SCENARIO_START=$SECONDS

  # Deploy
  group_start "Deploy: ${NAME}"
  echo "--- Deploy ---"
  PHASE_START=$SECONDS
  if ! "${SCENARIO_DIR}/deploy.sh" "${MANIFEST}"; then
    echo "FAIL: deploy failed for scenario '${NAME}'"
    RC=1
  fi
  DEPLOY_SECONDS=$((SECONDS - PHASE_START))
  group_end

  # Run certsuite (allow failure since some scenarios expect it).
  # Keep the probe daemonset between scenarios: deleting it waits for the
  # probe pods to terminate, which costs ~40s per scenario.
  if [[ $RC -eq 0 ]]; then
    group_start "certsuite: ${NAME}"
    echo "--- Run certsuite ---"
    PHASE_START=$SECONDS
    ./certsuite run \
      --label-filter="${LABEL_FILTER}" \
      --config-file="${RENDERED_CONFIG}" \
      --output-dir="${OUTPUT_DIR}" \
      --create-xml-junit-file \
      --log-level="${LOG_LEVEL}" \
      --cleanup-probe=false \
      --quiet || true
    RUN_SECONDS=$((SECONDS - PHASE_START))
    group_end
  fi

  # Validate
  if [[ $RC -eq 0 ]]; then
    group_start "Validate: ${NAME}"
    echo "--- Validate ---"
    PHASE_START=$SECONDS
    CLAIM_FILE="${OUTPUT_DIR}/claim.json"
    if [[ ! -f "$CLAIM_FILE" ]]; then
      echo "FAIL: claim.json not found at ${CLAIM_FILE}"
      RC=1
    else
      while IFS= read -r TEST_ID; do
        validate_test "$CLAIM_FILE" "$TEST_ID" "$(jq -c --arg id "$TEST_ID" '.[$id]' <<<"$TESTS")" || RC=1
      done < <(jq -r 'keys_unsorted[]' <<<"$TESTS")
      validate_error_fields "$CLAIM_FILE" "${OUTPUT_DIR}/${JUNIT_FILE_NAME}" || RC=1
    fi
    VALIDATE_SECONDS=$((SECONDS - PHASE_START))
    group_end
  fi

  # Print debug log on failure. A lane writes its own certsuite.log, so point
  # at that instead of whatever the previous entry left in the repo root.
  if [[ $RC -ne 0 ]]; then
    echo "--- Scenario certsuite.log (debug) ---"
    if [[ -f "certsuite.log" ]]; then
      grep -iE "${LABEL_FILTER//,/|}|tls|tlsversion|Probing|exec.*fallback|probe.*pod|probePods|DaemonSet|daemonset|cnf-suite|unsecured|plaintext|openssl|TLS probe" certsuite.log || tail -n 50 certsuite.log
    else
      echo "(certsuite.log not found)"
    fi
  fi

  # Cleanup (always runs)
  group_start "Cleanup: ${NAME}"
  echo "--- Cleanup ---"
  PHASE_START=$SECONDS
  cleanup_current_scenario
  CLEANUP_SECONDS=$((SECONDS - PHASE_START))
  group_end

  printf '%s\t%s\t%s\t%s\t%s\t%s\n' \
    "$NAME" "$DEPLOY_SECONDS" "$RUN_SECONDS" "$VALIDATE_SECONDS" "$CLEANUP_SECONDS" \
    "$((SECONDS - SCENARIO_START))" >>"$TIMINGS_FILE"

  if [[ $RC -ne 0 ]]; then
    echo "Scenario '${NAME}' FAILED."
    return 1
  fi
  echo "Scenario '${NAME}' passed."
  return 0
}

# --- Serial run -------------------------------------------------------------

if [[ "$JOBS" -eq 1 ]]; then
  for i in $SCENARIO_INDICES; do
    run_scenario "$i" || OVERALL_RC=1
  done
fi

# --- Parallel run -----------------------------------------------------------
#
# Every lane runs in a subshell so its SCENARIO_NAMESPACE, rendered config and
# timings stay its own. Entries inside a lane stay serial. Lanes are started
# longest first, which is what keeps the suite from ending with one lane still
# working.

lane_output() {
  local lane=$1
  local log="${LANE_LOG_DIR}/${lane}.log"
  echo ""
  echo "=== Lane '${lane}' output ==="
  if [[ -f "$log" ]]; then
    cat "$log"
  fi
}

lane_finished() {
  local pid=$1
  ! kill -0 "$pid" 2>/dev/null
}

# Print the output of every lane that has finished and forget its pid. Polling
# instead of `wait -n` keeps this working on the bash 3.2 that ships with macOS.
collect_finished_lanes() {
  local idx new_pids=() new_names=()
  for idx in "${!LANE_PIDS[@]}"; do
    if lane_finished "${LANE_PIDS[$idx]}"; then
      wait "${LANE_PIDS[$idx]}" || true
      lane_output "${LANE_NAMES[$idx]}"
    else
      new_pids+=("${LANE_PIDS[$idx]}")
      new_names+=("${LANE_NAMES[$idx]}")
    fi
  done
  if [[ ${#new_pids[@]} -gt 0 ]]; then
    LANE_PIDS=("${new_pids[@]}")
    LANE_NAMES=("${new_names[@]}")
  else
    LANE_PIDS=()
    LANE_NAMES=()
  fi
}

stop_lanes() {
  local pid
  if [[ ${#LANE_PIDS[@]} -eq 0 ]]; then
    return 0
  fi
  for pid in "${LANE_PIDS[@]}"; do
    kill "$pid" 2>/dev/null || true
    wait "$pid" 2>/dev/null || true
  done
  LANE_PIDS=()
  LANE_NAMES=()
}

run_lane() {
  local lane=$1 indices=$2 i rc=0
  TIMINGS_FILE="${LANE_TIMINGS_DIR}/${lane}.tsv"
  printf 'scenario\tdeploy\trun\tvalidate\tcleanup\ttotal\n' >"$TIMINGS_FILE"
  for i in $indices; do
    run_scenario "$i" || rc=1
  done
  echo "$rc" >"${LANE_RC_DIR}/${lane}"
}

# The probe DaemonSet is created by whichever certsuite run gets there first.
# Two lanes starting at the same time can therefore race on it, so the first
# lane is given a head start and the others wait for the DaemonSet to exist.
probe_present() {
  kubectl get daemonset -n cnf-suite certsuite-probe >/dev/null 2>&1
}

wait_for_probe() {
  local waited=0
  while ! probe_present; do
    if [[ "$waited" -ge 90 ]]; then
      echo "WARNING: cnf-suite/certsuite-probe did not appear after 90s; starting the other lanes anyway." >&2
      return 0
    fi
    sleep 1
    waited=$((waited + 1))
  done
}

if [[ "$JOBS" -gt 1 ]]; then
  LANE_LOG_DIR="${TMPDIR:-/tmp}/certsuite-smoke-lanes-$$"
  LANE_TIMINGS_DIR="${LANE_LOG_DIR}/timings"
  LANE_RC_DIR="${LANE_LOG_DIR}/rc"
  mkdir -p "$LANE_TIMINGS_DIR" "$LANE_RC_DIR"

  # Distinct lanes, longest first, so the longest lane is never the one that
  # starts last.
  LANE_ORDER=$(
    for lane in $(for i in $SCENARIO_INDICES; do lane_for "$i"; done | sort -u); do
      count=0
      for i in $SCENARIO_INDICES; do
        [[ "$(lane_for "$i")" == "$lane" ]] && count=$((count + 1))
      done
      printf '%s %s\n' "$count" "$lane"
    done | sort -rn | awk '{print $2}'
  )

  # Work is handed out in units. A constrained lane is a single unit: its
  # entries must not overlap anything, including each other. The namespaced
  # lane is the one that is safe to overlap freely, so it is spread round-robin
  # over as many shards as there are workers, which keeps a long lane from
  # leaving the other workers idle.
  UNIT_NAMES=()
  UNIT_INDICES=()
  NAMESPACED_ENTRIES=""

  for lane in $LANE_ORDER; do
    if [[ "$lane" == "namespaced" ]]; then
      continue
    fi
    indices=""
    for i in $SCENARIO_INDICES; do
      if [[ "$(lane_for "$i")" == "$lane" ]]; then
        indices="${indices} $i"
      fi
    done
    UNIT_NAMES+=("$lane")
    UNIT_INDICES+=("$indices")
  done

  for i in $SCENARIO_INDICES; do
    if [[ "$(lane_for "$i")" == "namespaced" ]]; then
      NAMESPACED_ENTRIES="${NAMESPACED_ENTRIES} $i"
    fi
  done
  namespaced_count=$(wc -w <<<"$NAMESPACED_ENTRIES" | tr -d '[:space:]')
  if [[ "$namespaced_count" -gt 0 ]]; then
    shards=$JOBS
    if [[ "$shards" -gt "$namespaced_count" ]]; then
      shards=$namespaced_count
    fi
    shard_index=0
    while [[ "$shard_index" -lt "$shards" ]]; do
      shard_entries[shard_index]=""
      shard_index=$((shard_index + 1))
    done
    position=0
    for i in $NAMESPACED_ENTRIES; do
      shard=$((position % shards))
      shard_entries[shard]="${shard_entries[shard]} $i"
      position=$((position + 1))
    done
    shard_number=1
    while [[ "$shard_number" -le "$shards" ]]; do
      UNIT_NAMES+=("namespaced-${shard_number}")
      UNIT_INDICES+=("${shard_entries[shard_number - 1]}")
      shard_number=$((shard_number + 1))
    done
  fi

  echo "Units: ${UNIT_NAMES[*]}"
  first_unit=true
  unit=0
  while [[ "$unit" -lt "${#UNIT_NAMES[@]}" ]]; do
    lane="${UNIT_NAMES[$unit]}"
    LANE_INDICES="${UNIT_INDICES[$unit]}"

    while [[ ${#LANE_PIDS[@]} -ge $JOBS ]]; do
      collect_finished_lanes
      if [[ ${#LANE_PIDS[@]} -ge $JOBS ]]; then
        sleep 2
      fi
    done

    (
      trap 'cleanup_current_scenario' EXIT
      run_lane "$lane" "$LANE_INDICES"
    ) >"${LANE_LOG_DIR}/${lane}.log" 2>&1 &
    LANE_PIDS+=($!)
    LANE_NAMES+=("$lane")

    if [[ "$first_unit" == "true" ]] && ! probe_present; then
      echo "Waiting for the first unit to create the probe DaemonSet..."
      wait_for_probe
    fi
    first_unit=false
    unit=$((unit + 1))
  done

  while [[ ${#LANE_PIDS[@]} -gt 0 ]]; do
    collect_finished_lanes
    if [[ ${#LANE_PIDS[@]} -gt 0 ]]; then
      sleep 2
    fi
  done

  # Merge the per-unit timing files under the header written above.
  for lane in "${UNIT_NAMES[@]}"; do
    tail -n +2 "${LANE_TIMINGS_DIR}/${lane}.tsv" >>"$TIMINGS_FILE"
    lane_rc=$(cat "${LANE_RC_DIR}/${lane}" 2>/dev/null || echo 1)
    if [[ "$lane_rc" != "0" ]]; then
      OVERALL_RC=1
    fi
  done
  rm -rf "$LANE_LOG_DIR"
fi

print_timings() {
  [[ -s "$TIMINGS_FILE" ]] || return 0
  echo ""
  echo "=== Scenario timings (seconds) ==="
  # Sort by total (last column), descending, keeping the header first.
  {
    head -n 1 "$TIMINGS_FILE"
    tail -n +2 "$TIMINGS_FILE" | sort -t "$(printf '\t')" -k6,6 -rn
  } >"${TIMINGS_FILE}.sorted"
  if command -v column >/dev/null 2>&1; then
    column -t -s "$(printf '\t')" "${TIMINGS_FILE}.sorted"
  else
    cat "${TIMINGS_FILE}.sorted"
  fi
  rm -f "${TIMINGS_FILE}.sorted"
  echo "(written to ${TIMINGS_FILE})"
}

reap_namespaces

print_timings

echo ""
if [[ $OVERALL_RC -ne 0 ]]; then
  echo "=== One or more scenarios FAILED ==="
  exit 1
fi
echo "=== All scenarios passed ==="
