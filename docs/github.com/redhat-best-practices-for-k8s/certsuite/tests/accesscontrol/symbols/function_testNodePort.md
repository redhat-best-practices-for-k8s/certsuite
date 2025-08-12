testNodePort` – Node‑Port compliance checker

### Purpose
`testNodePort` inspects all Kubernetes services in the test environment to verify that each service of a particular type is exposed as a **NodePort** (i.e., its `spec.type == "NodePort"`).  
The function produces a compliance report that lists:

* compliant services – those correctly configured with NodePort
* non‑compliant services – those not using NodePort

It then sets the result of the corresponding check (`Check`) to **PASS** if every service is compliant, otherwise **FAIL**.

### Signature
```go
func testNodePort(check *checksdb.Check, env *provider.TestEnvironment)
```

| Parameter | Type                           | Description |
|-----------|--------------------------------|-------------|
| `check`   | `*checksdb.Check`              | The check instance that will receive the result and report. |
| `env`     | `*provider.TestEnvironment`    | Test environment containing all Kubernetes objects to analyse (services, nodes, etc.). |

### Key dependencies
| Dependency | Role |
|------------|------|
| `LogInfo`, `LogError` | Logging helpers that write diagnostic information to the test log. |
| `ToString` | Utility for converting complex values into human‑readable strings. |
| `AddField` | Method of a report object used to add key/value pairs describing each service. |
| `NewReportObject` | Factory that creates a new, empty report entry for a service. |
| `SetResult` | Marks the check as PASS/FAIL and attaches the aggregated report. |

### How it works
1. **Iterate over services** – The function loops through all services in `env`.  
2. **Classification** – For each service it checks the `spec.type` field:
   * If it is `"NodePort"`, a “compliant” report object is created and appended to a slice of compliant objects.
   * Otherwise, a “non‑compliant” report object is created and appended to another slice.  
3. **Logging** – Each service’s status (compliant / non‑compliant) is logged using `LogInfo` or `LogError`.  
4. **Result aggregation** – After processing all services:
   * The function calls `SetResult` on the provided check.
   * If any non‑compliant objects exist, the result is set to `FAIL`; otherwise `PASS`.
   * All report objects are attached to the check for later inspection or export.

### Side effects
* No mutation of the test environment – it only reads service data.  
* Generates log output and populates the check’s internal state with a detailed report.

### Package context
`testNodePort` lives in the `accesscontrol` test suite, which validates Kubernetes security best practices.  
It is invoked as part of a larger test harness that runs multiple checks against a live cluster or a mock environment.  
The function exemplifies how individual compliance checks collect data, build structured reports, and report pass/fail status back to the orchestrator.

### Suggested Mermaid diagram (optional)

```mermaid
flowchart TD
    A[Start] --> B{For each Service}
    B --> |NodePort| C[Add to compliant list]
    B --> |Other| D[Add to non‑compliant list]
    C --> E[LogInfo]
    D --> F[LogError]
    E & F --> G[End loop]
    G --> H{Any non‑compliant?}
    H -- Yes --> I[SetResult(FAIL)]
    H -- No  --> J[SetResult(PASS)]
    I & J --> K[Attach report to check]
```

This diagram visualises the decision flow for each service and how the final result is determined.
