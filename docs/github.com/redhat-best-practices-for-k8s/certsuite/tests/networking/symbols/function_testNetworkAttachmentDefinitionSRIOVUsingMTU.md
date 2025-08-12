testNetworkAttachmentDefinitionSRIOVUsingMTU`

| Item | Details |
|------|---------|
| **Package** | `networking` (`github.com/redhat-best-practices-for-k8s/certsuite/tests/networking`) |
| **File** | `suite.go:427` |
| **Exported?** | No – used internally in the test suite. |
| **Signature** | `func(*checksdb.Check, []*provider.Pod)()` |

### Purpose
Runs a network‑attachment‑definition (NAD) check that verifies whether the pods in the provided list are using SR‑IOV networking with an MTU value equal to the cluster default (`defaultMTU`).  
The function builds a series of `PodReportObject`s, collects their results, and stores them back into the supplied `*checksdb.Check` instance.

### Inputs
| Parameter | Type | Description |
|-----------|------|-------------|
| `c` | `*checksdb.Check` | The check record that will receive the test result. |
| `pods` | `[]*provider.Pod` | Pods to evaluate for SR‑IOV/MTU usage. |

### Workflow
1. **Iterate over each pod**  
   For every pod in `pods`, call `IsUsingSRIOVWithMTU(pod)` to determine if the pod’s network interfaces match the SR‑IOV + MTU criteria.
2. **Build a report object**  
   * If the pod passes, create a success `PodReportObject` and append it to `c.Report.PodReports`.  
   * If the pod fails or an error occurs, log the error (`LogError`) and create a failure `PodReportObject`, also appending it.
3. **Aggregate results**  
   After all pods are processed, set the overall result of the check via `SetResult(c)`.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `IsUsingSRIOVWithMTU` | Determines SR‑IOV/MTU compliance for a single pod. |
| `LogError`, `LogInfo` | Emit test‑suite logs. |
| `NewPodReportObject` | Constructs per‑pod result objects. |
| `append` | Adds report objects to the check’s report slice. |
| `SetResult` | Finalizes and records the overall test outcome in the database. |

### Side Effects
* Logs diagnostic messages for each pod.
* Mutates the supplied `checksdb.Check` by adding `PodReportObject`s and setting its final result.

### Context within the Package
The `networking` package hosts end‑to‑end tests that validate various networking configurations on a Kubernetes cluster.  
`testNetworkAttachmentDefinitionSRIOVUsingMTU` is one of several helper functions used in the suite’s test cases to verify SR‑IOV support and MTU alignment. It operates on data structures (`checksdb.Check`, `provider.Pod`) defined elsewhere in the tests but plays a crucial role in producing structured, machine‑readable results for downstream reporting or alerting.

### Suggested Mermaid Flow
```mermaid
flowchart TD
    A[Start] --> B{For each pod}
    B -->|Passes SR‑IOV+MTU?| C[Create success PodReportObject]
    B -->|Fails?| D[LogError & create failure PodReportObject]
    C --> E[Append to c.Report.PodReports]
    D --> E
    E --> F{All pods processed?}
    F --> G[SetResult(c)]
    G --> H[End]
```

---
