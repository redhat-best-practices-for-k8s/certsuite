Pod.IsRuntimeClassNameSpecified`

| Item | Details |
|------|---------|
| **Receiver** | `Pod` – a struct representing a Kubernetes pod (defined in `pods.go`). |
| **Signature** | `func () bool` |
| **Exported** | Yes – can be called by other packages. |

### Purpose
Determines whether the pod has a non‑empty *RuntimeClassName* field set, indicating that the pod requests a specific runtime class (e.g., `kata`, `runc`).  
The method returns `true` only when the field exists and contains at least one character; otherwise it returns `false`.

### Inputs / Outputs
- **Input** – none. The method inspects the receiver’s internal state.
- **Output** – a boolean:
  - `true`: `RuntimeClassName` is present and non‑empty.
  - `false`: field missing or empty.

### Key Dependencies
* Relies on the struct field `RuntimeClassName` within the `Pod` type.  
* No external packages, globals, or side‑effects are used; it purely reads pod data.

### Side Effects
None – the function is read‑only and does not modify the pod or any global state.

### Package Context
Within the **provider** package, this helper is part of a suite that validates pod configurations.  
It is used by higher‑level checks (e.g., verifying runtime class usage across deployments) to quickly decide whether further inspection of the `RuntimeClassName` value is necessary. The method complements other Pod‑related helpers such as `IsContainerImageSpecified` or `IsCommandSpecified`.

### Usage Example
```go
p := Pod{RuntimeClassName: "kata"}
if p.IsRuntimeClassNameSpecified() {
    // Proceed with runtime class specific checks
}
```

The function is intentionally simple to keep the validation logic modular and testable.
