Warn` – Convenience wrapper for warning‑level logging

| Item | Details |
|------|---------|
| **Signature** | `func Warn(msg string, args ...any)` |
| **Exported?** | Yes (public API) |
| **File / Line** | `internal/log/log.go:138` |

### Purpose
`Warn` is a thin convenience wrapper that logs a formatted warning message at the *warn* level.  
It forwards its arguments to the package‑wide `Logf` function, which handles formatting, timestamping and routing to the configured logger.

### Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `msg` | `string` | A printf‑style format string describing the warning. |
| `args ...any` | variadic | Optional values that will be substituted into `msg`. If omitted, `msg` is logged verbatim. |

### Return value
None – it performs side effects only.

### Key dependencies
* **`Logf`** – The core logging routine that actually writes to the global logger (`globalLogger`).  
  `Warn` simply calls `Logf(LevelWarn, msg, args...)`.
* **`LevelWarn`** – The slog level constant representing warnings. This is defined in the same package (`log.go`) and used by `Logf` to decide whether a message should be emitted based on the current global log level.

### Side effects
1. Invokes `Logf`, which may:
   * Write the formatted message to the underlying file (`globalLogFile`) if a log file is configured.
   * Emit the record to any attached handlers (e.g., console, external sinks).
2. No state mutation occurs in `Warn` itself; it relies on global logger configuration set elsewhere.

### How it fits into the package
The `log` package exposes several level‑specific helpers (`Debug`, `Info`, `Warn`, `Error`, `Fatal`) for ergonomic use throughout the codebase.  
Each helper simply forwards to `Logf` with its corresponding level constant, ensuring consistent formatting and output handling while keeping the API concise.

```mermaid
graph LR
    A[Call site] -->|Warn(msg, args)| B(Warning)
    B --> C{Global Logger}
    C --> D[Output Destinations]
```

### Usage example
```go
import "github.com/redhat-best-practices-for-k8s/certsuite/internal/log"

func checkConfig() {
    if err := load(); err != nil {
        log.Warn("config file %s could not be loaded: %v", configPath, err)
    }
}
```

`Warn` is therefore the preferred way to emit warning messages from any package that imports this internal logging subsystem.
