Check.SetResultAborted`

```go
func (c Check) SetResultAborted(msg string) func()
```

### Purpose

`SetResultAborted` marks a **check** as *aborted* and records an optional message that explains why the check could not be executed.  
It returns a function that, when called, restores the previous state of the check (i.e., it re‑locks the check after the abort has been recorded). This pattern is used by callers to ensure the lock is released even if the caller panics or returns early.

### Parameters

| Name | Type   | Description |
|------|--------|-------------|
| `msg` | `string` | A human‑readable message that will be stored in the check’s `ResultMsg`. If empty, no message is set. |

> **Note**: The function does *not* modify any global state directly; it only touches the fields of the receiver.

### Return value

| Type | Description |
|------|-------------|
| `func()` | A closure that calls `c.Unlock()`. It should be invoked by the caller to release the check’s lock. |

> The returned function is deliberately simple: it only unlocks the mutex associated with this particular `Check`. This design mirrors other “SetResult…” methods in the package, which also return a cleanup closure.

### Key Steps & Side Effects

1. **Lock acquisition**  
   The method begins by calling `c.Lock()`. Since `Check` embeds a `sync.Mutex`, this guarantees exclusive access to the check’s fields while it is being mutated.

2. **State mutation**  
   * `c.Result` is set to `CheckResultAborted` (a constant defined in `check.go`).  
   * If `msg` is non‑empty, `c.ResultMsg` is populated with that message; otherwise the existing message remains unchanged.

3. **Return cleanup function**  
   The method returns a closure that simply calls `c.Unlock()`. This allows callers to defer or manually invoke it to release the lock after performing any additional logic (e.g., logging, metrics).

4. **No global side effects**  
   Apart from modifying the receiver’s fields, the method does not touch package‑level globals such as `dbByGroup` or `resultsDB`.

### How It Fits in the Package

* The `checksdb` package manages a registry of checks (`Check`) and groups them into `ChecksGroup`.  
* Each `Check` has a lifecycle: it may **pass**, **fail**, be **skipped**, or become **aborted**.  
* Methods like `SetResultPassed`, `SetResultFailed`, etc., all follow the same pattern—locking, mutating state, and returning an unlock closure.  
* `SetResultAborted` is used in scenarios where a check cannot proceed (e.g., missing prerequisites). It signals to callers that no further evaluation should occur for this check.

### Example Usage

```go
func runCheck(c checksdb.Check) {
    // Mark as aborted if pre‑condition fails
    defer c.SetResultAborted("missing required config")()

    // ... perform actual check logic ...
}
```

In the example above, `defer` ensures that the lock is released when the function exits, regardless of whether it aborts early or completes successfully.

### Summary

- **What**: Marks a check as aborted with an optional message.  
- **Why**: Signals that the check cannot be executed and preserves the abort state for reporting.  
- **How**: Locks the check, updates `Result`/`ResultMsg`, returns an unlock closure.  
- **Where**: Part of the consistent set‑result API within the `checksdb` package.
