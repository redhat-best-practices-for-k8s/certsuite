TestEnvironment.GetGuaranteedPodContainersWithIsolatedCPUsWithoutHostPID`

```go
func (te *TestEnvironment) GetGuaranteedPodContainersWithIsolatedCPUsWithoutHostPID() []*Container
```

### Purpose

Return all containers that belong to **guaranteed** pods with **isolated CPUs** and that are **not** running in the host PID namespace.

The method is used by tests that need to exercise CPU‑pinning logic on real workloads while excluding any pod that might interfere with isolation due to a shared PID namespace (`hostPID: true`).

### Inputs / Receiver

| Parameter | Type        | Description |
|-----------|-------------|-------------|
| `te`      | *TestEnvironment | The test environment holding the current cluster state (nodes, pods, containers, etc.). |

No other arguments.

### Output

* `[]*Container`: a slice of pointers to `Container` objects that satisfy all three criteria:
1. Pod has **Guaranteed** QoS class (`Requests == Limits` for CPU and memory).
2. Pod declares **isolated CPUs** (via the `CPUManagerPolicy: "static"` or similar annotation).
3. The pod does **not** have `hostPID: true`.

### Key Dependencies

| Called function | Role |
|-----------------|------|
| `getContainers()` | Retrieves all containers in the environment. |
| `filterPodsWithoutHostPID([]*Pod)` | Filters pods that are not using host PID namespace. |
| `GetGuaranteedPodsWithIsolatedCPUs()` | Returns pods that are guaranteed and have isolated CPUs. |

These helpers operate on the same `TestEnvironment` data structures, so the call chain is:

```
getContainers() -> filterPodsWithoutHostPID(...) -> GetGuaranteedPodsWithIsolatedCPUs() -> container list
```

### Side Effects

None – the method performs read‑only filtering of in‑memory objects. It does not modify any cluster state or configuration.

### Package Context

*Location*: `pkg/provider/filters.go`  
*Package*: `provider`

The `provider` package encapsulates logic for interacting with a Kubernetes test cluster. Within that context, this function is part of the *filtering utilities* that allow tests to narrow down the set of containers based on pod properties (QoS, CPU isolation, namespace behavior). It complements other filter helpers such as `GetGuaranteedPodContainersWithIsolatedCPUs` and `GetAllPodContainers`.

### Usage Example

```go
// Retrieve all relevant containers for a CPU‑pinning test.
containers := env.GetGuaranteedPodContainersWithIsolatedCPUsWithoutHostPID()
for _, c := range containers {
    // Run assertions or instrumentation on each container
}
```

This function is typically called early in a test suite that verifies CPU affinity, huge page usage, or other low‑level scheduling features.
