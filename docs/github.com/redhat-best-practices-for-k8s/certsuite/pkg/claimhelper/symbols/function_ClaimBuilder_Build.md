ClaimBuilder.Build` – Documentation

## Purpose
`Build` creates a JSON claim file from the current test run state and writes it to disk.  
The method is typically called once per test suite execution (e.g., in a `TestMain` or finalizer) after all feature validations have been performed.

---

## Signature

```go
func (cb ClaimBuilder) Build() func(string)
```

* **Receiver** – `ClaimBuilder` holds the configuration for claim creation (output path, report key, etc.).  
* **Return value** – a closure that accepts a string (usually a test name or ID). The returned function is invoked to *record* individual test results into the claim structure before the final write.

---

## Workflow

| Step | Action | Key Functions |
|------|--------|---------------|
| 1 | Capture current time | `time.Now()`, `UTC()` |
| 2 | Format timestamp with layout defined by `DateTimeFormatDirective` | `Format(DateTimeFormatDirective)` |
| 3 | Retrieve reconciled test results (e.g., from an in‑memory store) | `GetReconciledResults()` |
| 4 | Build a claim object that contains: <br>• The timestamp <br>• A map of test IDs to result statuses <br>• Additional metadata keyed by `CNFFeatureValidationReportKey` | `MarshalClaimOutput(claim)` |
| 5 | Persist the claim JSON to disk at the path stored in `cb.outputPath` (file permissions set via `claimFilePermissions`) | `WriteClaimOutput()` |
| 6 | Log a success message | `Info()` |

---

## Inputs & Outputs

* **Inputs** – The method uses no explicit parameters; all required data is read from the `ClaimBuilder` instance and global state:
  * `cb.outputPath` – destination file path for the claim JSON.
  * Global map of reconciled results (populated during test execution).

* **Outputs** –  
  * A closure that accepts a string argument. When called, it records the status of a single test in the claim structure.
  * On completion, a JSON file is written to `cb.outputPath`. The file contains:
    ```json
    {
      "timestamp": "<UTC ISO8601>",
      "reportKey": "<CNFFeatureValidationReportKey>",
      "results": { "<testID>": "<status>" }
    }
    ```
  * A log entry indicating successful claim creation.

---

## Side Effects

| Effect | Description |
|--------|-------------|
| File I/O | Writes the claim JSON to disk; uses `claimFilePermissions` for file mode. |
| Logging | Emits an informational message upon successful write (`Info`). |
| State mutation | The returned closure mutates the internal claim map; subsequent calls update test results before final serialization. |

---

## Dependencies

* **Time utilities** – `time.Now()`, `UTC()`, and `Format(DateTimeFormatDirective)` format timestamps.
* **Claim generation helpers**
  * `GetReconciledResults()` – collects all test results that have been reconciled during the run.
  * `MarshalClaimOutput(claim)` – serializes the claim struct to JSON bytes.
  * `WriteClaimOutput(bytes, path)` – writes those bytes to disk with appropriate permissions.
* **Logging** – `Info()` logs success messages.

---

## Package Context

`claimhelper` is responsible for converting test run metadata into a machine‑readable “claim” that can be uploaded or inspected by external tooling.  
`Build` sits at the end of the pipeline: after tests have been executed and results reconciled, it produces the final artifact.  

```mermaid
graph LR
  A[Test Run] --> B[Feature Validation]
  B --> C[Reconciled Results (global map)]
  C --> D{ClaimBuilder.Build}
  D --> E[Marshal Claim Output]
  E --> F[Write Claim to Disk]
```

---

### Quick Usage

```go
cb := claimhelper.NewClaimBuilder("/tmp/claim.json")
recordTestResult := cb.Build()          // get the recorder closure
recordTestResult("test-foo")            // record a single test result
// ... after all tests:
_ = recordTestResult("")                // final call to trigger write (or use Build’s internal logic)
```

> **Note:** The actual `ClaimBuilder` struct definition is not shown here, but it must expose an `outputPath` field and any other configuration required by the helper functions above.
