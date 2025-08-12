Logger.Fatal` – package‑level crash logger

### Purpose
`Fatal` logs a message at the *fatal* severity and immediately terminates the process.  
It is intended for unrecoverable errors where continuing execution would be unsafe or meaningless.

### Signature
```go
func (l Logger) Fatal(msg string, args ...any)
```
| Parameter | Type   | Description |
|-----------|--------|-------------|
| `msg`     | `string` | The message format string. |
| `args...` | `…any`  | Optional values to interpolate into the format string (passed to `fmt.Fprintf`). |

### How it works
1. **Log the message**  
   Calls `l.Logf(LevelFatal, msg, args...)`.  
   - `Logf` writes a formatted log line using the logger’s underlying handler.  
   - The log entry is tagged with the *fatal* level (`LevelFatal`) which can be filtered or processed specially by the handler.

2. **Print to standard error**  
   Executes `fmt.Fprintf(os.Stderr, msg+"\n", args...)`.  
   This guarantees that the fatal message appears in the process’s stderr even if the logger is misconfigured (e.g., log file not opened).

3. **Exit the program**  
   Calls `os.Exit(1)` via the package‑level helper `Exit`, ensuring an immediate, non‑graceful shutdown.

### Dependencies & Side Effects
| Dependency | Role |
|------------|------|
| `Logf`     | Formats and records the log line with severity. |
| `fmt.Fprintf` | Prints to stderr as a safety net. |
| `os.Exit(1)` (via `Exit`) | Terminates the process; no deferred functions run. |

Because `Fatal` exits, any code after its call will not execute. This is by design: fatal errors are considered unrecoverable.

### Package Context
The `log` package provides a lightweight structured logger built on Go’s standard `slog`.  
- `Logger` wraps an `slog.Logger`, exposing convenience methods (`Debug`, `Info`, `Warn`, `Error`, and `Fatal`).  
- Global variables (`globalLogLevel`, `globalLogFile`, `globalLogger`) hold the package‑wide configuration.  
- `CustomLevelNames` allows user‑defined log levels.

`Fatal` sits at the top of the severity hierarchy. It is the most drastic action a logger can take and should be used sparingly, typically for bugs or critical failures that cannot be handled gracefully.
