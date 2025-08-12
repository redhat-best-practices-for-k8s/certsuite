runCheck`

```
func runCheck(c *Check, g *ChecksGroup, deps []*Check) error
```

### Purpose  
`runCheck` executes a single test (`*Check`) inside the context of a check group (`*ChecksGroup`).  
It handles:

1. **Dependency resolution** – ensures that all checks listed in `deps` have already run and were successful.
2. **Execution** – calls the underlying check function via its exported `Run()` method.
3. **Result handling** – records the outcome (`Passed`, `Failed`, `Error`, `Skipped`, or `Aborted`) on the `Check` object, logs any failures, and triggers failure callbacks.

The function is intentionally *private*; callers in the package orchestrate execution order through `ChecksGroup.Run()`.

### Parameters

| Name | Type | Description |
|------|------|-------------|
| `c`  | `*Check` | The check to execute. Holds metadata (`ID`, `Name`, `Result`, etc.) and a reference to its implementation. |
| `g`  | `*ChecksGroup` | Group that owns the check; used for logging context and accessing group‑wide settings (e.g., `OnFailure`). |
| `deps` | `[]*Check` | Checks that must have completed successfully before this one runs. They are typically the elements of `c.Depends`. |

### Return Value

| Type | Meaning |
|------|---------|
| `error` | Non‑nil if an internal error occurs (e.g., panic recovery, missing dependencies). The caller may ignore it; the check’s own result field is still set. |

### Key Dependencies & Side Effects

| Dependency | How it’s used |
|------------|---------------|
| `c.Run()` | Executes the actual test logic and returns a `CheckResult` enum. |
| `c.OnFailure` | If the check fails or errors, this callback (if non‑nil) is invoked with the error message. |
| `Warn`, `Errorf`, `LogError` | Logging utilities from the surrounding package; used for diagnostic output. |
| `Stack()` | Provides stack trace on panic recovery to aid debugging. |
| `Sprintf`, `Sprint`, `string` | Helper formatting functions for log messages. |
| `dbLock`, `dbByGroup` | *Not* accessed directly in this function; they are part of the broader package but may be used by callers that wrap `runCheck`. |

### Execution Flow

1. **Dependency Check**  
   - Iterate over `deps`; if any has a result other than `Passed`, skip execution (`c.Result = Skipped`) and log a warning.

2. **Panic Recovery**  
   - A deferred function catches panics during `Run()`. On panic, the check’s result is set to `Error`, an error message is logged (including stack trace), and the error is returned.

3. **Actual Run**  
   - Calls `c.Run()` which returns a `CheckResult`.
   - Maps that enum to string for logging (`PASSED`, `FAILED`, etc.).

4. **Failure Handling**  
   - If result is `Failed` or `Error`, invoke `onFailure(g, c)` to run any group‑level failure handlers.
   - If the check has an individual `OnFailure` callback, it’s called with the error string.

5. **Return**  
   - The function returns any recovered panic error; otherwise nil.

### Integration in the Package

- `ChecksGroup.Run()` iterates over all checks in the group and calls `runCheck` for each.
- The results stored on each `*Check` are later aggregated by higher‑level components (e.g., report generators).
- Because `runCheck` is *not exported*, external callers must use the public API (`ChecksGroup.Run()`) to trigger execution.

### Mermaid Diagram (suggested)

```mermaid
flowchart TD
    A[Start] --> B{Dependencies OK?}
    B -- No --> C[Set Result=Skipped]
    B -- Yes --> D[Run Check()]
    D --> E{Result}
    E -->|Passed| F[Log Passed]
    E -->|Failed/Err| G[Invoke onFailure, Log]
    E -->|Aborted| H[Handle Aborted]
    C & F & G & H --> I[Return error?]
```

This function is the core execution engine for individual checks and ensures robust handling of dependencies, failures, and panics while keeping side‑effects confined to logging and callback invocation.
