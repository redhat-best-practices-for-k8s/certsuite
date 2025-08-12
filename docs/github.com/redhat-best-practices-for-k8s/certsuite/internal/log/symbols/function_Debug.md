Debug` – Top‑level convenience wrapper

```go
func Debug(msg string, args ...any) ()
```

| Item | Description |
|------|-------------|
| **Purpose** | Emit a log record at the *debug* level for callers that do not want to manage an explicit logger instance. It is part of the package’s public API for simple “fire‑and‑forget” logging. |
| **Inputs** | `msg string` – The format string (as used by `fmt.Sprintf`).<br>`args ...any` – Optional arguments that will be formatted into `msg`. |
| **Output** | None. The function returns immediately after delegating to the underlying logger. |
| **Key dependencies** | Calls the package‑internal `Logf` helper, which in turn writes to the global `globalLogger` (a `*Logger`) and respects the current `globalLogLevel`. No external packages are imported directly by this wrapper; it relies on the global state set up elsewhere in the package. |
| **Side effects** | *Writes a log entry* at level `LevelDebug` if the configured global level permits it.<br>*Does not modify any global variables.*<br>It uses the shared logger instance, so concurrency safety depends on the underlying logger implementation (the package ensures this with appropriate mutexes). |
| **Package context** | The `log` package provides a lightweight wrapper around Go’s standard `slog`.  It exposes five top‑level helpers (`Debug`, `Info`, `Warn`, `Error`, `Fatal`) that mirror common log levels. These functions are meant for quick, one‑liner logging where the caller does not need to construct a logger or handle context. The package also offers more granular control via the exported `Logger` type and configuration functions. |

---

### Usage example

```go
log.Debug("Processing %d items", len(items))
```

If the global log level is set to `LevelDebug` (or lower), this will produce a debug‑level entry in the configured output (console or file). If the level is higher, the call becomes a no‑op.

---

### Suggested Mermaid diagram

```mermaid
flowchart TD
    A[Caller] -->|Debug(msg,args)| B(log.Debug)
    B --> C(Logf(level=LevelDebug,msg,args))
    C --> D(globalLogger.Write)
```

This diagram illustrates that `Debug` is a thin wrapper that forwards the call to `Logf`, which writes the record using the shared global logger.
