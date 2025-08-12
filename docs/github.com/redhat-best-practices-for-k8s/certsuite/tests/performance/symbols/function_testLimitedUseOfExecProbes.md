testLimitedUseOfExecProbes`

**File:** `tests/performance/suite.go` (line 132)  
**Package:** `performance`

---

### Purpose
`testLimitedUseOfExecProbes` is a test helper that verifies the Kubernetes **exec probe** mechanism does not overwhelm system resources when used on a large number of containers.  
The function:

1. Iterates over all containers in a pod (or selected subset) and attaches an exec‑probe.
2. Monitors each probe’s output, collecting success/failure metrics.
3. Aggregates results into a `ReportObject` that is later stored by CertSuite.

It enforces the following limits:
- **Maximum number of probes** (`maxNumberOfExecProbes`)
- **Minimum interval between executions** (`minExecProbePeriodSeconds`)

If the pod does not contain any containers that satisfy the required CPU isolation guarantees, the function exits early with a warning (see `noProcessFoundErrMsg`).

---

### Signature
```go
func testLimitedUseOfExecProbes(check *checksdb.Check, env *provider.TestEnvironment)
```

| Parameter | Type                 | Description |
|-----------|----------------------|-------------|
| `check`   | `*checksdb.Check`    | Reference to the current check. Used for logging and result storage. |
| `env`     | `*provider.TestEnvironment` | Test environment that holds context such as pod metadata, logger, and configuration. |

---

### Key Steps & Dependencies

| Step | Operation | Called Functions / Types | Notes |
|------|-----------|--------------------------|-------|
| 1 | **Logging** | `LogInfo`, `LogError` | Provides visibility into test progress and errors. |
| 2 | **Probe Creation** | `NewContainerReportObject` (multiple times) | Builds a container‑level report entry for each exec probe. |
| 3 | **Result Aggregation** | `NewReportObject`, `SetResult` | Wraps per‑container results into a top‑level test report. |
| 4 | **Error Handling** | String formatting via `Sprintf` | Generates human‑readable messages for failures or missing containers. |

The function relies on constants defined earlier in the file:

- `maxNumberOfExecProbes`: upper bound of probes to launch.
- `minExecProbePeriodSeconds`: minimal sleep time between consecutive probe executions.
- `noProcessFoundErrMsg`: message printed when no suitable container is found.

No global state is modified; all data structures are created locally or passed through the `check` and `env` parameters.

---

### Inputs / Outputs

| Input | Effect |
|-------|--------|
| `check` | Holds a logger, configuration, and an interface to set the final test result. |
| `env` | Supplies pod/container information and execution context (e.g., namespace). |

| Output | Effect |
|--------|--------|
| **Logs** – via `LogInfo`/`LogError`. |
| **ReportObject** – appended to `check`’s report tree; contains per‑container probe status. |
| **Result Status** – `SetResult` marks the check as `Pass`, `Fail`, or `Skip`. |

---

### Side Effects

- No global variables are mutated.
- The function may write logs and update the provided `Check` object.
- It may terminate early if no suitable containers are found, returning a skipped status.

---

### How it Fits in the Package

`performance` contains end‑to‑end tests that validate Kubernetes workloads under various CPU isolation scenarios.  
`testLimitedUseOfExecProbes` is one of several helper functions invoked by high‑level test cases (e.g., “Guaranteed pod with exclusive CPUs”). It ensures that exec probes behave correctly and do not overload the system, contributing to the overall reliability score reported by CertSuite.

---

### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[Start] --> B{Containers Exist?}
    B -- No --> C[Log Skip & Return]
    B -- Yes --> D[Create Probe for Each Container]
    D --> E[Execute Probe (with min period)]
    E --> F[Collect Result]
    F --> G[Aggregate into ReportObject]
    G --> H[Set Check Result]
    H --> I[End]
```

This diagram illustrates the decision flow and key operations performed by `testLimitedUseOfExecProbes`.
