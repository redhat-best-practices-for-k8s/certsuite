PrintCheckErrored` – CLI helper for test‑run error handling  

**File:** `internal/cli/cli.go` (line 213)  
**Package:** `cli`

---

## Purpose
`PrintCheckErrored` is a *factory* that returns a zero‑argument function.  
The returned closure is intended to be called when a test/check has failed or
encountered an error. It:

1. Stops the live‑update goroutine that prints a progress line (`stopCheckLineGoroutine`).  
2. Prints a formatted message indicating the check failed.

This pattern lets callers register a single function that will run at the end of a
check without needing to pass any context or state into it.

---

## Signature

```go
func PrintCheckErrored(check string) func()
```

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `check`   | `string` | The name of the check that errored. |

| Return value | Type   | Description |
|--------------|--------|-------------|
| `func()`     | closure | Function that performs the cleanup and prints the error message. |

---

## Key Dependencies

| Dependency | Role |
|------------|------|
| `stopCheckLineGoroutine` | Signals the goroutine printing the live check line to stop (`stopChan`). |
| `Print` | Helper that writes a colored line to stdout; used to display the error message. |

No other global state is accessed directly inside this function.

---

## Side Effects

* **Channel communication** – Sends a value on `stopChan`, causing the goroutine
  spawned by `CliCheckLogSniffer` to terminate.
* **Console output** – Calls `Print` once, producing a single line that looks like:

```
[❌] Check <check> failed
```

(The actual color and emoji come from constants defined earlier in the file.)

---

## How It Fits Into the Package

The `cli` package provides terminal UI helpers for CertSuite.  
During a test run:

1. A goroutine (`CliCheckLogSniffer`) prints a live status line for the current
   check, updating every second (controlled by `tickerPeriodSeconds`).  
2. When a check finishes, one of several helper functions is called:
   * `PrintCheckPassed`
   * `PrintCheckSkipped`
   * `PrintCheckErrored` (this function)
3. Each helper stops the live‑update goroutine and prints a final status line,
   ensuring that only one status line remains visible at any time.

Thus, `PrintCheckErrored` is part of the “finalize‑check” workflow that keeps
the CLI output clean and informative.

---

## Example Usage

```go
// Inside a check runner:
if err != nil {
    // Register error handling once; it will be invoked later.
    defer PrintCheckErrored(checkName)()
}
```

When `err` is non‑nil, the deferred call stops the progress line and prints
the failure message.

---

## Summary

- **What** – Returns a cleanup/print closure for failed checks.  
- **Why** – Keeps CLI output tidy by stopping live updates and showing an error line.  
- **How** – Uses `stopChan` to stop the status‑line goroutine and `Print` to emit the message.  
- **Where** – Part of the terminal UI helpers in `internal/cli`.
