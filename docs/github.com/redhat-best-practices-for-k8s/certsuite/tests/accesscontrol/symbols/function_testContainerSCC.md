testContainerSCC`

| Item | Description |
|------|-------------|
| **Purpose** | Determines whether containers running in the cluster comply with the *least‑privileged* Security Context Constraint (SCC) policy. The function inspects every container referenced by a test check, classifies them into compliance buckets, and records a pass/fail result on the supplied `Check`. |
| **Signature** | `func (*checksdb.Check, *provider.TestEnvironment)` |

---

### Inputs

| Parameter | Type | Role |
|-----------|------|------|
| `check` | `*checksdb.Check` | The compliance check that will receive the final verdict (`Pass`, `Fail`, or `Unknown`). It also holds a container list that this test will iterate over. |
| `env` | `*provider.TestEnvironment` | Holds context for the current test run: cluster client, logger, and helper functions such as `CheckPod`. |

---

### Workflow

1. **Logging** – The function starts with an informational log entry (`LogInfo`) to indicate the start of SCC evaluation.

2. **Container Iteration**  
   For each container referenced by the check:
   - Call `CheckPod` to retrieve pod information (namespace, name, labels, etc.).
   - If a pod cannot be found or any error occurs, log an error (`LogError`) and skip that container.

3. **SCC Classification**  
   Each successful pod is wrapped into a *container report object* (`NewContainerReportObject`).  
   The function uses the SCC name associated with the pod (obtained via `String()`) to decide whether the container is in the least‑privileged bucket or a higher‑privilege one.

4. **Result Aggregation**  
   - Containers that meet the least‑privileged criteria are appended to a *compliant* slice.
   - Others go into a *non‑compliant* slice.
   Both slices are logged and stored in the check’s result object (`NewReportObject`).

5. **Final Verdict**  
   If any container is non‑compliant, `SetResult(check, checksdb.Failed)` is called; otherwise, the test passes.

---

### Key Dependencies

| Dependency | Role |
|------------|------|
| `LogInfo`, `LogError` | Structured logging for progress and error reporting. |
| `CheckPod` | Retrieves pod details needed to determine SCC assignment. |
| `NewContainerReportObject`, `NewReportObject` | Builds structured report objects that are attached to the check result. |
| `SetResult` | Finalizes the compliance status on the `Check`. |

---

### Side Effects

- Adds detailed fields (`AddField`) and types (`SetType`) to the check’s report object.
- Produces logs for every container processed.
- Does **not** modify cluster state; all interactions are read‑only.

---

### Package Context

`testContainerSCC` lives in the `accesscontrol` test suite.  
Its job is part of a larger validation workflow that ensures workloads run with minimal privileges, complementing other tests (e.g., namespace restrictions, service account checks). The function operates only on data supplied by the check and environment objects, keeping it isolated from global state.

---

### Suggested Mermaid Diagram

```mermaid
flowchart TD
  A[Start] --> B{Iterate Containers}
  B -->|Found Pod| C[Wrap in ReportObject]
  C --> D{SCC compliant?}
  D -- Yes --> E[Add to Compliant List]
  D -- No --> F[Add to Non‑Compliant List]
  B -->|Error| G[Log Error & Skip]
  E & F --> H[Aggregate Results]
  H --> I{Any non‑compliant?}
  I -- Yes --> J[SetResult(Failed)]
  I -- No --> K[SetResult(Passed)]
  J & K --> L[End]
```

This diagram visualizes the decision flow and highlights how compliant/non‑compliant containers are handled.
