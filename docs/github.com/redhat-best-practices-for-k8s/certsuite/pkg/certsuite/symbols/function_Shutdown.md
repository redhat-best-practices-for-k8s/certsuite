Shutdown` – Package‑wide Graceful Termination

### Overview
`certsuite.Shutdown()` is a **public helper** that cleanly terminates the CertSuite process.  
It performs three actions in order:

1. **Closes the global log file** (ensuring all buffered logs are flushed).  
2. **Prints an exit message** to `stderr`.  
3. **Calls `os.Exit(0)`** to terminate the program with a success status.

This function is typically invoked from signal‑handling logic or at the end of a test run, ensuring that resources are released and the process exits deterministically.

### Signature
```go
func Shutdown() ()
```
- No parameters.  
- Returns an empty tuple (i.e., no return value).  

### Key Dependencies
| Dependency | Role |
|------------|------|
| `CloseGlobalLogFile` | Flushes and closes the file used for global logging (`globalLog`). |
| `fmt.Fprintf` | Writes a concise exit message to standard error. |
| `os.Exit` | Terminates the process with exit code `0`. |

### Side Effects
- **Resource cleanup**: The global log file is closed, preventing file descriptor leaks.
- **Process termination**: The calling goroutine exits; all other goroutines are stopped immediately by the OS.

### Package Context
`Shutdown` belongs to the top‑level `certsuite` package (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/certsuite`).  
The package orchestrates test execution, logging, and reporting. When tests finish or an unrecoverable error occurs, calling `Shutdown()` guarantees that:

- All logs are persisted.
- The user sees a clear “exiting” message.
- The process exits with a clean status code.

### Usage Pattern
```go
func main() {
    // Setup CertSuite …
    defer certsuite.Shutdown()

    // Run tests …
}
```
or in signal handling:
```go
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
go func() {
    <-sigCh
    certsuite.Shutdown()
}()
```

---

> **Note**: `Shutdown` is intentionally simple; it does not handle asynchronous cleanup beyond the global log file. Any other resources should be released before invoking this function.
