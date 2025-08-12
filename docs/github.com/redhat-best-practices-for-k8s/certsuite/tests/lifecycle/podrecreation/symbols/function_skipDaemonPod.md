skipDaemonPod`

### Purpose
`skipDaemonPod` is a helper predicate used by the pod‑recreation tests to filter out DaemonSet pods from a list of Pods that are being inspected for recreation logic.

- **Why skip?**  
  DaemonSet pods run one copy per node. The test suite focuses on controllers that create *replicated* workloads (Deployments, ReplicaSets, StatefulSets). DaemonSet pods have different lifecycle semantics and should not be subject to the same recreation checks, so they are excluded early.

### Signature
```go
func skipDaemonPod(pod *corev1.Pod) bool
```
- **Input**  
  - `pod`: pointer to a `k8s.io/api/core/v1.Pod` instance that is being considered for processing.
- **Output**  
  - Returns `true` if the Pod should be *skipped* (i.e., it originates from a DaemonSet).  
  - Returns `false` otherwise.

### How It Works
The function inspects the Pod’s controller reference:

```go
if pod.OwnerReferences != nil && len(pod.OwnerReferences) > 0 {
    if pod.OwnerReferences[0].Kind == DaemonSetString { // "DaemonSet"
        return true
    }
}
return false
```

- It checks that `OwnerReferences` is non‑nil and contains at least one entry.  
- If the first owner reference’s kind equals `"DaemonSet"` (the value of the exported constant `DaemonSetString`), it identifies the pod as belonging to a DaemonSet and returns `true`.  
- For any other controller type or absence of an owner, it returns `false`.

### Dependencies
| Dependency | Role |
|------------|------|
| `corev1.Pod` | Kubernetes Pod type from `k8s.io/api/core/v1`. |
| `DaemonSetString` | Exported constant `"DaemonSet"` used to compare controller kind. |

No other global variables or external calls are involved.

### Side Effects
- **None** – the function is pure: it only reads the pod and returns a boolean.

### Package Context
The `podrecreation` package implements tests that verify whether pods created by various controllers are correctly recreated when their owning controller is updated.  
During these tests, a list of Pods may contain DaemonSet pods which should be ignored. `skipDaemonPod` is used in filtering pipelines (e.g., with `filter`, `map`) to exclude such pods before applying recreation logic.

### Suggested Mermaid Diagram
```mermaid
flowchart TD
    A[All Pods] --> B{Is OwnerKind == "DaemonSet"?}
    B -- Yes --> C[Skip Pod]
    B -- No --> D[Process Pod]
```

This function is a small but essential part of ensuring that the recreation tests focus only on replicated workloads.
