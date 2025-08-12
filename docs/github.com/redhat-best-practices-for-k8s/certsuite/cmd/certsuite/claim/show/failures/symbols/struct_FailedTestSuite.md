FailedTestSuite`

`FailedTestSuite` is the core data type used by the **failures** sub‑package of the Certsuite CLI to report which test suites contain at least one failing test case.

| Field | Type | Purpose |
|-------|------|---------|
| `TestSuiteName` | `string` | The name of the Kubernetes Test Suite (e.g., `"podsecuritypolicy"`).  It is used as a key when filtering results and for display in text or JSON output. |
| `FailingTestCases` | `[]FailedTestCase` | A slice containing every test case that failed within this suite.  Each element is a `FailedTestCase`, which holds the test‑case name, failure reason, and any non‑compliant objects extracted from the claim data. |

## How it is created

The helper function `getFailedTestCasesByTestSuite` (see `failures.go`) builds a slice of `FailedTestSuite`.  
Its workflow:

1. **Input** – two maps are supplied:
   * `claimResultsByTestSuite map[string][]*claim.TestCaseResult`
     - Parsed claim results keyed by suite name.
   * `targetTestSuites map[string]bool`
     - The set of suites that were actually executed in the current run.

2. **Filtering** – Only suites present in `targetTestSuites` are considered; others are discarded to avoid reporting on unused tests.

3. **Transformation** – For each qualifying suite, its list of `*claim.TestCaseResult`s is converted into a slice of `FailedTestCase`.  
   * Each test case that failed (i.e., has a non‑empty `FailureReason`) becomes a `FailedTestCase`.
   * The helper `getNonCompliantObjectsFromFailureReason` extracts any objects mentioned in the failure reason.

4. **Return** – A slice of fully populated `FailedTestSuite`s is returned to the caller, which then forwards it to the output routines (`printFailuresText`, `printFailuresJSON`).

## Dependencies

* **claim.TestCaseResult** – Raw data parsed from a claim file; contains fields like `FailureReason`.
* **FailedTestCase** – The struct that represents an individual failing test case (defined in the same package).
* **getNonCompliantObjectsFromFailureReason** – Parses failure reasons to pull out non‑compliant objects.
* **targetTestSuites map[string]bool** – Provided by the CLI’s command execution context; dictates which suites are relevant.

## Side effects

`FailedTestSuite` itself is a plain data container—no methods modify global state.  
The function that builds it (`getFailedTestCasesByTestSuite`) does not mutate its input maps, only reads them to produce new values.

## Usage in the package

1. **Parsing** – After claims are parsed, `getFailedTestCasesByTestSuite` is called to create a list of failing suites.
2. **Output** – The resulting slice is passed to either:
   * `printFailuresText` for human‑readable console output, or
   * `printFailuresJSON` for machine‑parseable JSON.
3. **CLI integration** – These functions are invoked by the `certsuite claim show failures` command to display a concise report of all test failures.

---

### Mermaid diagram (suggested)

```mermaid
flowchart TD
  A[claimResultsByTestSuite] -->|filter| B{targetTestSuites}
  B -->|yes| C[getFailedTestCasesByTestSuite]
  C --> D[[]FailedTestSuite]
  D --> E[printFailuresText / printFailuresJSON]
```

This diagram visualises the data flow from parsed claim results to the final failure report.
