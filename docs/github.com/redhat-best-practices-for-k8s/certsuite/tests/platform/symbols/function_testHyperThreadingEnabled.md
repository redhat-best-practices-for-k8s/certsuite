testHyperThreadingEnabled`

| Aspect | Detail |
|--------|--------|
| **Location** | `suite.go:178` – part of the *platform* test suite |
| **Signature** | `func (*checksdb.Check, *provider.TestEnvironment)` |
| **Exported?** | No (unexported helper used only inside the package) |

---

### Purpose

`testHyperThreadingEnabled` verifies that all bare‑metal nodes in a given test environment have hyper‑threading enabled.  
It populates the supplied `*checksdb.Check` with a report of each node’s status, and marks the overall check as **Failed** if any node is found without hyper‑threading.

---

### Inputs

| Parameter | Type | Role |
|-----------|------|------|
| `check` | `*checksdb.Check` | The check record that will receive the test results. |
| `env`   | `*provider.TestEnvironment` | Provides access to the environment’s node inventory and helper methods (e.g., `GetBaremetalNodes`). |

---

### Workflow

1. **Retrieve Bare‑Metal Nodes**  
   ```go
   nodes := env.GetBaremetalNodes()
   ```
   All nodes that are physically present in the test cluster.

2. **Iterate Over Nodes**  
   For each node:
   * Log the node name.
   * Determine if hyper‑threading is enabled by calling `IsHyperThreadNode(node)`.
   * Create a `NodeReportObject` describing the node’s status (`Passed`, `Failed`, or `Error`).

3. **Handle Outcomes**
   * **Success** – Node has HT enabled → add a passed report.
   * **Failure** – Node lacks HT → add a failed report and flag the overall check as failed.
   * **Error** – Unexpected condition (e.g., empty node list) → add an error report, log the issue, and abort early.

4. **Finalize Check Result**  
   If any node failed, `check.SetResult(checksdb.ResultFailed)` is called to mark the entire test as failed.

---

### Key Dependencies

| Function | Package / Context | Notes |
|----------|------------------|-------|
| `GetBaremetalNodes` | `provider.TestEnvironment` | Retrieves the list of nodes. |
| `LogInfo`, `LogError` | Logging helpers in the same package | Used for diagnostic output. |
| `IsHyperThreadNode` | Utility function in the package | Determines HT status for a node. |
| `NewNodeReportObject` | `checksdb` helper | Creates per‑node report entries. |
| `SetResult` | `*checksdb.Check` | Sets the overall test outcome. |

---

### Side Effects

* **Mutable** – Updates the supplied `check` with node reports and possibly changes its result status.
* **Logging** – Emits informational and error logs to aid debugging.

---

### Package Context

The *platform* package implements a suite of system‑level checks for Kubernetes clusters.  
`testHyperThreadingEnabled` is one such check that runs as part of the broader test harness, ensuring that bare‑metal nodes meet performance requirements (hyper‑threading). It relies on the environment abstraction (`provider.TestEnvironment`) to query node information and feeds results into the `checksdb` reporting system.

---

### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[Start] --> B{GetBareMetalNodes}
    B --> C{Iterate Nodes}
    C --> D[Log Node]
    D --> E{IsHyperThreadNode?}
    E -- Yes --> F[Create Passed Report]
    E -- No  --> G[Create Failed Report & flag fail]
    E -- Error --> H[Create Error Report, Log, Abort]
    F --> I[Continue Iteration]
    G --> I
    H --> Z[End with Error]
    I --> J{All Nodes?}
    J -- Yes --> K[SetResult(Failed?) if any failed]
    K --> L[End]
```

This visual captures the decision flow and side‑effects of `testHyperThreadingEnabled`.
