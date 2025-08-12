GetMultiLogger`

```go
func GetMultiLogger(...io.Writer) *Logger
```

Creates a **multi‑writer logger** that writes log records to several destinations in parallel.

---

### Purpose
`GetMultiLogger` is the package’s factory for a logger that forwards each emitted record to multiple underlying writers (e.g., `os.Stdout`, a file, or any `io.Writer`).  
It is used when an application needs to keep logs both on‑screen and persisted, or send them to different backends.

---

### Parameters
| Parameter | Type      | Description |
|-----------|-----------|-------------|
| `writers ...io.Writer` | variadic slice of `io.Writer` | The destinations where log records will be written. At least one writer is required; passing zero writers returns a logger that discards all output (via an empty handler). |

---

### Return value
| Value | Type   | Description |
|-------|--------|-------------|
| `*Logger` | A pointer to the newly created `Logger` instance | The logger uses a *multi‑handler* (`NewMultiHandler`) internally. All log records are forwarded to each supplied writer. |

---

### Implementation Flow

1. **Wrap each writer**  
   ```go
   for _, w := range writers {
       h = append(h, NewCustomHandler(w))
   }
   ```
   Each `io.Writer` is wrapped in a custom handler that respects the package’s log‑level semantics (`NewCustomHandler`).  

2. **Create a multi‑handler**  
   ```go
   mh := NewMultiHandler(h...)
   ```
   The handlers are combined into one `slog.Handler` that dispatches to all of them.

3. **Instantiate the logger**  
   ```go
   return New(mh)
   ```
   Calls `New`, which constructs a `Logger` with the multi‑handler and the current global log level (`globalLogLevel`).  

---

### Key Dependencies

| Dependency | Role |
|------------|------|
| `slog.New` (via `New`) | Creates a `*Logger` from a handler. |
| `NewCustomHandler` | Wraps an `io.Writer` into a slog-compatible handler that applies level filtering. |
| `NewMultiHandler` | Combines several handlers into one that broadcasts messages. |
| `globalLogLevel` | Supplies the logger’s default log level at construction time. |

---

### Side Effects & Global State

- **No modification of globals** – The function only reads `globalLogLevel`; it does not alter any package‑level variables.
- **Non‑destructive** – Existing loggers and handlers remain untouched.

---

### Usage Context in the Package

`GetMultiLogger` is a convenience wrapper for scenarios where logs need to be emitted simultaneously to multiple sinks.  
Typical usage:

```go
// Log both to stdout and a file
file, _ := os.Create("app.log")
logger := log.GetMultiLogger(os.Stdout, file)
logger.Info("Application started")
```

This function complements the package’s other helpers (`GetStdOutLogger`, `GetFileLogger`) by allowing arbitrary combinations of destinations.
