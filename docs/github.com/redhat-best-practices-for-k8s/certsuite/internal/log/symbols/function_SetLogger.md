SetLogger` – Package‑wide Logger Initialisation

```go
func SetLogger(*Logger)()   // located at internal/log/log.go:91
```

| Item | Description |
|------|-------------|
| **Purpose** | Provides a global, package‑level logger instance that can be reused throughout the `certsuite` code base. |
| **Parameters** | A pointer to a `Logger` struct (the concrete implementation is defined in the same file). |
| **Return value** | Returns an empty function (`func()`) that, when called, will reset the global logger back to its zero state. |

### How it works

1. **Capture the incoming logger**
   ```go
   globalLogger = l
   ```
   The supplied `*Logger` is stored in the unexported package variable `globalLogger`. This makes it accessible to all other functions in the package that rely on a shared logger.

2. **Return a cleanup closure**  
   The function returns an anonymous function that, when executed, will set `globalLogger` back to `nil`.  
   ```go
   return func() { globalLogger = nil }
   ```
   This pattern allows callers (e.g., tests) to temporarily override the logger and then restore the original state automatically.

### Dependencies & Side‑Effects

| Dependency | Role |
|------------|------|
| `globalLogger` (`*Logger`) | Holds the active logger instance. |
| `globalLogLevel`, `globalLogFile` | Not directly modified by `SetLogger`; they are part of the overall logging subsystem but remain unchanged when only the logger pointer is swapped. |

**Side‑effects**

- **Global state mutation:** The function writes to the package‑level variable `globalLogger`.  
- **Thread safety:** No synchronization primitives are used; callers must ensure that concurrent access to `SetLogger` or the global logger is safe (e.g., by restricting usage to init time or tests).

### Fit in the Package

The `log` package centralises all logging behaviour for `certsuite`.  
- `SetLogger` is the entry point for installing a custom logger implementation.  
- The returned cleanup function is especially useful in test suites where a temporary logger (e.g., capturing output to a buffer) needs to be installed and then removed cleanly.

Other package functions (such as `Info`, `Error`, etc.) internally reference `globalLogger` to emit messages. By swapping the global pointer, all parts of the application automatically use the new logger without further changes.  

### Suggested Diagram

```mermaid
flowchart TD
    A[Caller] -->|SetLogger(l)| B[log package]
    B --> C{globalLogger}
    subgraph Test Setup
        D[Temp Logger] --> E[Cleanup function]
    end
```

This diagram shows the caller installing a logger, the global pointer being updated, and the cleanup closure returned for later invocation.
