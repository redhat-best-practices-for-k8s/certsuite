CreateGlobalLogFile`

### Purpose
`CreateGlobalLogFile` sets up a single, shared log file that will be used by the entire CertSuite test harness.  
It performs three tasks:

1. **Remove any existing file** – ensures a clean start for each run.
2. **Open (or create) the new log file** with the desired name and permissions.
3. **Configure the global logger** to write to that file.

Once called, all subsequent logging in the package will be routed through `globalLogger`, which writes to `globalLogFile`.

---

### Signature
```go
func CreateGlobalLogFile(fileName string, perm os.FileMode) error
```

| Parameter | Type          | Description                                        |
|-----------|---------------|----------------------------------------------------|
| `fileName`| `string`      | Path (relative or absolute) of the log file.       |
| `perm`    | `os.FileMode` | Unix‑style permissions for the created file.       |

| Return | Type   | Meaning                                                        |
|--------|--------|----------------------------------------------------------------|
| `error`| `error`| Non‑nil if any step fails (removal, creation, or logger setup).|

---

### Key Dependencies

| Call | What it does | Notes |
|------|--------------|-------|
| `os.Remove(fileName)` | Deletes the file if it already exists. | Used to guarantee a fresh log for each run. |
| `os.IsNotExist(err)` | Detects “file not found” errors from removal. | Allows graceful handling when nothing existed. |
| `os.OpenFile(fileName, flags, perm)` | Opens/creates the file with write permissions. | Flags: `os.O_CREATE|os.O_WRONLY|os.O_APPEND`. |
| `log.SetupLogger(globalLogLevel, globalLogFile)` | Installs the configured logger as the package’s default. | Uses `globalLogLevel` and the newly opened file. |

---

### Side‑effects & Global State

- **`globalLogFile`** is set to the opened file handle.
- **`globalLogger`** becomes a new `*slog.Logger` that writes to this file.
- The function does *not* close any previously opened log file; callers must ensure cleanup if re‑initializing.

---

### Package Integration

The `internal/log` package provides a lightweight wrapper around Go’s standard logger (`slog`).  
Typical usage pattern:

```go
// In main test harness
err := log.CreateGlobalLogFile("certsuite.log", 0644)
if err != nil { panic(err) }

// Now anywhere in the repo:
log.Info("Starting test suite")
```

`CreateGlobalLogFile` is usually invoked once during application bootstrap, before any other logging calls. It centralizes file handling and guarantees that all logs share the same level and formatting.

---

### Mermaid Diagram (Optional)

```mermaid
flowchart TD
    A[Call CreateGlobalLogFile] --> B{Remove existing file}
    B -->|Exists| C[Delete]
    B -->|Not Exists| D[Proceed]
    C & D --> E[Open new file with perms]
    E --> F[Set globalLogFile]
    F --> G[SetupLogger(globalLogLevel, globalLogFile)]
    G --> H[Logging available globally]
```

This diagram visualizes the sequence of operations and state changes performed by `CreateGlobalLogFile`.
