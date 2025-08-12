showQeCoverageForTestCaseName`

**Location**

```
cmd/certsuite/generate/qe_coverage/qe_coverage.go:61
package qecoverage
```

---

## Purpose

`showQeCoverageForTestCaseName` is an internal helper that formats and prints a coverage report for a single test case.  
It receives the name of the test case (`string`) and a `TestCoverageSummaryReport` (the type definition is elsewhere in the package).  The function then emits human‑readable output to standard out, summarizing which QE tests were run and how many were covered.

This routine is invoked by the command implementation (`qeCoverageReportCmd`) when generating coverage reports for individual test cases.

---

## Signature

```go
func showQeCoverageForTestCaseName(testCase string, report TestCoverageSummaryReport)
```

| Parameter | Type                   | Description |
|-----------|------------------------|-------------|
| `testCase`| `string`               | The identifier of the test case whose coverage is being shown. |
| `report`  | `TestCoverageSummaryReport` | A structure containing per‑QE‑test coverage counts and lists. |

The function has no return value; its effect is printing to stdout.

---

## Key Operations

1. **Header** – Prints a header line with the test case name.
2. **Total QE tests** – Displays the total number of QE tests that exist (`len(report.QETests)`).
3. **Covered vs. Uncovered** – Shows how many QE tests are covered by this test case and how many remain uncovered.
4. **List of Covered Tests** – If any tests were covered, it prints a comma‑separated list using `strings.Join`.
5. **Empty coverage handling** – If no tests are covered, it simply reports that the test case covers none.

The function relies only on standard library functions (`fmt.Println`, `fmt.Printf`, `len`, `strings.Join`) and the data inside `TestCoverageSummaryReport`. No global state is modified.

---

## Dependencies

| Dependency | Purpose |
|------------|---------|
| `fmt`      | Output formatting (Println/Printf). |
| `strings`  | Join covered test names into a single string. |
| `qeCoverageReportCmd` | The Cobra command that triggers this helper; it supplies the context but is not directly used by the function. |

---

## Side‑Effects

* Writes to standard output.
* Does **not** modify any global variables or mutate its arguments.

---

## How It Fits the Package

The `qecoverage` package builds a coverage report for QE (Quality Engineering) tests against certsuite test cases.  
`showQeCoverageForTestCaseName` is a small, focused routine that takes the pre‑computed summary (`TestCoverageSummaryReport`) and presents it to the user.  It is called by higher‑level command handlers when iterating over all test cases or when the user requests coverage for a specific case.

The function keeps output logic isolated from data processing logic, enabling easy testing of formatting without touching the report generation code.
