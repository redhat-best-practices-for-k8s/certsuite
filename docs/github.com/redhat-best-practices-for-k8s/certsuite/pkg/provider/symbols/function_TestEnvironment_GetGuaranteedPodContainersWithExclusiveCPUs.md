TestEnvironment.GetGuaranteedPodContainersWithExclusiveCPUs`

### Purpose
`GetGuaranteedPodContainersWithExclusiveCPUs` extracts all **containers** that run in a *guaranteed* pod and are scheduled with **exclusive CPUs** on the test cluster.  
The returned slice is used by tests that validate CPU‑affinity, NUMA placement or other CPU‑related guarantees.

### Signature
```go
func (te TestEnvironment) GetGuaranteedPodContainersWithExclusiveCPUs() []*Container
```

- **Receiver:** `TestEnvironment` – holds the state of a test run and has access to cluster information.  
- **Return value:** `[]*Container` – a list of container objects; each element corresponds to a single container in the environment that satisfies both conditions.

### How it works

| Step | Description |
|------|-------------|
| 1 | Calls `GetGuaranteedPodsWithExclusiveCPUs()` (defined elsewhere in the same package). This helper returns all pods that: <br>• are of *Guaranteed* QoS class, and <br>• have at least one container requesting exclusive CPUs. The result is a slice of `*Pod`. |
| 2 | Calls `getContainers(pods []*Pod) []*Container` (private helper). This walks through each pod and collects every container in that pod into a flat slice. |
| 3 | Returns the slice produced by `getContainers`. |

No additional filtering or state mutation occurs; the function is purely read‑only.

### Dependencies

- **External functions**
  - `GetGuaranteedPodsWithExclusiveCPUs()` – provides the set of pods to filter.
  - `getContainers(pods []*Pod)` – flattens pods into containers.

- **Types used**
  - `*Container` – defined in `containers.go`.  
  - `*Pod` – defined elsewhere in the provider package (usually a wrapper around Kubernetes Pod objects).

### Side effects
None. The function does not modify the environment, pods, or any global state. It only reads data that has already been gathered by the test harness.

### Package context
Within the **provider** package this method is part of the *filter* utilities that help test scenarios pick relevant subsets of objects from a cluster snapshot:

```
filters.go  ← contains GetGuaranteedPodContainersWithExclusiveCPUs
pods.go     ← defines pod‑related helpers (including GetGuaranteedPodsWithExclusiveCPUs)
containers.go ← defines Container struct and getContainers helper
```

The method is typically invoked by test cases that need to assert properties about CPU allocation on containers, e.g.:

```go
for _, c := range env.GetGuaranteedPodContainersWithExclusiveCPUs() {
    // validate CPU affinity or NUMA placement
}
```

### Diagram (Mermaid)

```mermaid
flowchart TD
    A[GetGuaranteedPodsWithExclusiveCPUs] --> B[getContainers]
    B --> C[Return []*Container]
```

This diagram shows the simple linear flow: first gather the pods, then flatten to containers.

--- 

**Bottom line:**  
`TestEnvironment.GetGuaranteedPodContainersWithExclusiveCPUs()` is a lightweight helper that provides all containers running in guaranteed‑QoS pods with exclusive CPU requests, ready for downstream validation logic.
