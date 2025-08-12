TestCaseResult`

| Field | Type | Purpose |
|-------|------|---------|
| `CapturedTestOutput` | `string` | Full stdout/stderr captured during the test run. Used for debugging and reporting. |
| `CatalogInfo` | struct | Metadata that ties a test to the Red‑Hat best‑practice catalog.  <br>• `BestPracticeReference`: ID of the relevant best practice.<br>• `Description`: Human‑readable description of the check.<br>• `ExceptionProcess`: Note on any known exceptions.<br>• `Remediation`: Suggested fix if the test fails. |
| `CategoryClassification` | `map[string]string` | Key/value pairs classifying the test (e.g., `"severity":"high"`, `"compliance":"NIST"`).  These are used to group results in reports and dashboards. |
| `CheckDetails` | `string` | Human‑readable description of what the test actually verifies. |
| `Duration` | `int` | Time taken to execute the test, in milliseconds. |
| `EndTime` | `string` | ISO‑8601 timestamp when the test finished. |
| `FailureLineContent` | `string` | If the test failed, the exact line from the log that caused the failure. |
| `FailureLocation` | `string` | File or resource path where the failure occurred (e.g., a YAML file). |
| `SkipReason` | `string` | Reason why the test was skipped (if `State == "skipped"`). |
| `StartTime` | `string` | ISO‑8601 timestamp when the test started. |
| `State` | `string` | One of `"passed"`, `"failed"`, or `"skipped"`.  Determines how the result is processed downstream. |
| `TestID` | struct | Unique identifier for the test case. <br>• `ID`: Test ID (e.g., `TC001`).<br>• `Suite`: The suite name that owns the test.<br>• `Tags`: Comma‑separated tags for filtering or grouping. |

---

### Purpose

`TestCaseResult` is a **data transfer object** used by the *claim* package to aggregate all information produced during the execution of an individual compliance test.  It is serialized (typically to JSON) and stored in the claims database so that other components—report generators, dashboards, or external consumers—can consume a complete snapshot of what happened for each test.

### Inputs / Outputs

| Phase | Action | Effect |
|-------|--------|--------|
| **Test execution** | A test runner calls `runTest(...)` (not shown in the snippet). | Populates fields such as `StartTime`, `Duration`, and `CapturedTestOutput`. |
| **Result handling** | After a test finishes, the runner creates a `TestCaseResult{}` instance. | Fields like `State`, `FailureLineContent`, or `SkipReason` are set based on the outcome. |
| **Persistence** | The result is marshalled to JSON and written to the claims store (e.g., etcd or a local file). | Provides durable storage for audit and reporting. |

### Key Dependencies

* **Logging / capture** – The runner must intercept stdout/stderr; these bytes become `CapturedTestOutput`.
* **Catalog data source** – The mapping from test ID → catalog info comes from the best‑practice catalog, usually loaded at startup.
* **Timing utilities** – `time.Now()` and duration calculations fill `StartTime`, `EndTime`, and `Duration`.
* **Failure detection logic** – Determines if a test failed or was skipped, setting `State` and associated failure fields.

### Side Effects

* Writing the struct to disk/etcd is the primary side effect; it does not modify any global state beyond that.
* The struct itself is immutable after construction (no exported setters), so once created it can be safely shared across goroutines.

---

## How It Fits in `claim`

The *claim* package orchestrates compliance checks and claim generation.  
`TestCaseResult` is the central artifact representing **one** test outcome.  During a run, the package:

1. Loads all test cases (from YAML or embedded Go code).
2. Executes each case, producing a `TestCaseResult`.
3. Aggregates many such results into a *claim* that can be uploaded to an external system.

Thus, this struct is the bridge between raw test execution and higher‑level claim logic.  It enables downstream consumers (reporters, CI pipelines) to query, filter, or display detailed compliance information without needing access to the internals of the test runner.
