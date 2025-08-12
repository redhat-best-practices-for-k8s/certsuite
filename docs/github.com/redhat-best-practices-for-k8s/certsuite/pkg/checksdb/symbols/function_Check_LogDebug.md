Check.LogDebug`

```go
func (c Check) LogDebug(format string, args ...any) func()
```

### Purpose  
`LogDebug` is a convenience wrapper that logs a debug‑level message associated with a specific check.  
It delegates to the generic `Logf` helper and returns an empty function so it can be used in
deferred calls or as a no‑op placeholder when debugging is disabled.

> **Why return `func()`?**  
> The package often uses patterns such as:
> ```go
> defer c.LogDebug("cleaning up…")()
> ```
> which guarantees the message is emitted when the surrounding function returns, without
> affecting normal execution flow. Returning a zero‑value function also allows callers to
> ignore the result if they don't need deferred logging.

### Parameters  
| Name | Type | Description |
|------|------|-------------|
| `format` | `string` | Go `fmt.Sprintf` style format string. |
| `args` | `...any` | Values to interpolate into `format`. |

### Return value  
A zero‑value function of type `func()`.  
When invoked, it performs the debug log; otherwise it does nothing.

### Key dependencies  
* **`Check.Logf`** – the underlying logger method that accepts a log level and a format string.  
  It is responsible for formatting the message, applying any check‑specific context (e.g., ID,
  group), and writing to the configured logger.
* The method uses no other global state or package variables directly.

### Side effects  
* Emits a debug message via the check’s logger when the returned function is called.  
* No modification of package‑level globals, the `Check` instance, or other checks.

### How it fits in the package  

The `checksdb` package models a collection of testable checks (`Check`). Each check can log
messages at different severity levels: *INFO*, *WARN*, *ERROR*, and *DEBUG*.  
While high‑level logs are common, debug logs are typically sprinkled throughout the code to
trace execution paths or variable values. `LogDebug` centralises this pattern:

* Keeps debug‑log statements succinct (`c.LogDebug("msg %d", i)()`).
* Allows callers to easily toggle debug output by modifying the underlying logger’s level.
* Maintains consistency with other log helpers (`LogInfo`, `LogWarn`, etc.) that are part of
  the same `Check` interface.

```mermaid
flowchart TD
    A[Call site] --> B{c.LogDebug(...)}
    B -->|returns| C[func()]
    C --> D[Execute deferred/log]
    D --> E[Check.Logf(DEBUG, …)]
```

---

**Note:** The actual implementation of `LogDebug` is trivial – it simply forwards to
`Check.Logf` with the `DEBUG` level and returns a no‑op function.  All behaviour is therefore
fully determined by `Check.Logf`.
