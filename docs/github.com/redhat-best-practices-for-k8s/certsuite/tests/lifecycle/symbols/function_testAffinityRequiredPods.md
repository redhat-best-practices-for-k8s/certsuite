testAffinityRequiredPods`

| Item | Details |
|------|---------|
| **Package** | `lifecycle` (`github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle`) |
| **Signature** | `func testAffinityRequiredPods(check *checksdb.Check, env *provider.TestEnvironment)` |
| **Exported?** | No – used only inside the test suite. |

### Purpose
This helper is part of a larger end‑to‑end lifecycle testing framework for Kubernetes workloads.  
It verifies that every pod in the current `TestEnvironment` satisfies the affinity rules defined by the workload’s `AffinityRequiredPods` list.

The function performs the following steps:

1. **Collect required pods** – Calls `GetAffinityRequiredPods(check, env)` to obtain a slice of `PodReportObject`s that should be present according to the test's affinity configuration.
2. **Log the intent** – Emits an informational log describing how many pods are expected.
3. **Check compliance** – Uses `IsAffinityCompliant(pods, check)` to confirm that all required pods exist and match the defined criteria.
4. **Record results** –  
   * If the check passes, a “pass” result is appended to the report.  
   * If it fails, an error message is logged, the failure is recorded in the test result set via `SetResult`, and detailed pod objects are added to the report for debugging.

### Inputs

| Parameter | Type | Role |
|-----------|------|------|
| `check` | `*checksdb.Check` | Contains metadata for the current test case (e.g., name, description). |
| `env` | `*provider.TestEnvironment` | Represents the runtime environment: list of pods, nodes, and other Kubernetes resources relevant to the test. |

### Outputs

The function **does not return** a value; its side‑effects are:

- **Logging** – Uses `LogInfo` and `LogError` for diagnostics.
- **Report mutation** – Appends `PodReportObject`s via `NewPodReportObject` to the check’s report data structure.
- **Result state** – Calls `SetResult(check, true/false)` to mark the test as passed or failed.

### Key Dependencies

| Dependency | What it does |
|------------|--------------|
| `GetAffinityRequiredPods` | Pulls the list of pods that must exist for affinity compliance. |
| `IsAffinityCompliant` | Validates actual pod presence against required list. |
| `NewPodReportObject` | Creates a structured report entry for each relevant pod. |
| `SetResult` | Persists the pass/fail status of the test case. |

### Interaction with the Package

* **Test Lifecycle** – The function is invoked from a Ginkgo/Go test (likely within a `Describe` block). It runs after the environment has been provisioned but before any cleanup.
* **Reporting** – Its output feeds into the overall test report that is later serialized to JSON or displayed on the console.

### Mermaid Diagram (suggestion)

```mermaid
flowchart TD
    A[Start] --> B{GetRequiredPods}
    B -->|Success| C[LogInfo]
    C --> D{IsCompliant?}
    D -->|Yes| E[SetResult Pass]
    D -->|No| F[LogError & SetResult Fail]
    E --> G[End]
    F --> H[Append PodReportObject(s)]
    H --> G
```

This diagram visualises the decision flow inside `testAffinityRequiredPods`.
