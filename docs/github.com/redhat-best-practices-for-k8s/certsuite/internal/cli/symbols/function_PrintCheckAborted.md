PrintCheckAborted`

| Aspect | Detail |
|--------|--------|
| **Signature** | `func(PrintCheckAborted(name string, id string) func())` |
| **Exported** | Yes – part of the public API of the `cli` package. |

### Purpose
`PrintCheckAborted` is a helper that produces a *cleanup* function used when a test check (identified by `name` and `id`) is aborted before completion.  
The cleanup routine:

1. **Stops** any ongoing goroutine that prints a progress line for the check.
2. **Outputs** an “aborted” message to the CLI using the standard `Print` helper.

This function keeps the code that aborts checks small – callers just invoke `defer PrintCheckAborted(...)()` and the side‑effects happen automatically when the deferred function runs.

### Parameters
| Parameter | Type   | Meaning |
|-----------|--------|---------|
| `name`    | `string` | The human‑readable name of the check (e.g. “TLS Certificate Check”). |
| `id`      | `string` | A unique identifier for the check instance, typically a UUID or short hash. |

### Return Value
A **zero‑argument function** (`func()`) that performs the abort handling described above.  
The returned function is meant to be used with `defer`.

```go
defer PrintCheckAborted(checkName, checkID)()
```

When the deferred function executes (e.g., due to a panic or explicit early return), it stops the progress‑line goroutine and prints an aborted status.

### Key Dependencies & Side Effects

| Dependency | Role |
|------------|------|
| `stopCheckLineGoroutine` | A package‑level helper that signals the line‑printing goroutine (started by `PrintCheckRunning`) to terminate. It uses the global `stopChan`. |
| `Print` | The generic CLI printer used for all status messages. It writes a formatted string containing the check name, ID, and an “aborted” tag (`CheckResultTagAborted`). |
| **Global state** | None beyond the two channels (`checkLoggerChan`, `stopChan`) that are already in use by other print helpers. No new goroutines or files are created. |

### Integration into the Package
The `cli` package centralises all user‑facing output for CertSuite.  
`PrintCheckAborted` sits alongside:

- `PrintCheckRunning`: starts a progress line and returns a stop function.
- `PrintCheckPass`, `PrintCheckFail`, etc.: emit final status messages.

By returning a cleanup function, it mirrors the pattern used by `PrintCheckRunning`, allowing callers to cleanly abort long‑running checks while still presenting consistent output to the user.  

The function is intentionally lightweight; its only side effects are stopping an existing goroutine and printing one line, ensuring minimal impact on performance or resource usage.
