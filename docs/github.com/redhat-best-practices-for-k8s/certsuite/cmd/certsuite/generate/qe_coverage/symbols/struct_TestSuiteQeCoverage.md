TestSuiteQeCoverage` – QE Coverage Summary

| Field | Type | Purpose |
|-------|------|---------|
| `Coverage` | `float32` | Percentage of test cases in the suite that have been mapped to a QE (Quality‑Engineering) case ID. Calculated as `(TestCasesWithQe / TestCases) * 100`. |
| `NotImplementedTestCases` | `[]string` | List of test case identifiers that are **present** in the repository but **do not yet have an associated QE case**. These are the entries that will be surfaced to QA for implementation or linking. |
| `TestCases` | `int` | Total number of test cases discovered in the suite (regardless of QE mapping). |
| `TestCasesWithQe` | `int` | Number of test cases that already have a QE case ID linked. |

### What it represents

`TestSuiteQeCoverage` is a lightweight DTO used by the **qe_coverage** generator to produce a concise report on how well a test suite is covered by QE tracking. The generator scans all Go test files in a given package, extracts test identifiers, and cross‑references them against an external QE database (typically a CSV or API). After aggregation it populates this struct and writes the data to JSON/YAML for consumption by CI dashboards or documentation generators.

### Dependencies

| Dependency | Role |
|------------|------|
| `qe_coverage.go` (package `qecoverage`) | Contains the logic that builds the struct from raw test discovery data. |
| External QE source (e.g., `qe_cases.csv`) | Provides mapping of test names → QE case IDs. The generator reads this to determine `TestCasesWithQe`. |
| Standard library (`encoding/json`, `os`, etc.) | Serializes the struct and writes it to disk. |

### Side effects

* **File generation** – When the generator runs, a file named `<suite>-qe_coverage.json` (or similar) is written into the output directory.
* **No in‑memory mutation of tests** – The struct itself is immutable after creation; it merely reflects state at generation time.

### How it fits the package

```
cmd/certsuite/generate/qe_coverage/
├── qe_coverage.go        // scans test files, builds TestSuiteQeCoverage
└── output/               // JSON/YAML report written here
```

`TestSuiteQeCoverage` is the *single source of truth* for coverage metrics that other tools (CI pipelines, web dashboards) consume. It encapsulates both quantitative data (`Coverage`, `TestCases`, `TestCasesWithQe`) and actionable lists (`NotImplementedTestCases`). By keeping it simple and read‑only, the generator can be safely reused across multiple test suites without side effects.
