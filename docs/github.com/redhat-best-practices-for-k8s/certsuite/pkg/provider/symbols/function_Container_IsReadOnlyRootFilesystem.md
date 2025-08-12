Container.IsReadOnlyRootFilesystem`

**Package:** `provider`  
**File:** `pkg/provider/containers.go` (line 190)  
**Exported?** Yes

---

### Purpose
Determines whether the container’s root filesystem is mounted read‑only.

The function is intended to be used by the test suite when validating that a
container follows the security best practice of having its root volume marked
as read‑only.  It logs an informational message and returns a boolean value
representing the result of the check.

---

### Signature
```go
func (c *Container) IsReadOnlyRootFilesystem(lg *log.Logger) bool
```

| Parameter | Type          | Description |
|-----------|---------------|-------------|
| `lg`      | `*log.Logger` | Logger used to emit an info message. The logger is optional; if nil, the call will panic when dereferenced (the current implementation assumes a non‑nil logger). |

| Return | Type  | Meaning |
|--------|-------|---------|
| `bool` | `true` if the container’s root filesystem is read‑only, otherwise `false`. |

---

### Key Steps & Logic

1. **Log Context**  
   The function immediately logs an informational line using the provided
   logger:  
   ```go
   lg.Info("Checking if container %s has a read-only root filesystem", c.Name)
   ```
   This helps trace which containers are being evaluated during test runs.

2. **Read‑Only Check**  
   It inspects the `ReadOnlyRootFilesystem` field of the underlying
   Kubernetes container spec (`c.Spec.ReadOnlyRootFilesystem`).  
   - If this boolean is set to `true`, the function returns `true`.  
   - Otherwise, it returns `false`.

3. **No Side Effects on State**  
   The method only reads state from the `Container` struct and logs; it does
   not modify any fields or external resources.

---

### Dependencies

| Dependency | Role |
|-------------|------|
| `log.Logger` | Provides structured logging (`Info`). No other packages are called. |

The function itself is self‑contained: it uses only the container’s own data and a logger.

---

### Interaction with Package

- **Container struct**  
  The method lives on `Container`, which represents an individual pod/container
  in the test harness. It is part of the suite that evaluates various
  security‑related properties (e.g., capability bounds, image integrity).

- **Test Harness**  
  In tests such as `ValidateRootFilesystemReadOnly`, this method will be called
  for each container under inspection to produce a boolean result that feeds
  into assertions.

---

### Usage Example

```go
c := &Container{Spec: corev1.Container{ReadOnlyRootFilesystem: true, Name: "app"}}
if c.IsReadOnlyRootFilesystem(log.Default()) {
    fmt.Println("✅ Root FS is read‑only")
} else {
    fmt.Println("❌ Root FS is writable")
}
```

---

### Summary

`Container.IsReadOnlyRootFilesystem` is a small helper that logs its intent and
returns whether the container’s root filesystem has been configured as
read‑only.  It plays a role in the broader provider package by enabling tests
to verify this particular security best practice across all containers.
