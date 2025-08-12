testSchedulingPolicyInCPUPool`

| Aspect | Detail |
|--------|--------|
| **Location** | `suite.go:244` in the *performance* test package |
| **Exported?** | No – it is a private helper used only inside this file. |
| **Signature** | ```go
func (*checksdb.Check, *provider.TestEnvironment,
        []*provider.Container, string) ()```
> It receives a check record, the test environment, a slice of containers that belong to the same CPU pool, and the name of the scheduling policy being tested. It returns a `func()` which is the actual test body executed by the testing framework. |

### Purpose

The function verifies that all processes running inside a set of containers (belonging to one CPU pool) are scheduled according to the expected *CPU scheduling policy* (e.g., `CFS`, `RT`, or `Idle`).  
It does so by:

1. **Identifying PIDs** – for each container it fetches its PID namespace, then lists all processes inside that namespace.
2. **Collecting CPU‑policy data** – calls `ProcessPidsCPUScheduling` on the gathered PIDs to obtain their scheduling class.
3. **Aggregating results** – builds a report per container and a global summary of how many processes use each policy.
4. **Reporting outcome** – sets the check result (pass/fail) based on whether any process is found with an unexpected policy.

The returned closure is executed by the test harness; it keeps the outer function lightweight while letting the test framework handle setup/teardown and error handling.

### Inputs

| Parameter | Type | Role |
|-----------|------|------|
| `check` | `*checksdb.Check` | Holds metadata for this particular test run (e.g., ID, name). The function writes results into it via `SetResult`. |
| `env` | `*provider.TestEnvironment` | Provides access to the Kubernetes environment, including logger and helper functions. |
| `containers` | `[]*provider.Container` | All containers that share a CPU pool; each container is inspected for its processes. |
| `policyName` | `string` | The human‑readable name of the scheduling policy being verified (used in log messages). |

### Key Dependencies

- **Logging** – `LogInfo`, `LogDebug`, and `LogError` output progress and error details through `env.GetLogger()`.
- **Namespace utilities** – `GetContainerPidNamespace` obtains a container’s PID namespace; `GetPidsFromPidNamespace` enumerates processes within that namespace.
- **Scheduling analysis** – `ProcessPidsCPUScheduling` returns a map of PIDs to their scheduling class, which the test uses to validate against `policyName`.
- **Result handling** – `SetResult` records whether the test passed or failed.

### Side Effects

1. **Console output** – Detailed logs for each container and process are emitted.
2. **Check record mutation** – The function updates the supplied `check` with a result (`PASS`, `FAIL`) and a descriptive message if any mis‑scheduled process is found.
3. **No external state changes** – It does not modify containers, namespaces, or the environment.

### How it fits the package

Within the *performance* test suite, several tests create CPU pools (e.g., exclusive CPUs, isolated CPUs). Each of those tests calls `testSchedulingPolicyInCPUPool` with an appropriate set of containers and a target policy. This helper centralizes the common logic for inspecting process scheduling, ensuring consistent reporting and reducing duplication across different performance checks.

---

#### Suggested Mermaid diagram

```mermaid
flowchart TD
    A[Containers] -->|PID namespace| B[GetContainerPidNamespace]
    B --> C[GetPidsFromPidNamespace]
    C --> D[ProcessPidsCPUScheduling]
    D --> E{Policy matches}
    E -- yes --> F[Mark container OK]
    E -- no  --> G[Record error]
    G --> H[SetResult(FAIL)]
```

This diagram illustrates the flow from containers to final result determination.
