CloseGlobalLogFile`

**Location**

`internal/log/log.go:56`

---

### Purpose
Gracefully shuts down the package‑wide logger by closing the underlying log file that has been opened for appending log entries.

When the application terminates or when a manual cleanup is required, this function ensures that any buffered data is flushed and the OS resource (`*os.File`) is released. Failing to close the file can lead to:

- Data loss (unflushed writes)
- File descriptor leaks
- Inability to rename/rotate the log file later

### Signature

```go
func CloseGlobalLogFile() error
```

| Parameter | Type | Description |
|-----------|------|-------------|
| _        | –    | None |

| Return | Type   | Description |
|--------|--------|-------------|
| err    | `error`| Non‑nil if the underlying file close operation fails. `nil` indicates success. |

### Dependencies & Side Effects

- **Global state**: Uses the unexported variable `globalLogFile *os.File`.  
  - If `globalLogFile` is `nil`, the function simply returns `nil` (no-op).  
  - Otherwise it calls `globalLogFile.Close()`.

- **External call**: Relies on `(*os.File).Close()`. No other side effects beyond closing the file descriptor.

### How It Fits the Package

The package maintains a single global logger instance (`globalLogger`) that writes to a persistent log file. The lifecycle of this file is:

1. **Initialization** – Created during `InitGlobalLogFile()` (not shown here) and stored in `globalLogFile`.  
2. **Logging** – All log calls funnel through the global logger, which writes directly to `globalLogFile`.  
3. **Shutdown** – `CloseGlobalLogFile()` is called at program exit or when a clean shutdown is desired.

This function complements other package utilities such as:

- `InitGlobalLogFile` – opens/creates the file.
- `SetGlobalLogLevel` – adjusts verbosity.
- `NewCustomHandler` – custom log level handling.

A typical usage pattern in an application might look like:

```go
func main() {
    // Setup logging
    if err := log.InitGlobalLogFile("app.log"); err != nil { ... }

    defer func() {
        if err := log.CloseGlobalLogFile(); err != nil {
            fmt.Println("failed to close log file:", err)
        }
    }()

    // Application logic …
}
```

### Summary

`CloseGlobalLogFile` is a small, but essential helper that finalizes the global logging subsystem by closing its backing file. It respects the current state (no-op if already closed) and propagates any error from `os.File.Close()` to the caller for proper handling.
