## `GetNonGuaranteedPods`

| Aspect | Detail |
|--------|--------|
| **Package** | `provider` (github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider) |
| **Receiver** | `TestEnvironment` – the current test environment that holds all inspected Kubernetes objects. |
| **Signature** | `func (te TestEnvironment) GetNonGuaranteedPods() []*Pod` |

### Purpose
Return a slice containing *only* those pods in the test environment whose scheduling guarantees are **not** met.  
In Kubernetes, a pod is considered *guaranteed* when every container has explicit resource requests and limits that match or exceed each other. This helper filters out such guaranteed pods so that downstream tests can focus on non‑guaranteed workloads (e.g., to validate proper resource allocation, eviction policies, etc.).

### Inputs
The function receives no arguments; it operates solely on the `TestEnvironment` instance (`te`).  
Internally it accesses:

* `te.Pods` – the full list of `Pod` objects that were loaded into the environment.

### Outputs
A slice of pointers to `Pod` objects:
```go
[]*Pod
```
Each element satisfies `!IsPodGuaranteed(pod)`.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `te.Pods` | The data source; all known pods are iterated over. |
| `IsPodGuaranteed(*Pod) bool` | Determines whether a pod is guaranteed or not. It is called for each pod to decide inclusion in the result slice. |

### Algorithm (simplified)
```go
func (te TestEnvironment) GetNonGuaranteedPods() []*Pod {
    var res []*Pod
    for _, p := range te.Pods {
        if !IsPodGuaranteed(p) { // only non‑guaranteed pods
            res = append(res, p)
        }
    }
    return res
}
```

1. Initialise an empty slice `res`.
2. Iterate over every pod in the environment.
3. If `IsPodGuaranteed` returns `false`, append that pod to `res`.
4. Return `res`.

### Side‑Effects & Mutability
* The function **does not modify** any state of `TestEnvironment`.  
* It only reads from `te.Pods`; no global variables are accessed or mutated.

### Where It Fits in the Package
`GetNonGuaranteedPods` is a *filter* helper located in `filters.go`.  
Other package components (e.g., test suites, compliance checks) call this method when they need to focus on non‑guaranteed workloads. By centralising the filtering logic here, the codebase avoids duplication and ensures consistent semantics across all tests.

### Suggested Mermaid Diagram
A small flowchart can help visualise the relationship between `TestEnvironment`, pods, and the guarantee check:

```mermaid
flowchart TD
    TE[TestEnvironment]
    Pods[te.Pods]
    Pod[*Pod]
    Check[IsPodGuaranteed(pod)]
    Result[Non‑guaranteed pods]

    TE -->|contains| Pods
    Pods -->|iterates| Pod
    Pod -->|checked by| Check
    Check -- false --> Result
```

This diagram illustrates that the method simply traverses the pod list and collects those failing the guarantee check.
