SetupLogger`

```go
func SetupLogger(io.Writer, string) func()
```

## Purpose

`SetupLogger` is the entry point that configures the global logging system used throughout **certsuite**.  
It creates a new `*slog.Logger`, attaches a custom handler capable of handling user‑defined log levels, and replaces the package’s default logger (`globalLogger`).  
The returned closure restores the original logger state, enabling temporary overrides (e.g., in tests).

## Parameters

| Name | Type   | Description |
|------|--------|-------------|
| `w`  | `io.Writer` | Destination for log output. Usually `os.Stdout`, `os.Stderr`, or a file handle. |
| `levelStr` | `string` | Desired minimum log level (e.g., `"debug"`, `"error"`). Parsed by `parseLevel`. |

## Return Value

A **zero‑argument function** that, when called, restores the previous logger configuration (`globalLogger`, `globalLogFile`, `globalLogLevel`).  
Typical usage:

```go
restore := log.SetupLogger(os.Stdout, "debug")
defer restore() // reset after main or test finishes
```

## Key Dependencies

| Dependency | Role |
|------------|------|
| `parseLevel` | Converts the string level into an `slog.Level`. |
| `NewCustomHandler` | Creates a handler that understands both standard and custom log levels (`CustomLevelNames`). |
| `slog.New`, `slog.HandlerOptions` | Build the logger instance. |
| `io.Writer` (`w`) | Supplies output destination to the handler. |

## Side Effects

1. **Global state mutation**  
   - `globalLogger`, `globalLogFile`, and `globalLogLevel` are overwritten with new values.
2. **Handler creation**  
   - A new `CustomHandler` is instantiated; it may open a file if a filename is supplied via the handler options.
3. **Return closure captures old state**  
   - The returned function restores the original global logger when invoked.

No other package variables are modified, and the function itself does not write logs directly (delegating to the handler).

## How It Fits in the Package

The `log` package centralises logging configuration:

- `globalLogger` holds the active `*slog.Logger`.
- `SetupLogger` is called once during application bootstrap or test setup.
- Other parts of **certsuite** use `GetLogger()` (not shown here) to obtain the current logger.

By providing a restoration closure, the package supports flexible logging contexts without leaking state across components.
