Node.HasWorkloadDeployed`

```go
func (n Node) HasWorkloadDeployed(pods []*Pod) bool
```

### Purpose  
Determines whether the node on which the `Node` instance resides has at least one *non‑terminated* workload pod running. In other words, it answers **“is there any pod that represents a user workload on this node?”** This is used by CertSuite to decide if certain tests (e.g., connectivity or security) should be executed against a node.

### Parameters

| Name | Type | Description |
|------|------|-------------|
| `pods` | `[]*Pod` | A slice of pointers to the pod objects that are currently scheduled on the node. The caller is expected to filter pods by this node before invoking the method. |

### Return Value

- **bool** –  
  *`true`* if at least one pod in the list is a user workload and is not terminated; otherwise *`false`*.

### Key Dependencies & Internal Logic (inferred)

| Dependency | Role |
|------------|------|
| `ignoredContainerNames` | A global slice of container names that should be excluded when determining if a pod is a workload. For example, system containers such as `coredns`, `kube-proxy`, or side‑car proxies may appear in the pod but do not count as user workloads. |
| `Node` struct (receiver) | Contains metadata about the node; typically its name and labels are used to correlate pods with the node. The method does **not** modify the node. |

The implementation likely follows this pattern:

1. Iterate over each `Pod` in `pods`.
2. Skip pods that are in a terminal phase (`Succeeded`, `Failed`) or whose status indicates they are no longer running.
3. For each non‑terminated pod, inspect its containers:
   - If *all* container names are present in `ignoredContainerNames`, the pod is considered system‑only and ignored.
4. If any pod passes step 3 (i.e., contains at least one non‑ignored container), return `true`.
5. After checking all pods, if none matched, return `false`.

### Side Effects

- The method has **no side effects**: it does not modify the node or any pod; it only reads their state.
- It may read global configuration (`ignoredContainerNames`), but that is read‑only.

### How It Fits Into the Package

Within the `provider` package, CertSuite needs to know which nodes actually host user workloads before running certain tests (e.g., security hardening checks or connectivity validations).  
`HasWorkloadDeployed` provides a lightweight predicate that:

- Helps filter out control plane or infrastructure‑only nodes.
- Enables the test harness to skip expensive or irrelevant checks on nodes that do not run workloads.

This function is called by higher‑level orchestration code (e.g., node selection logic) and forms part of the decision tree for test execution.
