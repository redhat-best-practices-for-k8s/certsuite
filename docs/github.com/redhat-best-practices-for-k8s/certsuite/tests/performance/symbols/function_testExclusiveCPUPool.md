testExclusiveCPUPool`

| Item | Detail |
|------|--------|
| **Signature** | `func(*checksdb.Check, *provider.TestEnvironment)` |
| **Visibility** | Unexported (package‑private) – used only within the performance test suite. |

### Purpose
`testExclusiveCPUPool` verifies that a pod’s containers are correctly scheduled onto an **exclusive CPU pool** (i.e., the CPUs allocated to the pod are not shared with other workloads). It runs as part of the performance test harness and reports its outcome in a `PodReportObject`.

The function:

1. Determines whether exclusive CPUs have been assigned (`HasExclusiveCPUsAssigned`).
2. If they exist, it creates two `PodReportObject`s – one for the **CPU‑exclusive** pod and another for a **non‑exclusive** container (if present) – to capture metrics like CPU usage or latency.
3. Sets the test result status based on whether exclusive CPUs were found.

### Parameters
| Parameter | Type | Role |
|-----------|------|------|
| `check` | `*checksdb.Check` | The test definition; used only for logging and reporting purposes. |
| `env`   | `*provider.TestEnvironment` | Provides context such as the logger, pod information, and helper functions (e.g., `HasExclusiveCPUsAssigned`). |

### Key Dependencies
| Dependency | What it does |
|------------|--------------|
| `HasExclusiveCPUsAssigned(env)` | Returns a boolean indicating whether the current pod has exclusive CPU allocation. |
| `GetLogger()` | Retrieves a structured logger for recording test progress and errors. |
| `NewPodReportObject(...)` | Constructs a report entry that will be attached to the final test result. |
| `SetResult(result)` | Marks the test outcome (`pass`, `fail`, or `skip`). |

### Workflow (high‑level)
```mermaid
flowchart TD
    A[Start] --> B{HasExclusiveCPUsAssigned?}
    B -- No --> C[Skip test]
    B -- Yes --> D[Create CPU‑exclusive PodReportObject]
    D --> E[Create non‑exclusive PodReportObject (if applicable)]
    E --> F[SetResult(pass)]
    C --> G[SetResult(skip)]
```

1. **Check CPU exclusivity**  
   `HasExclusiveCPUsAssigned` is called; if it returns `false`, the test is marked as skipped and exits early.

2. **Logging**  
   The function logs key details (e.g., pod name, namespace) using the package logger.

3. **Report objects**  
   - A report for the exclusive‑CPU pod is created with fields like `PodName`, `Namespace`, `ContainerName`, and `IsExclusive=true`.  
   - If a non‑exclusive container exists in the same pod, another report object is built similarly but with `IsExclusive=false`.

4. **Result**  
   On success (i.e., exclusive CPUs found), the test result is set to *pass*; otherwise it remains *skip*.

### Side Effects
- No state changes are made outside of logging and report generation.
- The function may log an error if creating a `PodReportObject` fails, but this does not alter the test outcome.

### Placement in Package
Within `github.com/redhat-best-practices-for-k8s/certsuite/tests/performance`, this helper is used by the broader test runner to validate CPU isolation policies. It complements other functions that skip tests when prerequisites (e.g., guaranteed containers, host‑PID restrictions) are not met. The function’s name and signature suggest it is invoked via a closure in a `beforeEachFn` or similar hook during test execution.

---

**TL;DR:**  
`testExclusiveCPUPool` checks if a pod has exclusive CPU allocation, logs the status, creates detailed report objects for both exclusive and non‑exclusive containers, and marks the test as passed or skipped accordingly.
