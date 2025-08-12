## `RecordChecksResults`

| Feature | Details |
|---------|---------|
| **Package** | `checksdb` |
| **Receiver type** | `*ChecksGroup` |
| **Signature** | `func (cg *ChecksGroup) RecordChecksResults() func()` |
| **Exported?** | ✅ |

### Purpose
After all checks in a `ChecksGroup` have finished executing, this method produces a closure that can be invoked to:

1. Log a summary of the group’s execution via `Info`.
2. Persist the individual check results into the group database by calling `recordCheckResult`.

The returned function is intentionally deferred by callers so that the recording happens **after** all checks have run, guaranteeing that the result map contains every check's outcome.

### Inputs / Outputs
- **Input:** none – it operates on the receiver (`cg`).
- **Output:** a zero‑argument closure `func()` that performs the side effects described above.  
  The closure has no return value and may be safely deferred.

### Key Dependencies & Side Effects
| Dependency | Role |
|------------|------|
| `Info(msg string)` | Emits a log line with group name, execution duration and result counts (passed/failed/skipped). |
| `recordCheckResult(check *Check, result CheckResult)` | Stores the check’s outcome in the internal `results` map of the group. |

Both dependencies are methods on `ChecksGroup`; thus the closure captures the current state of the group at the time it is returned.

### How It Fits Into the Package
- **Execution Flow:**  
  1. A test harness creates a `ChecksGroup`, registers checks, and starts them (often concurrently).  
  2. After all goroutines finish, the harness defers `cg.RecordChecksResults()` to ensure that recording happens *after* execution.  
  3. When the deferred closure runs, it logs summary info and persists each check’s result.

- **Persistence:** The results are stored in the group's internal map and later exposed via `GetResults()` for reporting or further analysis.

### Suggested Mermaid Diagram
```mermaid
flowchart TD
    A[Start Test] --> B[Create ChecksGroup]
    B --> C{Run Checks (concurrently)}
    C --> D[All checks finished]
    D --> E[defer cg.RecordChecksResults()]
    E --> F[Record closure runs]
    F --> G[Info log + recordCheckResult for each check]
```

This function is the bridge between execution and reporting, ensuring that every check’s outcome is logged and stored exactly once.
