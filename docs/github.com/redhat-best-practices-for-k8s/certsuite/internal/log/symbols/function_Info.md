Info` – a convenience wrapper around the package logger

### Purpose
`Info` is a small helper that logs an informational message at **log level = `LevelInfo`**.  
It forwards the call to the internal `Logf` routine, so it inherits all of its behaviour (formatting, context handling, file‑output, etc.).  The function exists mainly for ergonomics: callers can simply write

```go
log.Info("starting server on %s", addr)
```

instead of constructing a full log record manually.

### Signature
```go
func Info(msg string, args ...any) ()
```
* `msg` – the format string (passed verbatim to `fmt.Sprintf`).  
* `args` – optional arguments for formatting.  
The function returns nothing; it performs its work via side‑effects on the global logger.

### Key Dependencies
| Dependency | Role |
|------------|------|
| **`Logf`** | The actual logging implementation that formats the message, attaches a level (`LevelInfo`) and writes to the configured log destination. |
| **`globalLogger` / `globalLogLevel`** | Stored globally in the package; `Logf` uses them to determine whether the record should be emitted and where it goes. |

### Side Effects
* Calls `Logf`, which may:
  * Write a formatted string to the log file (if one is configured via `globalLogFile`).
  * Emit to stdout/stderr if no file is set.
  * Add contextual fields such as timestamp, level name, and source location.

The function itself does not modify any global state; it only triggers logging through `Logf`.

### Package Context
The `log` package centralises all application‑wide logging.  
* Global variables (`globalLogger`, `globalLogLevel`, `globalLogFile`) hold the logger instance, current level threshold, and optional file handle.  
* Level constants (`LevelDebug`, `LevelInfo`, …) define severity tiers.  
* Helper functions like `Fatal`, `Warn`, `Error` mirror `Info` but use different levels.

`Info` therefore sits at the core of the public API: it is one of the most frequently used entry points for emitting logs across the project, keeping code concise and consistent while delegating actual formatting and I/O to the shared logger implementation.
