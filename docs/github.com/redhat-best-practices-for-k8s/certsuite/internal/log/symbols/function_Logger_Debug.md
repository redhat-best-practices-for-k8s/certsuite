Logger.Debug`

| | |
|---|---|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/internal/log` |
| **Signature** | `func (l *Logger) Debug(msg string, args ...any)` |
| **Exported** | ✅ |

---

#### Purpose
`Debug` is a convenience method on the `Logger` type that writes a log entry at the **debug** level.  
It delegates to the more generic `Logf`, which formats the message and handles all I/O logic.

> The function exists so callers can write concise, readable code:
> ```go
> logger.Debug("processing request: %s", req.ID)
> ```

---

#### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `msg` | `string` | The message template to log. It may contain `fmt`‑style verbs (`%v`, `%d`, etc.). |
| `args ...any` | variadic | Arguments that will be substituted into the template by `fmt.Sprintf`. If omitted, `msg` is logged verbatim. |

---

#### Return value
None – the method has a `void` return type.

---

#### Key dependencies

| Dependency | Role |
|------------|------|
| `Logger.Logf` | The underlying implementation that actually formats and writes the log record. |
| Global variables | `globalLogLevel`, `globalLogger` (via the receiver’s internal state) are not directly accessed here, but they influence whether a debug message is emitted when `Logf` checks the current level. |

---

#### Side‑effects

1. **Logging** – If the logger’s level permits it (`globalLogLevel <= slog.LevelDebug`), a record is written to the configured output (file or stdout).  
2. **No state mutation** – The method does not modify any fields of `Logger`; all changes happen inside `Logf`.  

---

#### Relationship to the package

* **`log.go`** – This file defines the core logging infrastructure, including the `Logger` struct and its methods (`Debug`, `Info`, `Warn`, `Error`, `Fatal`).  
* **Custom levels** – The package supports user‑defined log levels via `CustomLevelNames` in `custom_handler.go`. `Debug` uses the standard level `slog.LevelDebug`, so it is unaffected by custom names.  
* **Global logger** – When a package‑wide logger (`globalLogger`) is initialized, its methods—including `Debug`—are used throughout the application for consistent output.

---

#### Example usage

```go
// Create or obtain a logger instance
logger := log.New("app")

// Log at debug level
logger.Debug("Cache miss: key=%s", cacheKey)

// If globalLogLevel is set to slog.LevelInfo,
// this message will be suppressed automatically.
```

---

#### Suggested Mermaid diagram (optional)

```mermaid
flowchart TD
    A[Caller] --> B{Logger.Debug}
    B --> C[Formatter: fmt.Sprintf(msg, args...)]
    C --> D[Logger.Logf(level=Debug, msgFormatted)]
    D --> E[Output stream]
```

This diagram visualises the flow from a call to `Debug` through formatting, delegation to `Logf`, and eventual output.
