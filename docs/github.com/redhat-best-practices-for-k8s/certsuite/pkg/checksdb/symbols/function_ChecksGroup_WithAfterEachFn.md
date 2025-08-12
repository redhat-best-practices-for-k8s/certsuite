ChecksGroup.WithAfterEachFn`

> **Signature**  
> ```go
> func (cg *ChecksGroup) WithAfterEachFn(fn func(check *Check) error) *ChecksGroup
> ```

### Purpose
`WithAfterEachFn` attaches a user‑supplied callback that will be executed after each individual check in the group has finished. The callback receives the `*Check` that just ran and may perform any side effect (logging, cleanup, metrics, etc.).  
The function returns the same `ChecksGroup` instance to allow method chaining when building a group.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `fn` | `func(check *Check) error` | The callback invoked after each check. It receives a pointer to the finished `Check`. If it returns an error, that error is ignored by this method; it is only stored for reporting purposes (see `Check.Result`). |

### Return value
* `*ChecksGroup` – the same group instance (`cg`) so calls can be chained.

### Key dependencies
- **`Check` type** – The callback operates on a `*Check`. The `Check` struct (defined in `check.go`) holds metadata such as ID, description, and the result of the test.  
- **No external globals or side‑effects** – The method only updates an internal field (`afterEachFn`) of the `ChecksGroup`; it does not touch shared state like `dbByGroup` or `resultsDB`.

### Side effects
1. Assigns the supplied function to the group’s private `afterEachFn` field.  
2. Does **not** trigger any checks immediately; execution happens when the group runs its tests.

Because it only stores a reference, calling `WithAfterEachFn(nil)` will disable the after‑each callback for that group.

### How it fits in the package
`ChecksGroup` represents a collection of related checks (e.g., all tests for a particular Kubernetes resource).  
During test execution, each check is run and then the group's `afterEachFn`, if set, is invoked. This allows fine‑grained control over per‑check cleanup or reporting without polluting the core logic that runs the checks.

Typical usage pattern:

```go
group := NewChecksGroup("networking")
group.WithAfterEachFn(func(c *Check) error {
    // e.g., write result to a file
    return nil
})
```

This method is part of the fluent API for configuring a `ChecksGroup`. It has no direct impact on global registries (`dbByGroup`, `resultsDB`) or concurrency primitives (`dbLock`).
