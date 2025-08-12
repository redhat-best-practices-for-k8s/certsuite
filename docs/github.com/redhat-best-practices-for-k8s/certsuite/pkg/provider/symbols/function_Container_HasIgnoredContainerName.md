Container.HasIgnoredContainerName`

**File:** `pkg/provider/containers.go:167`  
**Exported?** Yes – it can be called from other packages that import `provider`.

---

### Purpose
Determines whether a container instance should be skipped when the provider runs its checks.  
A container is considered *ignored* if:

1. It is an Istio side‑car proxy (`IsIstioProxy` returns true), **or**
2. Its name appears in the global list `ignoredContainerNames`.

The function is used by the test runner to filter out containers that are not relevant for the certification checks (e.g., platform‑injected side‑cars or known non‑critical workloads).

---

### Receiver
```go
func (c Container) HasIgnoredContainerName() bool
```
`c` represents a single container from a pod; the method inspects only its `Name` field.

---

### Dependencies

| Call | Type | Purpose |
|------|------|---------|
| `IsIstioProxy()` | function on `Container` | Detects if the container is an Istio proxy. The implementation typically checks for a specific image name or label. |
| `Contains(slice []string, item string)` | utility function | Checks whether `item` exists in the slice of ignored names (`ignoredContainerNames`). |

No global variables are directly accessed; all required data comes from the receiver and the two helper functions.

---

### Side‑effects
None – the method only reads state and returns a boolean. It does not modify any package level or struct fields.

---

### Integration with the `provider` package

The `HasIgnoredContainerName` method is part of the **Container** type, which is used throughout the provider to model pod containers.  
When the provider iterates over all pods:

```go
for _, c := range pod.Spec.Containers {
    if !c.HasIgnoredContainerName() {
        // run checks on this container
    }
}
```

Thus, it acts as a gatekeeper that prevents unnecessary or harmful checks from being applied to known side‑cars (like Istio) or containers listed in the global `ignoredContainerNames` slice.

---

### Typical usage pattern

```go
if container.HasIgnoredContainerName() {
    // skip this container – no tests run against it
}
```

Because the logic is encapsulated, callers need not know how “ignoring” is determined; they simply ask the container if it should be ignored.

---

### Summary diagram (Mermaid)

```mermaid
graph LR
  C[Container] -->|HasIgnoredContainerName()| I(IstioProxy?)
  I -- yes --> IG[Ignore]
  C -->|Contains(ignoredContainerNames, Name)| IG2[Ignore]
  IG -.-> SKIP[Skip checks]
```

The method returns `true` when either branch leads to *Ignore*, signalling the caller to skip that container.
