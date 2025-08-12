TestEnvironment.GetGuaranteedPodsWithExclusiveCPUs`

### Purpose
`GetGuaranteedPodsWithExclusiveCPUs` is a helper that extracts the subset of Pods in the current test environment which **are guaranteed** to run on *exclusive* CPU resources.  
This is useful for tests that need to work only with workloads that have deterministic CPU placement (e.g., NUMA‑aware or security isolation tests).

### Receiver
```go
func (env TestEnvironment) GetGuaranteedPodsWithExclusiveCPUs() []*Pod
```
The method operates on a `TestEnvironment` value, which holds the full list of Pods discovered in the cluster.

### Inputs / Outputs
| Direction | Type            | Description |
|-----------|-----------------|-------------|
| Input     | None (receiver only) | The current state of the test environment. |
| Output    | `[]*Pod`        | A slice containing pointers to all Pods that satisfy the exclusive‑CPU guarantee criterion. |

### Implementation Highlights
1. **Iteration** – Loops over `env.Pods`, which is a collection of all discovered Pod objects.
2. **Filtering** – For each pod, it calls the function `IsPodGuaranteedWithExclusiveCPUs(pod)`.  
   - This helper inspects the pod’s resource requests/limits and CPU topology hints to determine if the pod will be scheduled on dedicated CPUs.
3. **Accumulation** – Pods that return `true` are appended to a result slice using Go’s built‑in `append`.
4. **Return** – The filtered slice is returned.

```go
func (env TestEnvironment) GetGuaranteedPodsWithExclusiveCPUs() []*Pod {
    var guaranteed []*Pod
    for _, p := range env.Pods {
        if IsPodGuaranteedWithExclusiveCPUs(p) {
            guaranteed = append(guaranteed, p)
        }
    }
    return guaranteed
}
```

### Dependencies
| Dependency | Role |
|------------|------|
| `IsPodGuaranteedWithExclusiveCPUs` | Determines exclusivity of CPU allocation for a pod. |
| `append` (built‑in) | Builds the result slice. |

No external packages are imported directly; all logic relies on other helpers within the same package.

### Side Effects
None – the function is read‑only and does not modify the environment or any global state.

### Context in the Package
- **Provider Role** – The `provider` package models Kubernetes objects (Pods, Nodes, etc.) for certsuite’s testing framework.  
- **Filter Utilities** – This method lives alongside other “filter” helpers that help tests narrow down relevant resources (e.g., `GetWorkerNodes`, `GetMasterNodes`).  
- **Test Usage** – Test cases that need to validate CPU isolation or NUMA placement typically call this function to obtain the pods they should inspect.

### Mermaid Diagram (Optional)

```mermaid
graph TD;
    A[TestEnvironment] --> B[Pods List];
    B --> C{For each Pod}
    C -->|IsPodGuaranteedWithExclusiveCPUs(pod)| D[Include in Result]
    C -->|else| E[Skip];
    D --> F[Append to slice];
```

This diagram visualizes the filtering loop: iterate over all pods, test each one, and build a new slice of those that qualify.
