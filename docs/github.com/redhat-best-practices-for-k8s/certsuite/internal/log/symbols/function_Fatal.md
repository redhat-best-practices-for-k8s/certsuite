Fatal` – Terminating Logger

### Purpose
`Fatal` logs a formatted message at the **fatal** level and then immediately terminates the process with an exit status of `1`.  
It is intended to be used when a non‑recoverable error occurs (e.g., configuration failure, missing file) that should stop program execution.

> **Note:** The function name follows Go’s convention for fatal logging functions (`log.Fatalf` in the standard library). It behaves similarly but uses the package’s custom logger and log levels.

### Signature
```go
func Fatal(msg string, args ...any)
```

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `msg`     | `string` | Format string (passed to `fmt.Sprintf`). |
| `args`    | `...any`  | Values substituted into the format string. |

The function does **not** return a value; it calls `os.Exit(1)` after logging.

### Key Dependencies
| Called Function | Purpose |
|-----------------|---------|
| `Logf(level, msg, args...)` | Emits the message using the global logger at level `LevelFatal`. |
| `Fprintf(file, format, args...)` | When a log file is configured (`globalLogFile != nil`) this writes the same fatal message directly to that file. |
| `Exit(code)`  | Calls `os.Exit(1)`, halting the process immediately. |

The function relies on the following package‑wide globals:

- **`globalLogger`** – The singleton logger instance used for all output.
- **`globalLogFile`** – If set, fatal messages are duplicated to this file before exit.

### Side Effects
1. **Logging** – Emits a fatal message to both stdout/stderr (via `globalLogger`) and optionally to the configured log file.
2. **Process Termination** – Calls `os.Exit(1)`, which stops all goroutines and exits the program without running deferred functions in other packages.

### Interaction with Other Package Elements
- **Log Levels** – Uses the constant `LevelFatal` defined in `log.go`. The fatal level is considered higher than any other custom or standard level, ensuring that only fatal messages are logged when the global log level is set lower.
- **Custom Handler** – If a custom handler has been installed via `custom_handler.go`, it may intercept the fatal message before termination.
- **Global State** – Since `Fatal` operates on package globals (`globalLogger`, `globalLogFile`), it must be called after those have been initialized (e.g., in `Init()` or during main startup).

### Usage Example
```go
func main() {
    // Initialize logging (sets globalLogger, globalLogLevel, etc.)
    log.Init()

    // Some critical error occurs
    err := startServer()
    if err != nil {
        log.Fatal("Failed to start server: %v", err)
    }
}
```

In the above example, if `startServer()` returns an error, the program logs a fatal message and exits immediately.

### Summary
`Fatal` is a convenience wrapper that guarantees a critical error is logged before the process stops. It combines level‑specific logging (`Logf(LevelFatal)`) with optional file output and a hard exit, mirroring the behaviour of `log.Fatalf` from Go’s standard library but within this package’s custom logging framework.
