Pod.IsShareProcessNamespace`

### Purpose
`IsShareProcessNamespace` reports whether the pod is configured to **share its process namespace** with other containers in the same pod.  
In Kubernetes this is controlled by the field `spec.shareProcessNamespace`.  
When set to `true`, all containers inside the pod can see each other’s processes and signals.

### Signature
```go
func (p Pod) IsShareProcessNamespace() bool
```

- **Receiver**: `Pod` – a lightweight wrapper around a Kubernetes pod object.  
- **Return value**: `bool`
  - `true` if the pod has `spec.shareProcessNamespace == true`.  
  - `false` otherwise (including when the field is omitted or set to `false`).

### Implementation details
```go
func (p Pod) IsShareProcessNamespace() bool {
    return p.spec.ShareProcessNamespace != nil && *p.spec.ShareProcessNamespace
}
```

- The function accesses the underlying pod spec (`p.spec`) and checks the pointer field `ShareProcessNamespace`.  
- A `nil` pointer means the field is not set, which Kubernetes treats as `false`.

### Dependencies & Side‑Effects

| Dependency | Role |
|------------|------|
| `Pod.spec.ShareProcessNamespace` | The actual configuration value. |
| No external packages or globals are touched. |

The function has **no side effects**: it only reads state from the pod object.

### How It Fits the Package
- **Package**: `provider`.  
- **Role in provider**: Provides a convenient method for tests and tooling that need to decide whether to enable certain features (e.g., signal handling, process discovery) that require a shared namespace.  
- **Interaction with other code**:  
  - Other helpers such as `Pod.IsPrivileged()` or `Pod.HasInitContainer()` also expose boolean checks on pod fields; this method follows the same pattern.  
  - Tests that iterate over all pods may call `IsShareProcessNamespace` to filter pods before performing namespace‑specific assertions.

### Usage Example
```go
for _, p := range allPods {
    if p.IsShareProcessNamespace() {
        fmt.Println(p.Name, "shares its process namespace")
    }
}
```

This function is a small but essential part of the provider’s introspection utilities, enabling higher‑level logic to adapt based on pod configuration.
