ChecksGroup.WithAfterAllFn`

### Purpose
`WithAfterAllFn` attaches a user‑supplied callback that will be invoked **once** after all checks in the group have finished running.  
The callback receives the slice of `*Check` objects that were executed and can inspect or modify their results, write logs, clean up resources, etc.

### Receiver
```go
func (cg *ChecksGroup) WithAfterAllFn(fn func([]*Check) error) *ChecksGroup
```
- **`cg`** – the group to which the callback will be attached.  
  `ChecksGroup` is defined in `checksgroup.go`; it holds metadata about a collection of checks and the execution mode (e.g., skip logic).

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `fn` | `func([]*Check) error` | A function that accepts the slice of checks run in this group. The function may return an error to signal a failure in the after‑all phase.

### Return value
- Returns the same `ChecksGroup` instance (`cg`) to allow method chaining (builder pattern).

### How it works
1. **Storage** – The provided function is stored inside the `ChecksGroup` struct under the field that holds the *after‑all* callback (not shown in the snippet but defined elsewhere in `checksgroup.go`).  
2. **Invocation** – When the group’s execution engine completes all checks, it looks up this field and calls the function with the slice of executed checks.  
3. **Error handling** – If the callback returns an error, the framework treats it as a failure for the entire group (typically logged and marked as `FAILED` in the results database).

### Dependencies
- Relies on the internal `ChecksGroup` struct; no external packages are called directly from this method.
- The function signature uses `*Check`, which is defined in `check.go`.  
- The callback may interact with global state (`resultsDB`, `dbLock`) if it needs to record results, but that is optional.

### Side effects
- None until the group finishes executing; at that point, any side effect performed by the supplied function will occur.  
- No modification of global maps or locks occurs directly inside `WithAfterAllFn`.

### Package context
`ChecksGroup` lives in the **checksdb** package, which manages a registry of check groups and orchestrates their execution against workloads.  
Adding an after‑all function is part of the *fluent API* that lets callers customize behavior per group:

```go
group := NewChecksGroup("my-group").
    WithAfterAllFn(func(checks []*Check) error {
        // custom cleanup or reporting
        return nil
    })
```

This method thus provides a hook for post‑processing after all checks in the group have run, fitting into the overall lifecycle of a check group execution.
