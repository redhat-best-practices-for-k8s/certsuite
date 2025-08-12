Check.Abort`

```go
func (c *Check) Abort(msg string) func()
```

`Abort` is a helper that turns a fatal failure of a check into an immediate panic while preserving the check’s mutex state and providing a consistent error message.

### Purpose
* **Terminate** execution of a check as soon as an unrecoverable condition is detected.
* Ensure that the `Check` instance remains in a safe, locked state for the caller to inspect results or cleanup before the panic unwinds the stack.
* Provide a clear, user‑friendly panic message that can be caught by higher‑level error handling logic.

### Inputs
| Parameter | Type   | Description |
|-----------|--------|-------------|
| `msg`     | `string` | A human‑readable description of why the check is aborted. |

### Output
* Returns a **zero‑argument function** (`func()`) that must be called by the caller immediately after calling `Abort`.  
  The returned closure performs:
  1. Unlocks the `Check`’s internal mutex (`c.Unlock()`).
  2. Panics with a message constructed by `AbortPanicMsg(msg)`.

The caller typically uses it in a deferred statement:

```go
defer check.Abort("unexpected nil cert")()
```

### Key Dependencies & Side Effects
| Dependency | Role |
|------------|------|
| `c.Lock()` / `c.Unlock()` | Ensures the check’s state is safely updated before panic. |
| `panic` | Triggers Go’s panic mechanism to unwind the stack. |
| `AbortPanicMsg(msg)` | Formats a standard error message for aborted checks (implementation not shown here). |

Because the returned function unlocks and panics, **no further code in the current goroutine will execute** after it is called.

### How It Fits the Package

`checksdb` manages a collection of `Check`s that are executed against Kubernetes resources.  
When a check encounters an unrecoverable error (e.g., missing required data), it calls `Abort`.  
The panic propagates up to the test harness or command‑line tool, which can catch it and report the failure in a consistent way.

A typical flow:

```go
func (c *Check) Run() {
    c.Lock()
    defer c.Unlock()

    // ...perform work...

    if err != nil {
        // Abort early; caller will handle cleanup.
        defer c.Abort(fmt.Sprintf("failed to process %s", resource))()
    }

    // continue with normal execution
}
```

By centralising abort logic, the package guarantees that all checks clean up their mutexes and produce uniform error messages.
