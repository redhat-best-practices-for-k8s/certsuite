MultiHandler.Enabled`

```go
func (mh *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool
```

### Purpose  
`Enabled` is the implementation of the [`slog.Handler`](https://pkg.go.dev/go.opentelemetry.io/otel/sdk/log#Handler) interface that determines whether a given log level should be processed by any of the underlying handlers stored in `MultiHandler`.  
It allows the logger to short‑circuit formatting and emission for levels that are not enabled, improving performance.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `ctx` | `context.Context` | Unused; present only because the interface requires it. The context can be used by future handlers to carry request‑scoped data. |
| `level` | `slog.Level` | Log level of the message being considered (e.g., `LevelDebug`, `CustomLevelFatal`). |

### Return value
* **`bool`** –  
  * `true` if at least one embedded handler reports that it is enabled for the supplied level.  
  * `false` otherwise.

### Key Dependencies & Side‑Effects
| Dependency | How it’s used |
|------------|---------------|
| `mh.handlers []slog.Handler` (implicit in the receiver) | Iterates over all handlers and calls each handler’s own `Enabled(ctx, level)` method. |
| `CustomLevelNames`, `globalLogFile`, `globalLogLevel`, `globalLogger` | None – this function does **not** read or modify any package‑level globals; it purely delegates to the contained handlers. |

### How It Fits in the Package
* The *log* package implements a flexible logging system based on Go’s standard `slog`.  
* `MultiHandler` aggregates multiple concrete handlers (e.g., file, console, custom) so that logs can be emitted simultaneously to several destinations.  
* `Enabled` is called by the top‑level logger before constructing or emitting any log record; it prevents unnecessary work when a message’s level would be discarded by all underlying handlers.

### Usage Example

```go
mh := &log.MultiHandler{
    handlers: []slog.Handler{
        console.NewConsoleHandler(os.Stdout, nil),
        file.NewFileHandler("app.log", nil),
    },
}

if mh.Enabled(context.Background(), slog.LevelInfo) {
    // Build and emit the log record
}
```

In this example, `Enabled` returns `true` if either the console or file handler accepts `LevelInfo`.  
If all handlers were disabled for that level, the logger would skip constructing the record entirely.
