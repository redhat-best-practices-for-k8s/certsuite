PrintCheckFailed`

> **Signature**  
> ```go
> func PrintCheckFailed(msg string) func()
> ```

`PrintCheckFailed` is a helper that signals the CLI that a check has failed and returns a cleanup function that must be executed when the caller finishes handling the failure.  

---

## Purpose

When a test or health‑check fails, the CLI prints a concise “failed” message and stops any ongoing progress indicator (the *check line goroutine*).  
The returned closure performs two important side effects:

1. **Stops the check line goroutine** – it signals `stopChan` to terminate the spinner that shows progress while a check is running.
2. **Logs the failure** – it prints the supplied message (`msg`) in red, prefixed by the “failed” tag.

This design keeps the caller free of any direct dependency on the internal logging/printing logic and guarantees that the goroutine is always stopped, even if the caller panics or returns early.

---

## Inputs

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `msg`     | `string` | Human‑readable message describing why the check failed. It will be printed after the failure tag. |

No other inputs are required; all state is captured from package globals (`stopChan`, `CliCheckLogSniffer`, etc.).

---

## Outputs

The function returns a **zero‑argument closure**:

```go
func()
```

When invoked, this closure performs the cleanup described above.  
It does not return any value or error.

---

## Key Dependencies & Side Effects

| Dependency | Role |
|------------|------|
| `stopChan` (`chan bool`) | Used to signal the check line goroutine to stop. The goroutine reads from this channel and exits when it receives a value. |
| `Print` (internal helper) | Formats and writes the failure message to stdout/stderr. It uses ANSI colour codes defined in the package constants (`Red`, `Reset`). |
| `CliCheckLogSniffer` | Not directly used by `PrintCheckFailed`; however, its presence indicates that logging is centralised through a sniffer goroutine that consumes messages from `checkLoggerChan`. The closure may indirectly trigger this logger if `Print` writes to the channel. |

Side effects:

- **Console output** – prints “FAILED” in red followed by the supplied message.
- **Goroutine termination** – ensures the progress spinner stops, preventing a stale UI.

---

## How It Fits Into the Package

The `cli` package implements an interactive command‑line interface for CertSuite.  
Checks run asynchronously and display a live progress indicator via a dedicated goroutine. When a check fails, it is essential to:

1. Stop the spinner immediately so that the user sees the failure without delay.
2. Notify the rest of the system (e.g., logging, metrics) about the failure.

`PrintCheckFailed` encapsulates these two responsibilities and returns a cleanup function that callers can defer or invoke explicitly.  
Typical usage pattern:

```go
defer PrintCheckFailed("certificate expired")()
// ... perform check logic
```

This ensures the spinner is stopped regardless of how the function exits, while keeping the failure message visible to the user.

---
