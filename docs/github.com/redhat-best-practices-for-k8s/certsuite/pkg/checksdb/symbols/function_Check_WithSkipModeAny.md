Check.WithSkipModeAny`

| Aspect | Detail |
|--------|--------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb` |
| **Receiver type** | `Check` (value receiver) |
| **Signature** | `func() *Check` |
| **Exported** | ✅ |

### Purpose
`WithSkipModeAny` is a convenience modifier for a `Check`.  
When a check is run it may be configured to skip certain sub‑checks.  
There are two skip modes:

- **SkipModeAll** – all configured skips apply (default for some checks).
- **SkipModeAny** – the check fails if *any* of its configured skips match.

`WithSkipModeAny` explicitly sets a check’s skip mode to `SkipModeAny`.  
In practice this is usually unnecessary because `SkipModeAny` is already the
default when a check has no explicit skip mode set. The method exists only for
API symmetry and code readability.

### Inputs / Outputs

| Parameter | Type | Description |
|-----------|------|-------------|
| *none* | – | No parameters are required; the function operates on the receiver. |

| Return value | Type | Description |
|--------------|------|-------------|
| `*Check` | Pointer to the same `Check` instance (modified) | The method returns a pointer so that callers can chain
modifiers, e.g., `check.WithSkipModeAny().WithSomeOtherModifier()`.

### Key Dependencies

- **`skipMode` type** – an internal enum representing skip modes.  
  `SkipModeAny` is one of its constants (see `check.go` line 26).
- No external globals or package variables are touched.
- The method does not trigger any database operations; it merely mutates
the `Check` object.

### Side Effects

The only side effect is setting the `skipMode` field on the receiver to
`SkipModeAny`. This change affects how the check will behave when evaluated:
if any of its skip expressions match, the check will be considered *skipped*
rather than *passed* or *failed*.

Because the method returns a pointer to the same instance, it can be used
in fluent chains without allocating new objects.  
No locks (`dbLock`) or other shared state are involved.

### Relationship within the Package

- **`Check` struct** – represents an individual test/check that can be executed against a Kubernetes cluster.
- **Skip Modes** – part of the check configuration that determines how skip rules influence execution results.
- **Other modifiers** (e.g., `WithSkipModeAll`) provide alternative ways to configure the same field.  
  `WithSkipModeAny` is essentially a no‑op in terms of effect but improves API symmetry.

### Usage Example

```go
// Create a new check and explicitly set SkipModeAny for clarity.
chk := NewCheck("my-check").
    WithLabel("category", "security").
    WithSkipModeAny() // optional – default behaviour.

// Run the check later...
result, err := chk.Execute(ctx)
```

### Summary

`WithSkipModeAny` is a lightweight, declarative way to set a `Check`’s skip mode
to “any” (the default). It modifies the check in place and returns a pointer,
facilitating fluent configuration. The method has no side effects beyond that
field change and does not interact with global state or persistence layers.
