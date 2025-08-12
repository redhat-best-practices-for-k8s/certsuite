GetLogger`

| Feature | Detail |
|---------|--------|
| **Signature** | `func() *Logger` |
| **Exported?** | Yes (`GetLogger`) |
| **Location** | `internal/log/log.go:95` |

### Purpose
`GetLogger` is the public accessor for the package‑wide logger instance.  
It guarantees that callers receive a non‑nil `*Logger` configured with the current global log level and output destination.

> **Why a function?**  
> The logger is stored in the unexported variable `globalLogger`. Exposing it directly would break encapsulation (e.g., tests could modify it). `GetLogger` provides read‑only access while allowing lazy initialization if needed.

### Inputs / Outputs
- **Inputs:** none.  
- **Outputs:** a pointer to the singleton logger (`*Logger`). The returned value is never `nil`.

### Key Dependencies & Side Effects
| Dependency | Role |
|------------|------|
| `globalLogger` | Holds the actual logger instance. `GetLogger` reads this variable. |
| `globalLogLevel` | Determines which log levels are enabled; used when the logger is created. |
| `globalLogFile` | File descriptor where logs are written; used to construct the logger’s output stream. |

**Side effects:**  
None – the function only returns a reference, it does not modify any state.

### How It Fits Into the Package

```mermaid
flowchart TD
    A[Caller] -->|GetLogger()| B[log package]
    B --> C{globalLogger}
    C --> D[*Logger]
```

* `log` is an internal package providing a lightweight wrapper around Go’s standard `slog`.
* The singleton pattern ensures that all components use the same logger configuration.
* `GetLogger` is the single entry point for other packages (e.g., `certsuite`) to obtain the configured logger.

### Typical Usage

```go
import "github.com/redhat-best-practices-for-k8s/certsuite/internal/log"

func main() {
    logger := log.GetLogger()
    logger.Info("Application started")
}
```

The returned `*Logger` can be used directly or wrapped in a higher‑level interface if desired.

---

**Summary:**  
`GetLogger` provides safe, read‑only access to the global logger instance. It relies on three internal globals (`globalLogger`, `globalLogLevel`, `globalLogFile`) and has no side effects, making it ideal for shared logging across the application.
