ChecksGroup.WithBeforeEachFn`

| Feature | Details |
|---------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/checksdb` |
| **Exported?** | ✅ |
| **Receiver type** | `ChecksGroup` (value receiver) |
| **Signature** | `func(func(check *Check) error)(*ChecksGroup)` |

### Purpose

`WithBeforeEachFn` is a convenience method that allows callers to register a **global “before‑each” hook** for all checks that belong to the current `ChecksGroup`.  
The provided function will be executed once for every individual `Check` in the group before that check runs. This enables per‑check setup such as:

* initializing test data,
* configuring environment variables,
* logging context, or
* performing common validation steps.

### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `fn` | `func(check *Check) error` | A callback that receives a pointer to the check being processed. It can return an error to abort the execution of that specific check (see `CheckResultAborted`). |

### Return Value

* The same `ChecksGroup` instance (`*ChecksGroup`) is returned, enabling **method chaining**:

```go
group.WithBeforeEachFn(setup).WithAfterEachFn(teardown)
```

The returned value is the receiver itself after it has been mutated to hold the new hook.

### Side Effects & Dependencies

| Aspect | Effect |
|--------|--------|
| **Internal state** | The `ChecksGroup` struct stores the supplied function in a field (likely named something like `beforeEachFn`). This field is consulted when the group’s checks are executed. |
| **No external side‑effects** | The method does not touch any global variables (`dbByGroup`, `resultsDB`, etc.). It merely registers the callback locally within the group. |
| **Error handling** | If the supplied function returns an error for a particular check, that check’s result will be marked as `CheckResultAborted`. The error is propagated to the caller of the check execution routine (e.g., test runner). |

### Interaction with Other Package Elements

* **`ChecksGroup` struct** – Holds metadata about the group and its checks. Adding a before‑each hook augments this metadata without affecting other groups.
* **Check execution flow** – When `RunAll()` or similar routines iterate over each `Check`, they will invoke the stored before‑each function prior to executing the check’s own logic.
* **Result constants** (`PASSED`, `FAILED`, `SKIPPED`) – Errors returned by the hook influence which of these states a check may end up in.

### Typical Usage

```go
group := NewChecksGroup("my-group")
group.WithBeforeEachFn(func(c *Check) error {
    // Common setup for every check
    if err := c.Prepare(); err != nil {
        return fmt.Errorf("setup failed: %w", err)
    }
    return nil
})
```

### Summary

`WithBeforeEachFn` is a small but powerful API that lets developers inject reusable pre‑execution logic into a collection of checks. By returning the group itself, it supports fluent configuration while keeping the internal state encapsulated and thread‑safe (the package uses `dbLock` for global access, but this method operates on an individual group instance).
