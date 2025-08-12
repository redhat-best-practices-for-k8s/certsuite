testContainerCertificationStatusByDigest`

| Aspect | Detail |
|--------|--------|
| **Package** | `certification` (`github.com/redhat-best-practices-for-k8s/certsuite/tests/certification`) |
| **Visibility** | Unexported – used only inside the test suite. |
| **Signature** | `func(*checksdb.Check, *provider.TestEnvironment, certdb.CertificationStatusValidator)()` |

### Purpose
Runs a single end‑to‑end certification test that validates the *container image* part of an operator’s Helm chart by using the image digest rather than the tag.  
It is intended to be used as a sub‑test inside the larger certification suite, where each call corresponds to one operator instance (or one specific image digest).

### Inputs

| Parameter | Type | Description |
|-----------|------|-------------|
| `*checksdb.Check` | A database record describing the current test run. It is passed around for reporting purposes. |
| `*provider.TestEnvironment` | Provides runtime information such as cluster state, Helm releases and operator data. |
| `certdb.CertificationStatusValidator` | Validates the certification status returned by the underlying check (`testContainerCertification`). |

### Workflow

1. **Logging**  
   The function logs an informational message that a container certification test is starting for the given digest.

2. **Report Object Creation**  
   A new `ContainerReportObject` is instantiated via `NewContainerReportObject`. Three fields are added:
   * `TestName`
   * `ImageDigest`
   * `OperatorNamespace`

3. **Inner Test Execution** (`testContainerCertification`)  
   The core certification logic is delegated to the helper function `testContainerCertification`, which performs the actual container image checks (signature verification, policy compliance, etc.) and returns a status object.

4. **Error Handling**  
   If `testContainerCertification` returns an error, it is logged and appended to the report object's errors slice.

5. **Result Setting**  
   The status returned by the inner test is stored in the report via `SetResult`.

6. **Final Reporting**  
   The completed report object (containing fields, any errors, and the result) is appended to a global slice of reports (via `append`).

### Key Dependencies

| Dependency | Role |
|------------|------|
| `LogInfo`, `LogError` | Structured logging for test progress and failures. |
| `NewContainerReportObject` | Factory that creates a report skeleton. |
| `AddField` | Adds metadata fields to the report. |
| `testContainerCertification` | Performs the actual certification logic. |
| `SetResult` | Stores the final status in the report. |
| `validator` (global) | Used indirectly by `testContainerCertification` to validate returned status. |

### Side Effects

* **Logging** – Emits log lines that can be captured by test harnesses.
* **Report Collection** – Adds a new container report object to an internal slice for later aggregation and output.
* **No state mutation** on the input objects beyond reading their data.

### Package Context

The `certification` package orchestrates end‑to‑end tests of operator Helm charts.  
`testContainerCertificationStatusByDigest` is one of several test helpers that:

1. Verify a container image satisfies certification requirements.
2. Record results in a structured report.
3. Feed those reports into the overall certification status evaluation.

Because it’s unexported, it’s only invoked from within `suite.go` (or other package‑internal helpers) as part of the suite's test matrix. The function exemplifies the pattern: **log → create report → run core check → record errors/results → append to global reports**.
