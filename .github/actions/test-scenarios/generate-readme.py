#!/usr/bin/env python3
"""Generate or freshness-check the smoke scenario and certsuite-qe coverage README."""

import argparse
import json
import re
import sys
from collections import defaultdict
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parent
ROOT = SCRIPT_DIR.parents[2]
CATALOG_FILE = ROOT / "CATALOG.md"
SCENARIOS_FILE = SCRIPT_DIR / "scenarios.json"
README_FILE = SCRIPT_DIR / "README.md"
SUITE_KEYS_FILE = ROOT / "tests" / "common" / "constant.go"
QE_FLAG_FILES = sorted((ROOT / "tests" / "identifiers").glob("*.go")) + [ROOT / "tests" / "preflight" / "catalog.go"]
QE_FLAG_ARG = 6

# Catalog IDs that cannot get a scenario on the Kind smoke cluster, and why.
# Anything not listed here and not in scenarios.json is an open gap, not a
# deliberate one.
UNCOVERABLE_IDS = {
    "access-control-sys-nice-realtime-capability": "needs a node running a realtime kernel",
}

# Non-batch tests that must assert an object, not only the claim state. Batch
# tests are always required to; these are the single-test scenarios where a
# matching state alone would also be produced by an unrelated failure.
OBJECT_ASSERTION_IDS = {
    "manageability-container-port-name-format",
    "lifecycle-readiness-probe",
    "networking-network-policy-deny-all",
    "lifecycle-liveness-probe",
    "lifecycle-startup-probe",
    "manageability-containers-image-tag",
    "access-control-pod-automount-service-account-token",
    "networking-undeclared-container-ports-usage",
}


def parse_catalog():
    content = CATALOG_FILE.read_text(encoding="utf-8")
    lines = content.splitlines()
    expected_total = re.search(r"^### Total test cases: (\d+)\s*$", content, re.MULTILINE)
    if not expected_total:
        raise ValueError("CATALOG.md does not contain its total test case count")

    suites = []
    in_suite_table = False
    for line in lines:
        if line.strip() == "|Suite|Tests per suite|Link|":
            in_suite_table = True
            continue
        if in_suite_table and line.strip().startswith("|"):
            cells = [cell.strip() for cell in line.strip().strip("|").split("|")]
            if len(cells) >= 2 and cells[1].isdigit():
                suites.append(cells[0])
            continue
        if in_suite_table:
            in_suite_table = False

    if not suites:
        raise ValueError("Could not find the suite summary table in CATALOG.md")

    ids_by_suite = {suite: [] for suite in suites}
    unknown_ids = []
    for line in lines:
        match = re.match(r"^#### ([A-Za-z0-9][A-Za-z0-9-]*)\s*$", line)
        if not match:
            continue
        test_id = match.group(1)
        matches = [suite for suite in suites if test_id.startswith(f"{suite}-")]
        if not matches:
            unknown_ids.append(test_id)
            continue
        suite = max(matches, key=len)
        ids_by_suite[suite].append(test_id)

    if unknown_ids:
        raise ValueError(f"Catalog test IDs do not match a listed suite: {', '.join(unknown_ids)}")

    actual_total = sum(len(test_ids) for test_ids in ids_by_suite.values())
    if actual_total != int(expected_total.group(1)):
        raise ValueError(
            f"Catalog summary says {expected_total.group(1)} test cases, found {actual_total} IDs"
        )

    return suites, ids_by_suite


def split_call_args(source, start):
    """Return the top-level arguments of the Go call whose "(" ends at start."""
    args = []
    depth = 0
    arg_start = index = start
    while index < len(source):
        char = source[index]
        if char == "`":
            index = source.index("`", index + 1)
        elif char == '"':
            index += 1
            while source[index] != '"':
                index += 2 if source[index] == "\\" else 1
        elif char in "({[":
            depth += 1
        elif char in ")}]":
            if depth == 0:
                args.append(source[arg_start:index].strip())
                return args
            depth -= 1
        elif char == "," and depth == 0:
            args.append(source[arg_start:index].strip())
            arg_start = index + 1
        index += 1
    raise ValueError("Unterminated AddCatalogEntry call")


def parse_qe_flags(catalog_ids):
    """Read the qe argument of every AddCatalogEntry call.

    Calls with a non-literal test name (the generated preflight entries) set the
    flag for every catalog ID in their suite.
    """
    suite_keys = dict(re.findall(r'(\w+TestKey)\s*=\s*"([^"]+)"', SUITE_KEYS_FILE.read_text(encoding="utf-8")))
    qe_by_id = {}
    qe_by_suite = {}
    for go_file in QE_FLAG_FILES:
        source = go_file.read_text(encoding="utf-8")
        for match in re.finditer(r"\bAddCatalogEntry\(", source):
            if source[: match.start()].endswith("func "):
                continue
            args = split_call_args(source, match.end())
            suite = suite_keys.get(args[1].rsplit(".", 1)[-1])
            qe = args[QE_FLAG_ARG]
            if not suite or qe not in {"true", "false"}:
                raise ValueError(f"Cannot read the qe flag from an AddCatalogEntry call in {go_file.relative_to(ROOT)}")
            if args[0].startswith('"'):
                qe_by_id[suite + "-" + args[0].strip('"')] = qe == "true"
            else:
                qe_by_suite[suite] = qe == "true"

    flags = {}
    for test_id in catalog_ids:
        suite = next((suite for suite in qe_by_suite if test_id.startswith(f"{suite}-")), None)
        if test_id in qe_by_id:
            flags[test_id] = qe_by_id[test_id]
        elif suite:
            flags[test_id] = qe_by_suite[suite]
        else:
            raise ValueError(f"No qe flag found for catalog ID {test_id!r}")
    return flags


def load_scenarios(catalog_ids):
    scenarios = json.loads(SCENARIOS_FILE.read_text(encoding="utf-8"))
    if not isinstance(scenarios, list):
        raise ValueError("scenarios.json must contain a JSON array")

    output_dirs = set()
    scenarios_by_id = defaultdict(list)
    for index, scenario in enumerate(scenarios):
        prefix = f"scenario {index + 1}"
        for key in ("name", "description", "path", "manifest", "output_dir"):
            if not scenario.get(key):
                raise ValueError(f"{prefix} is missing required field {key!r}")

        # Batch scenarios check several tests in one run; the others check
        # label_filter against the expectations on the scenario itself.
        is_batch = "tests" in scenario
        if is_batch:
            tests = scenario["tests"]
            if not isinstance(tests, dict) or not tests:
                raise ValueError(f"{prefix} must map at least one test ID to its expectations in 'tests'")
        else:
            for key in ("label_filter", "expected_result"):
                if not scenario.get(key):
                    raise ValueError(f"{prefix} is missing required field {key!r}")
            tests = {scenario["label_filter"]: scenario}

        for test_id, expectations in tests.items():
            if test_id not in catalog_ids:
                raise ValueError(f"{prefix} filters unknown catalog ID {test_id!r}")
            expected = expectations.get("expected_result")
            if expected not in {"passed", "failed"}:
                raise ValueError(f"{prefix} has invalid expected_result {expected!r} for {test_id}")
            if is_batch or test_id in OBJECT_ASSERTION_IDS:
                if expected == "passed":
                    asserted = int(expectations.get("min_compliant", 0)) >= 1
                else:
                    asserted = bool(expectations.get("noncompliant_contains")) or int(expectations.get("min_noncompliant", 0)) >= 1
                if not asserted:
                    raise ValueError(f"{prefix} must assert at least one {'compliant' if expected == 'passed' else 'non-compliant'} object for {test_id}")
            scenarios_by_id[test_id].append({"name": scenario["name"], "expected_result": expected})

        output_dir = scenario["output_dir"]
        if output_dir in output_dirs:
            raise ValueError(f"Scenario output directory {output_dir!r} is reused")
        output_dirs.add(output_dir)
        scenario["test_ids"] = list(tests)

        scenario_dir = SCRIPT_DIR / scenario["path"]
        for required_file in (
            scenario_dir / "deploy.sh",
            scenario_dir / "cleanup.sh",
            scenario_dir / "manifests" / "certsuite-config.yaml",
            scenario_dir / "manifests" / scenario["manifest"],
        ):
            if not required_file.is_file():
                raise ValueError(f"{prefix} references missing file {required_file.relative_to(ROOT)}")

    return scenarios, scenarios_by_id


def markdown_cell(value):
    return str(value).replace("|", "\\|").replace("\n", " ")


def generate_readme():
    suites, ids_by_suite = parse_catalog()
    catalog_ids = {test_id for ids in ids_by_suite.values() for test_id in ids}
    unknown_uncoverable = set(UNCOVERABLE_IDS) - catalog_ids
    if unknown_uncoverable:
        raise ValueError(f"Uncoverable IDs are missing from CATALOG.md: {', '.join(sorted(unknown_uncoverable))}")

    scenarios, scenarios_by_id = load_scenarios(catalog_ids)
    covered_and_uncoverable = set(UNCOVERABLE_IDS) & set(scenarios_by_id)
    if covered_and_uncoverable:
        raise ValueError(f"Uncoverable IDs have a scenario: {', '.join(sorted(covered_and_uncoverable))}")

    qe_flags = parse_qe_flags(catalog_ids)
    lines = [
        "<!-- Generated by generate-readme.py; run make check-scenario-readme to check freshness. -->",
        "<!-- markdownlint-disable line-length no-inline-html -->",
        "# Smoke scenario and certsuite-qe coverage",
        "",
        "This inventory lists every test ID in `CATALOG.md` and compares two kinds of coverage:",
        "",
        "- **certsuite-qe**: the `qe` flag passed to `AddCatalogEntry` in `tests/identifiers` (and `tests/preflight/catalog.go`), which marks tests that have a [certsuite-qe](https://github.com/redhat-best-practices-for-k8s/certsuite-qe) test. `certsuite generate qe-coverage-report` reads the same flag.",
        "- **Smoke scenarios**: the dedicated entries in `scenarios.json`. Each one deploys workloads into an isolated namespace and asserts the expected claim state and a relevant object for each test it covers. Batch entries (`tests`) check several tests in one certsuite run.",
        "",
        "The broader pre-main `--label-filter=all` smoke run is separate and does not count as dedicated scenario coverage.",
        "Catalog IDs that cannot be exercised on the Kind smoke cluster say why; all other uncovered IDs are marked `No dedicated scenario` and are open gaps.",
        "",
        "To add a scenario, see [CONTRIBUTING.md](CONTRIBUTING.md). For the multi-test batch scenarios, see [batches/README.md](batches/README.md).",
        "",
        "## Summary",
        "",
        "| Suite | Tests | Covered in certsuite-qe | Covered in smoke scenarios | Covered in both | Covered in neither |",
        "|---|---|---|---|---|---|",
    ]

    totals = [0, 0, 0, 0, 0]
    for suite in suites:
        ids = ids_by_suite[suite]
        qe = {test_id for test_id in ids if qe_flags[test_id]}
        smoke = {test_id for test_id in ids if test_id in scenarios_by_id}
        counts = [len(ids), len(qe), len(smoke), len(qe & smoke), len(set(ids) - qe - smoke)]
        totals = [total + count for total, count in zip(totals, counts)]
        lines.append(f"| {suite} | " + " | ".join(str(count) for count in counts) + " |")
    lines.extend(["| **Total** | " + " | ".join(f"**{total}**" for total in totals) + " |", ""])

    lines.extend(
        [
            "## Scenarios",
            "",
            f"The {len(scenarios)} entries in `scenarios.json`, in the order `run-scenarios.sh` runs them.",
            "",
            "| Scenario | Manifest | Tests | What it deploys |",
            "|---|---|---|---|",
        ]
    )
    for scenario in scenarios:
        manifest = f"`{scenario['path']}/manifests/{scenario['manifest']}`"
        test_ids = scenario["test_ids"]
        count = str(len(test_ids)) if len(test_ids) > 1 else f"`{test_ids[0]}`"
        lines.append(
            f"| {markdown_cell(scenario['name'])} | {manifest} | {count} | {markdown_cell(scenario['description'])} |"
        )
    lines.append("")

    for suite in suites:
        lines.extend(
            [
                f"## {suite}",
                "",
                "| Catalog ID | Covered in certsuite-qe | Smoke scenarios and expected results | Covered in smoke scenarios |",
                "|---|---|---|---|",
            ]
        )
        for test_id in ids_by_suite[suite]:
            scenarios = scenarios_by_id.get(test_id, [])
            if scenarios:
                scenario_text = "<br>".join(
                    f"{markdown_cell(scenario['name'])} (`{scenario['expected_result']}`)"
                    for scenario in scenarios
                )
                coverage = "Covered"
            else:
                scenario_text = "—"
                reason = UNCOVERABLE_IDS.get(test_id)
                coverage = f"Not reproducible on the smoke cluster: {reason}" if reason else "No dedicated scenario"
            qe = "Yes" if qe_flags[test_id] else "No"
            lines.append(f"| `{test_id}` | {qe} | {scenario_text} | {coverage} |")
        lines.append("")

    return "\n".join(lines).rstrip() + "\n"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="fail if README.md is stale")
    args = parser.parse_args()

    generated = generate_readme()
    if args.check:
        if not README_FILE.exists() or README_FILE.read_text(encoding="utf-8") != generated:
            print("Scenario coverage README is stale. Run python3 .github/actions/test-scenarios/generate-readme.py.", file=sys.stderr)
            return 1
        print("Scenario coverage README is up to date.")
        return 0

    README_FILE.write_text(generated, encoding="utf-8")
    print(f"Wrote {README_FILE.relative_to(ROOT)}")
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (OSError, ValueError, json.JSONDecodeError) as error:
        print(error, file=sys.stderr)
        sys.exit(1)
