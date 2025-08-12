recordCheckResult`

```go
func(record *Check) func()
```

### Purpose
`recordCheckResult` creates a **deferred callback** that records the outcome of a check run into the package‑wide `resultsDB`.  
It is intended to be used with Go’s `defer` statement so that the result is stored automatically when the surrounding function exits.

### Inputs
| Parameter | Type   | Description |
|-----------|--------|-------------|
| `record`  | `*Check` | The check instance whose result will be recorded. |

The returned closure has no parameters and returns nothing; it mutates shared state.

### Returned value
A **zero‑argument function** that, when invoked, writes the final status of `record` into `resultsDB`.  
The caller is expected to use it as:

```go
defer recordCheckResult(myCheck)()
```

so that the result is recorded even if the surrounding function panics.

### Key Dependencies & Side‑Effects

| Dependency | Role |
|------------|------|
| `dbLock` (`sync.Mutex`) | Protects concurrent writes to `resultsDB`. The closure locks, updates, then unlocks. |
| `resultsDB` (map) | Stores the mapping from check ID → `CheckResultInfo`. |
| Logging helpers (`LogFatal`, `LogInfo`) | Emit diagnostic messages when a check fails or passes. |
| Time functions (`time.Now()`, `Sub`, `Seconds`) | Compute execution duration for the check. |
| Label evaluator (`labelsExprEvaluator`) | (Used indirectly by other parts of the package; not directly in this function.) |

Side effects include:

1. **Updating global state** – the result entry is written to `resultsDB`.
2. **Logging** – a fatal or info message is emitted depending on the outcome.
3. **Duration calculation** – the elapsed time between check start and finish is recorded.

### How it fits the package

- The `checksdb` package manages registration, execution, and result collection of checks.  
- `recordCheckResult` is part of the *result‑collection* mechanism: after a check runs, this deferred function ensures its outcome is persisted in the shared database (`resultsDB`) and logged appropriately.
- Other functions (e.g., `RunChecks`, `GetResults`) rely on `resultsDB` to present aggregated results or generate reports.

---

#### Suggested Mermaid diagram

```mermaid
flowchart TD
    A[Check Execution] -->|defer recordCheckResult(check)| B{Deferred Callback}
    B --> C[Lock dbLock]
    B --> D[Compute duration]
    B --> E[Log message (Fatal/Info)]
    B --> F[Store in resultsDB]
    B --> G[Unlock dbLock]
```

This illustrates the flow from check execution to final result persistence.
