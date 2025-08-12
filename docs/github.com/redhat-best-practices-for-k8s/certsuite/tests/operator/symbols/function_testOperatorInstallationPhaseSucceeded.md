testOperatorInstallationPhaseSucceeded`

### Purpose
`testOperatorInstallationPhaseSucceeded` is an internal test helper used by the **operator** test suite to verify that a Cert‑Suite Operator instance reaches the *Ready* state after installation.  
It logs progress, waits for readiness, records the outcome in a check report and updates the global `env` state.

### Signature
```go
func (*checksdb.Check, *provider.TestEnvironment)()
```
- **Parameters**
  - `check *checksdb.Check` – the test‑specific database record that will receive the result of this phase.
  - `env *provider.TestEnvironment` – a shared environment object holding state for the current test run (e.g., Kubernetes client, namespace).

### Workflow
| Step | Action | Details |
|------|--------|---------|
| 1 | Log start | Calls `LogInfo("waiting for operator to be ready")`. |
| 2 | Wait for readiness | Invokes `WaitOperatorReady(env)` which blocks until the Operator’s status is *Ready* or a timeout occurs. |
| 3 | Record success | If the wait succeeds, it appends an **info** field (`AddField("operator_ready", "true")`) to a new report object created by `NewOperatorReportObject(check)`. The report is added to the check’s `Reports` slice via `append`. |
| 4 | Handle failure | If the wait fails, it logs an error with `LogError(err)`, then creates a **warning** field (`AddField("operator_ready", "false")`) and records it similarly. |
| 5 | Finalize | Calls `check.SetResult()` to mark the check as finished (success or failure depending on the earlier steps). |

### Key Dependencies
- **Logging helpers** – `LogInfo`, `LogError` provide structured console output.
- **Operator readiness waiter** – `WaitOperatorReady(env)` encapsulates the logic for polling the Operator’s status.
- **Report construction** – `NewOperatorReportObject(check)`, `AddField(...)`, and slice manipulation (`append`) form the check’s report data structure.
- **Result finalization** – `check.SetResult()` commits the outcome to the checks database.

### Side Effects
- Mutates the passed `*checksdb.Check` by adding a report entry and setting its result status.
- Modifies the global `env` indirectly via `WaitOperatorReady`, which may perform Kubernetes API calls but does not alter global variables.
- Emits logs for visibility in test output.

### Package Context
Within the `operator` package, this function is one of several phase‑specific tests that together validate the Operator lifecycle. It is invoked by higher‑level suite orchestration code (e.g., a Ginkgo `It` block) after the operator deployment has been initiated but before subsequent phases such as configuration or workload verification are run.

---

#### Suggested Mermaid Flowchart

```mermaid
flowchart TD
    A[Start] --> B{WaitOperatorReady}
    B -- success --> C[Create report (operator_ready=true)]
    B -- failure --> D[LogError & Create report (operator_ready=false)]
    C & D --> E[check.SetResult()]
    E --> F[End]
```

This diagram illustrates the decision point on operator readiness and how the outcome propagates to the check record.
