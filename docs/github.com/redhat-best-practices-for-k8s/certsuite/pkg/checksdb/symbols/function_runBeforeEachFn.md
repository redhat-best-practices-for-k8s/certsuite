runBeforeEachFn`

| Feature | Detail |
|---------|--------|
| **Visibility** | Unexported (internal helper) |
| **Signature** | `func(*ChecksGroup, *Check, []*Check) error` |
| **Location** | `pkg/checksdb/checksgroup.go:157` |

### Purpose
`runBeforeEachFn` is a small wrapper that executes the optional `beforeEach` hook defined on a test suite or a single check.  
The hook can perform setup work (e.g., creating temporary resources) and may fail, in which case all dependent checks are marked as failed.

### Parameters

| Name | Type | Meaning |
|------|------|---------|
| `cg`  | `*ChecksGroup` | The group that owns the check. It contains configuration such as the `beforeEachFn`. |
| `c`   | `*Check`      | The specific check whose hook should run. |
| `subchecks` | `[]*Check` | Checks that belong to the same group and will be executed after this one; used only for logging context. |

### Return Value
- `error`:  
  - `nil` if the hook ran successfully or was not defined.  
  - A wrapped error describing why the hook failed (panics are captured). This error is propagated back to the caller, which will mark the current check and its dependents as **failed**.

### Key Dependencies

| Dependency | Role |
|------------|------|
| `cg.beforeEachFn` | The actual function supplied by a test suite. It receives the current check and all checks in the group. |
| `onFailure(*Check, error)` | Called when the hook fails to mark the check (and possibly its dependents) as failed and record the failure reason. |
| `Debug`, `Error` from the package's logging utilities | Emit diagnostic messages. |
| `recover()` + `runtime.Stack` | Capture panics thrown by the user‑supplied hook and turn them into errors. |

### Control Flow

1. **Early exit** – If no `beforeEachFn` is defined, the function returns `nil`.  
2. **Execution with panic guard** – The hook is executed inside a `defer recover()` block to catch panics.  
3. **Success path** – On normal return, the function logs success and returns `nil`.  
4. **Failure path** – If the hook returns an error or panics:
   * Construct a descriptive message (including stack trace for panics).
   * Call `onFailure` to mark the check as failed.
   * Return the error so that the caller can propagate it.

### Side Effects

- The function may change the state of the passed `*Check` by marking it **failed** via `onFailure`.  
- It emits log messages but does not modify global variables or other checks directly.  

### Role in the Package

Within the `checksdb` package, tests are organized into groups (`ChecksGroup`). Each group may define a *before‑each* hook that runs once before each check in that group.  
`runBeforeEachFn` is invoked by the test runner right before executing an individual check:

```go
err := runBeforeEachFn(cg, c, subchecks)
if err != nil {
    // The check (and possibly its dependents) are already marked failed.
}
```

Thus it bridges user‑supplied setup logic with the internal failure handling mechanism of the test framework.
