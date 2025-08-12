testPodNodeSelectorAndAffinityBestPractices`

| Aspect | Details |
|--------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle` |
| **Visibility** | unexported (private to the package) |
| **Signature** | `func([]*provider.Pod, *checksdb.Check)()` |

### Purpose
Runs a check that verifies each pod in a given set follows Kubernetes best‑practice rules for **node selector** and **affinity/anti‑affinity** usage.  
The function is invoked by the lifecycle test suite when evaluating the state of all pods returned by the provider.

### Parameters

| Name | Type | Meaning |
|------|------|---------|
| `pods` | `[]*provider.Pod` | Slice of pod objects that were retrieved from the cluster (or a mock). |
| `check` | `*checksdb.Check` | Database record describing the check to be performed. The function updates this object with results and logs. |

### Workflow

1. **Logging start**  
   Calls `LogInfo("Checking node selector and affinity best practices")`.

2. **Iterate over each pod**  
   For every pod in `pods`:

   - If the pod *does not* have a node selector (`HasNodeSelector(pod)` returns false), the function records an error:
     ```go
     check.AddReport(NewPodReportObject(pod, "Missing Node Selector", true))
     ```
     It also logs the error via `LogError(...)`.

   - If the pod *does* have a node selector but has **no affinity** or **anti‑affinity** rules (the code checks for those properties directly), an informational report is added:
     ```go
     check.AddReport(NewPodReportObject(pod, "No Affinity or Anti-Affinity", false))
     LogInfo(...)
     ```

   - If the pod satisfies both node selector *and* affinity/anti‑affinity constraints, a positive report is created:
     ```go
     check.AddReport(NewPodReportObject(pod, "Node Selector and Affinity present", true))
     ```

3. **Finalize result**  
   After all pods have been processed, the overall result for this check is set to `true` (pass) with `SetResult(true)`.

### Key Dependencies

| Dependency | Role |
|------------|------|
| `LogInfo`, `LogError` | Emit logs that appear in test output and trace files. |
| `HasNodeSelector(pod)` | Helper that inspects a pod’s spec for a node selector map. |
| `NewPodReportObject(...)` | Constructs a structured report entry for the check database. |
| `SetResult` | Marks the check as passed or failed in the `checksdb.Check` object. |

### Side Effects

* Mutates the supplied `check` by adding pod‑level reports and setting the final result.
* Emits logs via the package’s logging helpers; these are visible to anyone running the test suite.

### Package Integration

This function is one of many “test” helpers used in the **lifecycle** test suite.  
During a lifecycle run, the suite collects all pods that belong to the workload under test and passes them to this helper (along with the corresponding `checksdb.Check`). The resulting reports are stored in the database for later analysis or display.

The function is intentionally small and focused: it does not interact with any global state, relies only on the supplied arguments, and returns immediately. This design keeps tests deterministic and easy to reason about.
