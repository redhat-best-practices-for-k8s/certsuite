## `GetNonGuaranteedPodContainersWithoutHostPID`

| Aspect | Details |
|--------|---------|
| **Location** | `pkg/provider/filters.go:207` |
| **Receiver** | `TestEnvironment` (value receiver) |
| **Signature** | `func() []*Container` |
| **Exported?** | Yes |

### Purpose
Collects the containers that belong to *non‑guaranteed* pods and whose pods do **not** run with the `HostPID` feature enabled.  
In OpenShift/Kubernetes, a pod is *non‑guaranteed* when its QoS class is not `Guaranteed`. The function returns only those containers from such pods that are running in a namespace where the pod’s spec has `hostPID: false` (or unset).

This list is typically used by tests that validate security constraints or resource isolation for containers that share the host PID namespace.

### Inputs
* **Receiver (`TestEnvironment`)** – The test environment holds all information about nodes, pods, and containers in the cluster under test.  
  - No other parameters are required; the function derives everything from the receiver’s internal state.

### Outputs
* `[]*Container` – A slice of pointers to `Container` objects that satisfy both criteria:
  1. Belong to a non‑guaranteed pod.
  2. The pod does **not** have `HostPID: true`.

The order of the slice is deterministic based on the underlying ordering returned by the helper functions.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `GetNonGuaranteedPods()` | Returns all pods that are not guaranteed (QoS != `Guaranteed`). |
| `filterPodsWithoutHostPID(pods []*Pod) []*Pod` | Filters a pod list, keeping only those whose `spec.hostPID` is false or unset. |
| `getContainers(pods []*Pod) []*Container` | Expands each pod into its constituent containers. |

The function chains these helpers in the following way:

```go
pods := GetNonGuaranteedPods()
filteredPods := filterPodsWithoutHostPID(pods)
return getContainers(filteredPods)
```

### Side Effects
* **None** – The function is read‑only; it does not modify the environment or any global state.
* It relies solely on the current snapshot of the `TestEnvironment`; no external calls are made.

### How It Fits the Package
`provider` contains a collection of filters that extract subsets of cluster objects for testing purposes.  
- `GetNonGuaranteedPodContainersWithoutHostPID` is part of the *container filtering* set, complementing functions like `GetAllContainers()` or `GetContainersWithSpecificLabels()`.  
- It allows test suites to focus on containers that are potentially vulnerable because they run in pods without host PID isolation and are not guaranteed QoS.  

By exposing this filter as a method on `TestEnvironment`, callers can easily integrate it into complex queries, e.g.:

```go
env := provider.NewTestEnvironment(...)
badContainers := env.GetNonGuaranteedPodContainersWithoutHostPID()
for _, c := range badContainers {
    // run security checks...
}
```

---

#### Mermaid diagram (suggestion)

```mermaid
graph TD
  A[GetNonGuaranteedPods] --> B[filterPodsWithoutHostPID]
  B --> C[getContainers]
  C --> D[Result: []*Container]
```
This visualizes the linear flow of data through the helper functions.
