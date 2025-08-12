Pod.HasHugepages`

```go
func (p Pod) HasHugepages() bool
```

### Purpose
`HasHugepages` checks whether **any container** in a pod requests hugepage memory resources.  
The function returns `true` if at least one resource name contains the substring `"hugepage"`; otherwise it returns `false`.

This helper is used by tests that need to distinguish between regular workloads and those that rely on hugepage support (e.g., database or VM‑like containers).

### Inputs & Receiver
- **Receiver**: `p Pod` – the pod instance whose containers are inspected.
- No additional arguments.

### Output
- `bool`:  
  - `true` if a container’s resource list includes a key containing `"hugepage"`.  
  - `false` otherwise.

### Key Operations
1. **Iterate over all containers** in `p.Spec.Containers`.
2. For each container, iterate over its `Resources.Requests` map.
3. Convert the resource name to string (`String()`).
4. Use `strings.Contains(..., "hugepage")` to test for a hugepage request.

The same logic is repeated for `p.Spec.InitContainers`.

### Dependencies
- **Standard library**:  
  - `strings.Contains` – checks substring presence.  
  - `strings.Stringer` (via `resource.Quantity.String()`) – converts resource quantity names.
- No external packages or global state are accessed.

### Side‑Effects
None. The method is pure; it only reads the pod structure and returns a value.

### Package Context
`HasHugepages` lives in `pkg/provider/pods.go`.  
The provider package models Kubernetes objects (Pods, Nodes, etc.) for certsuite’s validation logic.  
This helper supports other functions that filter or validate pods based on resource characteristics, such as:

```go
if pod.HasHugepages() {
    // special handling for hugepage workloads
}
```

### Mermaid Flow (Optional)

```mermaid
flowchart TD
  A[Pod] --> B{Containers?}
  B -- yes --> C[Iterate Containers]
  C --> D{InitContainers?}
  D -- yes --> E[Iterate InitContainers]
  E --> F[Check Resources.Requests]
  F --> G{Contains "hugepage"?}
  G -- true --> H[Return true]
  G -- false --> I[Continue]
  I --> J[End loop]
  J --> K[Return false]
```

This diagram visualizes the short, linear checks performed by `HasHugepages`.
