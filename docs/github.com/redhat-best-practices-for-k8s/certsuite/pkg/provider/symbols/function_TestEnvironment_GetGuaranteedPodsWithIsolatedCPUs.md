TestEnvironment.GetGuaranteedPodsWithIsolatedCPUs`

| Aspect | Detail |
|--------|--------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider` |
| **Receiver type** | `*TestEnvironment` – a struct that holds the state of the test environment (pods, nodes, etc.). |
| **Signature** | `func (env *TestEnvironment) GetGuaranteedPodsWithIsolatedCPUs() []*Pod` |

### Purpose

The method returns all pods that satisfy two conditions:

1. **CPU isolation compliance** – the pod has CPU requests/limits that are exclusive to a set of CPUs and no other pod can share those CPUs.
2. **Guarantee** – the pod is *guaranteed* to run on those isolated CPUs, meaning it uses a `Guaranteed` QoS class and requests equal or more CPUs than its limits.

In short: *“Give me every pod that really owns an isolated CPU set.”*

### Inputs / State

The function relies solely on the current state of the `TestEnvironment`.  
It iterates over `env.Pods`, which is a slice of pointers to `Pod` objects gathered during test discovery. No external arguments are required.

### Output

A slice of pointers to `Pod` (`[]*Pod`).  
Each element in the returned slice:

- Passes both `IsPodGuaranteedWithExclusiveCPUs` and `IsCPUIsolationCompliant`.
- Therefore it can be safely used by tests that need to exercise CPU‑isolated workloads (e.g., memory‑heavy containers, hyper‑thread detection).

### Key Dependencies

| Dependency | Role |
|------------|------|
| `IsPodGuaranteedWithExclusiveCPUs(pod *Pod) bool` | Checks QoS class and CPU request/limit equality. |
| `IsCPUIsolationCompliant(pod *Pod) bool` | Validates that the pod’s CPU annotations or resource limits match an isolated CPU set. |
| Go builtin `append` | Builds the result slice incrementally. |

These helper functions are defined elsewhere in the same package and encapsulate the logic for the two checks.

### Side Effects

None – the function is pure; it only reads from `env.Pods`.  
It does not modify any state, nor does it trigger API calls or network traffic.

### How It Fits Into the Package

- **Filtering Layer**: `GetGuaranteedPodsWithIsolatedCPUs` sits on top of lower‑level pod filtering helpers (`IsPodGuaranteedWithExclusiveCPUs`, `IsCPUIsolationCompliant`).  
- **Test Setup**: Tests that need to target CPU‑isolated workloads can call this method to obtain the relevant pods and then apply further assertions or actions.  
- **Extensibility**: Adding new isolation checks only requires updating the helper functions; the filtering logic here remains unchanged.

---

#### Suggested Mermaid Diagram

```mermaid
flowchart TD
    subgraph TestEnvironment
        env[TestEnvironment]
        pods[env.Pods]
    end
    subgraph Helpers
        h1[IsPodGuaranteedWithExclusiveCPUs(pod)]
        h2[IsCPUIsolationCompliant(pod)]
    end
    subgraph Output
        result[[]*Pod]
    end

    env --> pods
    pods --> podFilter{for each pod}
    podFilter --> h1
    h1 -- true --> h2
    h2 -- true --> result
```

This diagram illustrates how the function iterates over all pods, applies two predicates, and collects those that satisfy both.
