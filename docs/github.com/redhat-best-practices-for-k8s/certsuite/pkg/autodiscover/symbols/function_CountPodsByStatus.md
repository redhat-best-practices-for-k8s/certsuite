CountPodsByStatus`

| Item | Detail |
|------|--------|
| **Package** | `autodiscover` (github.com/redhat-best-practices-for-k8s/certsuite/pkg/autodiscover) |
| **Signature** | `func CountPodsByStatus(pods []corev1.Pod) map[string]int` |
| **Exported?** | Yes |

### Purpose
`CountPodsByStatus` aggregates the number of Kubernetes Pods per lifecycle status.  
It is used by other discovery helpers to determine, for example, how many Pods are *Running*, *Pending*, *Failed*, etc., within a set of resources returned from a cluster query.

### Inputs

| Parameter | Type | Notes |
|-----------|------|-------|
| `pods` | `[]corev1.Pod` | A slice of Pod objects (from the client-go API). The function does **not** modify this slice. |

### Output
A map keyed by pod status string (`pod.Status.Phase`) with integer counts.

```go
map[string]int{
    "Running":   12,
    "Pending":   3,
    "Failed":    1,
}
```

If the input slice is empty, an empty map is returned.

### Key Dependencies

* **`corev1.Pod`** – The function only reads `pod.Status.Phase`.  
* No external packages or global variables are accessed; it is a pure helper.

### Side‑Effects
None. The function performs no I/O, does not mutate the input slice, and has no effect on global state.

### How It Fits in the Package

`autodiscover` contains utilities that inspect cluster objects to decide which certificates need renewal or installation.  
- `CountPodsByStatus` is typically called after a call to a client-go list operation (e.g., `client.CoreV1().Pods(namespace).List(...)`).  
- The resulting map informs other functions like `CheckPodHealth()` or `DetectStuckDeployments()`, which look for abnormal pod counts.  

The function is deliberately small and side‑effect free so it can be reused across multiple discovery scenarios without risk of unintended state changes.

### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[GetPods] --> B{CountPodsByStatus}
    B --> C[Return map[string]int]
```

This visualises the flow: a list of Pods is fed into `CountPodsByStatus`, which outputs the aggregated counts.
