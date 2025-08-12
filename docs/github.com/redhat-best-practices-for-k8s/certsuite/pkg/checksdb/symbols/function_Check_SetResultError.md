Check.SetResultError`

```go
func (c *Check) SetResultError(errMsg string) func()
```

### Purpose
`SetResultError` marks a check as **failed** due to an error that occurred while executing it.  
The returned closure, when invoked, performs the following:

1. Locks the global `dbLock` mutex to ensure exclusive access to shared state.
2. Sets the check’s `result` field to `CheckResultError`.
3. Records the provided `errMsg` in the check’s `errorMessage` field.
4. Releases the lock.
5. Emits a warning log via `LogWarn`.

This pattern allows callers to defer the finalization of an error state until after additional cleanup or processing has completed, while still guaranteeing that the check result is updated atomically.

### Parameters
| Name | Type   | Description |
|------|--------|-------------|
| `errMsg` | `string` | Human‑readable message describing the failure. |

### Return Value
A **closure** (`func()`) with no arguments and no return value.  
When called, it performs the state mutation described above.

### Key Dependencies & Side Effects

- **Synchronization**
  - Uses the package‑level `dbLock` (`sync.Mutex`) to protect concurrent writes to shared structures such as `resultsDB`.
- **Logging**
  - Calls `LogWarn(errMsg)` to emit a warning log. This function is defined elsewhere in the package and typically writes to stdout/stderr or a logging framework.
- **Check State Mutation**
  - Sets the check’s internal fields:
    ```go
    c.result = CheckResultError
    c.errorMessage = errMsg
    ```
  - No other fields are modified.

### How It Fits the Package

`SetResultError` is part of the `Check` type, which represents an individual test or validation in CertSuite’s checks database.  
The package provides a consistent set of result constants (`CheckResultPassed`, `CheckResultFailed`, etc.) and global state (`resultsDB`, `dbByGroup`).  

When a check encounters an unexpected condition (e.g., failure to query a cluster, malformed data), the caller creates this closure via `SetResultError` and defers its execution. This ensures that:

- The error state is recorded atomically.
- Any necessary cleanup can happen before marking the check as failed.
- Logging occurs immediately when the closure runs.

Overall, `Check.SetResultError` provides a safe, deferred mechanism for recording errors in the checks database while maintaining thread‑safety and consistent logging.
