Logger.Error`

| Aspect | Details |
|--------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/internal/log` |
| **Signature** | `func (l *Logger) Error(msg string, args ...any)` |
| **Exported?** | Yes |

### Purpose
Convenience wrapper that logs a message at the **Error** level.  
It delegates to the underlying `Logf` method, automatically applying the `"error"` log level.

### Parameters
- `msg string` – The format string or plain text of the log entry.
- `args ...any` – Optional arguments that are passed through to formatting (e.g., `%s`, `%d`).

The function does **not** return a value; it simply performs logging side‑effects.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `l.Logf(level, msg, args...)` | The actual log output routine.  It writes to the global logger (`globalLogger`) which is configured with file and console handlers. |
| `LevelError` (constant) | Represents the slog level value for errors; passed to `Logf`. |

### Side‑Effects
- Emits a formatted log entry at **error** severity.
- If the underlying handler is a file, the message is appended to the log file (`globalLogFile`) with appropriate permissions (`LogFilePermissions`).
- No other state is mutated.

### How it Fits in the Package
The `log` package exposes several typed logging helpers (`Debug`, `Info`, `Warn`, `Error`, `Fatal`).  
`Logger.Error` is one of those helpers and follows the same pattern:

```go
func (l *Logger) Debug(msg string, args ...any) { l.Logf(LevelDebug, msg, args...) }
```

These helpers simplify calling code by hiding the explicit level argument.  
They also make it easy to change the underlying logging implementation without touching callers.

### Usage Example

```go
log := log.New()
log.Error("failed to load config: %v", err)
```

This produces a line similar to:

```
2025-08-11T12:34:56Z ERROR  failed to load config: <error details>
```

---

**Note:**  
`Logger.Error` is read‑only; it does not alter any global variables or logger configuration. It simply forwards the call to `Logf`.
