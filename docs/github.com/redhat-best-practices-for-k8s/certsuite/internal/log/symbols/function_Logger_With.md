Logger.With`

```go
func (l *Logger) With(args ...any) *Logger
```

### Purpose  
Creates a new `*Logger` instance that inherits the current logger’s configuration but augments it with additional key/value pairs supplied via `args`.  
The resulting logger can be used to emit log records that automatically include those extra fields, making it easier to add context (e.g., request IDs, user identifiers) without affecting other loggers.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `args` | `...any` | A variadic list of alternating keys and values. Keys must be string‑compatible; the function passes them directly to the underlying `slog.With`. If an odd number of arguments is supplied, the last one is ignored.

### Return value
| Type | Description |
|------|-------------|
| `*Logger` | A new logger that shares all handlers and level settings with the receiver but has its own set of fields derived from `args`.

### Key Dependencies
- **slog** – The method delegates to `l.slog.With(args...)`, where `l.slog` is an instance of Go’s standard `slog.Logger`.  
- **CustomHandler / CustomLevelNames** – If the global logger uses a custom handler, the new logger will also use that same handler. The function itself does not directly reference these globals.

### Side‑effects
- No modification occurs to the receiver (`l`).  
- Only creates a new `*Logger` value; no I/O or state changes beyond the internal slog logger.

### Package Context
The `log` package provides a thin wrapper around Go’s standard logging facilities, exposing convenient constructors and helpers. `Logger.With` is part of that wrapper’s API, allowing callers to create context‑rich loggers without repeatedly passing fields through every log call. It complements other helpers like `New`, `SetLevel`, and the global logger (`globalLogger`).
