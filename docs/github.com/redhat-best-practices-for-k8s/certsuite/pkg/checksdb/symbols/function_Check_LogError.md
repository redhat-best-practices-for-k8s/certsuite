Check.LogError`

```go
func (c Check) LogError(format string, args ...any) func()
```

### Purpose
`LogError` is a convenience wrapper around the package‑wide logging facility that:
1. Logs an error message for a specific check.
2. Returns a *deferred* function that will record the error in the check’s result set when invoked.

It is intended to be used with `defer` inside the execution of a check, e.g.:

```go
func (c Check) Run(ctx context.Context) {
    defer c.LogError("failed to load config: %v", err)()
    // ...check logic...
}
```

When the deferred function runs, it:
* Emits the formatted error message via `Logf`.
* Updates the check’s internal result slice with an entry of type `CheckResultError`.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `format` | `string` | printf‑style format string for the log message. |
| `args ...any` | variadic | Arguments that will be interpolated into `format`. |

> **Note**: The function does *not* evaluate or store the arguments immediately; they are captured in a closure and used only when the returned function is called.

### Return Value
A zero‑argument function (`func()`) which, when executed:
1. Calls `Logf` (the internal logging helper) with the same format and arguments to emit the message.
2. Appends an error result (`CheckResultError`) to the check’s `results` slice.

The returned function is typically used as a deferred call, ensuring that the log entry and result are recorded regardless of how the surrounding function exits (panic or normal return).

### Key Dependencies
| Dependency | Role |
|------------|------|
| `Logf` | The actual logging routine that prints the message to the configured logger. It expects the same signature as `fmt.Printf`. |
| `CheckResultError` constant | Used to mark the result entry as an error. |
| `Check.results` slice | Holds all result entries for this check; the deferred function appends a new one. |

### Side Effects
* The log output is produced immediately when the returned function runs.
* The check’s internal state (`results`) is mutated, adding a new error record.

No other global state or external systems are affected.

### Package Context
`Check.LogError` lives in the `checksdb` package, which manages a registry of checks and their results.  
Checks are defined as structs that implement an execution interface; each check has a `results` field where outcomes (`passed`, `failed`, `error`, etc.) are stored.  
`LogError` provides a standard way for checks to report failures in a consistent format while ensuring the result is captured for later aggregation.

### Suggested Mermaid Flow
```mermaid
flowchart TD
    A[Check.Run] --> B{defer c.LogError(...)}
    B --> C[Deferred function executed]
    C --> D[Logf emits message]
    C --> E[Append CheckResultError to results]
```

This diagram illustrates the typical usage pattern and the two actions performed by the deferred function.
