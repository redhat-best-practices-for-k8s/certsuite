NewCustomHandler`

**Location**

`internal/log/custom_handler.go:28`

---

### Purpose
Creates a new logging handler that writes log records to the supplied writer while honoring the standard `slog.HandlerOptions`.  
The returned `*CustomHandler` implements `slog.Handler` and can be used with Go’s structured logging package (`go.uber.org/zap`, `log/slog`, etc.) or as a drop‑in replacement for the default handler in this project.

### Signature
```go
func NewCustomHandler(w io.Writer, opts *slog.HandlerOptions) *CustomHandler
```

| Parameter | Type                | Description |
|-----------|---------------------|-------------|
| `w`       | `io.Writer`         | Destination for log output (e.g., a file or stdout). |
| `opts`    | `*slog.HandlerOptions` | Optional configuration controlling level filtering, time format, etc. If `nil`, default options are used. |

### Return Value
A pointer to a new `CustomHandler`.  
The handler internally stores:

- the writer (`io.Writer`)
- the provided or default options

It does **not** close the writer; that responsibility lies with the caller.

### Key Dependencies & Relationships
| Dependency | Role |
|------------|------|
| `slog.HandlerOptions` | Configures level masking, time formatting, and any other handler‑specific settings. |
| `CustomLevelNames` (global) | Provides a mapping from custom log levels to string names used by the handler’s formatting logic. |
| `io.Writer` | The actual sink; may be an OS file, buffer, or network connection. |

### Side Effects
- **No global state mutation**: The function only creates and returns a new struct instance.
- **Does not open/close files**: It assumes the caller has already prepared the writer (e.g., opened a log file with `os.OpenFile`).
- **Thread‑safe**: The returned handler can be used concurrently as long as the underlying writer is safe for concurrent use.

### How it fits the package
The `log` package provides a minimal wrapper around Go’s standard structured logging.  
*CustomHandler* is the core component that actually formats and emits log records, while higher‑level utilities (`SetLogger`, `GetLogger`, etc.) build on top of it.  

Typical usage pattern:

```go
file, _ := os.OpenFile(LogFileName, LogFilePermissions)
handler := NewCustomHandler(file, nil)      // use default options
logger  := slog.New(handler)                // create a slog.Logger
slog.SetDefault(logger)                     // make it the global logger
```

This function is therefore the bridge between user‑supplied output destinations and the package’s structured logging API.
