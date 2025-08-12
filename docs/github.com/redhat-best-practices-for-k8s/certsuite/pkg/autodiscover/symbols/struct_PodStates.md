PodStates` – Tracking Pod Lifecycle Counts

```go
type PodStates struct {
    AfterExecution  map[string]int // counts per pod name *after* a test run
    BeforeExecution map[string]int // counts per pod name *before* a test run
}
```

| Aspect | Detail |
|--------|--------|
| **Purpose** | Holds the number of Pods that exist for each test case **before** and **after** the execution of an automated test.  The two maps use the same key space – typically the Pod’s name or a test‑specific identifier – and store integer counters.  By comparing `BeforeExecution` to `AfterExecution`, callers can detect pod creation, deletion, or leakage during a test run. |
| **Inputs / Outputs** | - **Input**: The struct itself is populated by code that queries the Kubernetes API before and after executing a suite of tests (e.g., in an *autodiscover* workflow).  No functions are attached to this type, so the caller is responsible for filling the maps. <br> - **Output**: After population, other components can read the two maps to compute differences or generate diagnostics. |
| **Key Dependencies** | - The Kubernetes client-go library (used elsewhere in `autodiscover`) to list Pods.<br>- A test‑execution orchestration component that knows when “before” and “after” snapshots should be taken.  No other internal dependencies exist; the struct is a plain data holder. |
| **Side Effects** | None – `PodStates` has no methods, so it only stores state.  Side effects arise from whatever code writes to its maps (e.g., snapshotting logic). |
| **Package Fit** | The `autodiscover` package automatically discovers and runs tests against a cluster.  During this process it needs to know whether the test created or left behind Pods.  `PodStates` provides the minimal data structure for that bookkeeping, enabling higher‑level logic (e.g., cleaning up leaked resources, reporting metrics) without coupling those concerns directly to the Kubernetes API calls. |

---

## Typical Usage Flow

```mermaid
flowchart TD
    A[Start Test Run] --> B{Collect Before}
    B -->|List Pods| C[Populate PodStates.BeforeExecution]
    C --> D[Execute Tests]
    D --> E{Collect After}
    E -->|List Pods| F[Populate PodStates.AfterExecution]
    F --> G[Compare & Act (e.g., cleanup, report)]
```

* The `autodiscover` runner would instantiate a `PodStates`, call a helper that fills `BeforeExecution`, run the test suite, then fill `AfterExecution`.  
* A downstream component can then inspect `ps.AfterExecution[podName] - ps.BeforeExecution[podName]` to see if any Pods were added or removed.

---

### Summary

- **What**: Simple holder for pre‑ and post‑test pod counts.  
- **Why**: Enables detection of pod churn, leaks, or unintended side effects during automated test runs.  
- **How**: Populated by external code that queries the cluster; read-only thereafter.  
- **Where**: Used throughout `autodiscover` to support cleanup and reporting logic.
