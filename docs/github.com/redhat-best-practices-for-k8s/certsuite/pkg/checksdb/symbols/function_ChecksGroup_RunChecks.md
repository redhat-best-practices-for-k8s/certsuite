## `RunChecks` – Execute a group of checks

### Signature
```go
func (cg *ChecksGroup) RunChecks(abort <-chan bool, out chan string) ([]error, int)
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `abort`   | `<-chan bool` | A read‑only channel that signals an abort. When a value is received the group stops processing further checks and performs cleanup. |
| `out`     | `chan string` | Channel used to emit status messages (e.g., progress, errors). The caller typically reads from this channel concurrently. |

### Return values
1. **`[]error`** – slice of all errors that occurred during the run.  
   * Each element corresponds to a check or lifecycle hook that failed/panic’ed.  
2. **`int`** – number of checks that were actually executed (excluding those skipped by `skipCheck`).  

### High‑level flow

| Step | Action |
|------|--------|
| 1 | Call the group's `BeforeAll()` hook (`runBeforeAllFn`). If it panics or returns an error, treat the first check as failed and run `AfterAll()`. |
| 2 | Iterate over all checks in the group that match the label expression filter. For each check: |
| &nbsp;&nbsp;2a | Run `BeforeEach()` (`runBeforeEachFn`). If it fails, record the error, skip remaining checks, call `AfterEach()` and `AfterAll()`. |
| &nbsp;&nbsp;2b | Determine whether to skip the check via `skipCheck`/`shouldSkipCheck`. If skipped, emit a status message. |
| &nbsp;&nbsp;2c | Execute the actual check (`runCheck`). Record any error or panic as “panicked”. |
| &nbsp;&nbsp;2d | Run `AfterEach()` (`runAfterEachFn`) and handle its errors. |
| 3 | After all checks (or after an abort), call `AfterAll()` (`runAfterAllFn`). |
| 4 | Return the collected errors slice and the count of executed checks. |

### Key dependencies

* **Lifecycle helpers** – `runBeforeAllFn`, `runBeforeEachFn`, `runCheck`, `runAfterEachFn`, `runAfterAllFn`.  
  These are thin wrappers that execute the corresponding user‑supplied functions and capture panics/errors.
* **Label evaluation** – `labelsExprEvaluator.Eval` is used to decide if a check should be included in this run.
* **Abort handling** – The abort channel is monitored during the loop; receiving a value triggers early exit after cleanup.
* **Output formatting** – Uses `Info`, `Printf`, and `Join` from the package’s logging utilities to send human‑readable messages through `out`.

### Side effects

1. **State mutation** – Checks’ internal state (e.g., environment variables set in `BeforeEach`) may change, but the group itself remains immutable after construction.
2. **Error collection** – Errors are appended to a slice that is returned; no global error store is modified.
3. **Channel writes** – The function writes status messages to `out`; it never closes the channel (caller decides).
4. **Global DB access** – None; all operations are confined to the receiver and its checks.

### Placement in the package

`RunChecks` is the central execution engine for a `ChecksGroup`.  
A `ChecksGroup` represents a logical collection of CNF‑certification checks (e.g., “Security” or “Networking”).  
The function orchestrates lifecycle hooks, skip logic, error handling, and result aggregation, making it the primary public API used by test runners or CI pipelines to evaluate a group’s compliance.
