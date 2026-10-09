#!/usr/bin/env python3
"""Generate or freshness-check the isolated scenario coverage README."""

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

FOLLOW_UP_IDS = {
    "access-control-security-context-privilege-escalation",
    "access-control-security-context-read-only-root-file-system",
    "networking-reserved-partner-ports",
    "observability-termination-policy",
}

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


def load_scenarios(catalog_ids):
    scenarios = json.loads(SCENARIOS_FILE.read_text(encoding="utf-8"))
    if not isinstance(scenarios, list):
        raise ValueError("scenarios.json must contain a JSON array")

    output_dirs = set()
    scenarios_by_id = defaultdict(list)
    for index, scenario in enumerate(scenarios):
        prefix = f"scenario {index + 1}"
        for key in ("name", "label_filter", "path", "manifest", "expected_result", "output_dir"):
            if not scenario.get(key):
                raise ValueError(f"{prefix} is missing required field {key!r}")
        if scenario["expected_result"] not in {"passed", "failed"}:
            raise ValueError(f"{prefix} has invalid expected_result {scenario['expected_result']!r}")

        test_id = scenario["label_filter"]
        if test_id not in catalog_ids:
            raise ValueError(f"{prefix} filters unknown catalog ID {test_id!r}")

        output_dir = scenario["output_dir"]
        if output_dir in output_dirs:
            raise ValueError(f"Scenario output directory {output_dir!r} is reused")
        output_dirs.add(output_dir)

        scenario_dir = SCRIPT_DIR / scenario["path"]
        for required_file in (
            scenario_dir / "deploy.sh",
            scenario_dir / "cleanup.sh",
            scenario_dir / "manifests" / "certsuite-config.yaml",
            scenario_dir / "manifests" / scenario["manifest"],
        ):
            if not required_file.is_file():
                raise ValueError(f"{prefix} references missing file {required_file.relative_to(ROOT)}")

        if test_id in OBJECT_ASSERTION_IDS:
            assertion = "min_compliant" if scenario["expected_result"] == "passed" else "min_noncompliant"
            if int(scenario.get(assertion, 0)) < 1:
                raise ValueError(f"{prefix} must assert at least one {assertion} object")

        scenarios_by_id[test_id].append(scenario)

    return scenarios_by_id


def markdown_cell(value):
    return str(value).replace("|", "\\|").replace("\n", " ")


def generate_readme():
    suites, ids_by_suite = parse_catalog()
    catalog_ids = {test_id for ids in ids_by_suite.values() for test_id in ids}
    unknown_follow_ups = FOLLOW_UP_IDS - catalog_ids
    if unknown_follow_ups:
        raise ValueError(f"Follow-up IDs are missing from CATALOG.md: {', '.join(sorted(unknown_follow_ups))}")

    scenarios_by_id = load_scenarios(catalog_ids)
    lines = [
        "<!-- Generated by generate-readme.py; run make check-scenario-readme to check freshness. -->",
        "<!-- markdownlint-disable line-length no-inline-html -->",
        "# Isolated smoke scenario coverage",
        "",
        "This inventory lists every test ID in `CATALOG.md` and the dedicated entries in `scenarios.json`.",
        "Coverage here means an isolated workload scenario asserts the test's expected claim state and a relevant object.",
        "The broader pre-main `--label-filter=all` smoke run is separate and does not count as dedicated scenario coverage.",
        "",
        "The follow-up IDs screened for CNFCERT-1501 are marked `Follow-up`; all other uncovered catalog IDs are marked `No dedicated scenario`.",
        "",
    ]

    for suite in suites:
        lines.extend(
            [
                f"## {suite}",
                "",
                "| Catalog ID | Dedicated scenarios and expected results | Coverage |",
                "|---|---|---|",
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
                coverage = "Follow-up" if test_id in FOLLOW_UP_IDS else "No dedicated scenario"
            lines.append(f"| `{test_id}` | {scenario_text} | {coverage} |")
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
