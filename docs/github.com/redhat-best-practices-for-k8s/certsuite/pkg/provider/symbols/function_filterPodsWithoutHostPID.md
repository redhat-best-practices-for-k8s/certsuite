filterPodsWithoutHostPID`

| Aspect | Detail |
|--------|--------|
| **Package** | `provider` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider`) |
| **Visibility** | Unexported (internal helper) |
| **Signature** | `func filterPodsWithoutHostPID(pods []*Pod) []*Pod` |

### Purpose
`filterPodsWithoutHostPID` removes any pod that requests the *host PID namespace* from a slice of pods.  
The function is used by higher‑level filtering logic that only wants to examine pods running in their own PID namespaces, e.g., for tests that validate container isolation.

### Inputs & Outputs
| Parameter | Type | Description |
|-----------|------|-------------|
| `pods` | `[]*Pod` | Slice of pod pointers to be filtered. The slice may contain nil entries; these are preserved. |

| Return | Type | Description |
|--------|------|-------------|
| `[]*Pod` | New slice containing only those pods where `pod.Spec.HostPID == false` (or the field is omitted). Pods with `HostPID: true` are omitted. Nil entries are retained in their original positions relative to the filtered result. |

### Core Logic
```go
func filterPodsWithoutHostPID(pods []*Pod) []*Pod {
    var res []*Pod
    for _, p := range pods {
        if p != nil && !p.Spec.HostPID { // HostPID default is false
            res = append(res, p)
        }
    }
    return res
}
```

* The function iterates over the input slice.  
* For each non‑nil pod it checks `pod.Spec.HostPID`.  
  * If `HostPID` is **false** (the default when not set) or omitted, the pod is kept.  
  * If `HostPID` is **true**, the pod is skipped.  
* The result slice preserves only those pods that satisfy the condition.

### Dependencies
* Relies on the `Pod` type defined elsewhere in the package (`pkg/provider/pods.go`).  
* No external packages or global variables are accessed; the function is pure and side‑effect free.

### Side Effects
None. The function does not modify its input slice or any global state. It simply creates and returns a new slice.

### Context within the Package
`filterPodsWithoutHostPID` is part of the provider’s *pod filtering* utilities (`pkg/provider/filters.go`).  
Other filters (e.g., `filterPodsWithCNI`, `filterPodsWithSecurityContext`) operate on similar slices to narrow down pods that meet specific runtime requirements.  
By removing host‑PID pods early, subsequent analyses or tests can safely assume each remaining pod has its own PID namespace, which is a prerequisite for many isolation checks (e.g., ensuring no cross‑pod process leakage).

### Example Usage
```go
allPods := fetchAllPods()
filtered := filterPodsWithoutHostPID(allPods)
// `filtered` now contains only pods that are not running with HostPID=true
```

This helper keeps the filtering logic concise and reusable across different test scenarios.
