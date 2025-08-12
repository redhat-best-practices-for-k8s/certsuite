Logf` – Central Logging Helper

`Logf` is the core formatting routine that all public log wrappers (e.g., `Info`, `Error`, `Fatal`) delegate to.  
It centralises:

* **Level handling** – translates string names into `slog.Level`.
* **Caller resolution** – captures the correct source line for each log call.
* **Record creation & emission** – builds a `Record` and hands it off to the active `Handler`.

---

## Signature

```go
func Logf(logger *Logger, levelName string, msgFmt string, args ...any) ()
```

| Parameter | Type     | Description |
|-----------|----------|-------------|
| `logger`  | `*Logger`| The logger instance (usually the global one). |
| `levelName` | `string` | Human‑readable level (`"debug"`, `"info"`, `"warn"`, `"error"`, `"fatal"` or a custom name). |
| `msgFmt` | `string` | Go format string for the log message. |
| `args...` | `…any`   | Values to interpolate into `msgFmt`. |

The function returns nothing; all side effects are performed via the handler.

---

## How it Works

1. **Resolve level**  
   ```go
   lvl, err := parseLevel(levelName)
   ```
   *`parseLevel`* maps the string to a `slog.Level`. If parsing fails, `Fatal` is called with the error message (ensuring the application stops on malformed levels).

2. **Check if logging is enabled**  
   ```go
   if !Enabled(logger, lvl) { return }
   ```
   Skips formatting and emission when the global level filter would suppress this record.

3. **Build the log record**  
   * Timestamp: `Now()` (current time).  
   * Caller information: `Callers(2)` – skips the wrapper (`Logf` itself) to point at the user code.  
   * Message: `Sprintf(msgFmt, args...)`.  

4. **Emit**  
   ```go
   Handle(logger, NewRecord(...))
   ```
   The active handler (retrieved via `Handler(logger)`) writes the record to its destination.

5. **Fatal handling** – if the level is `"fatal"`, `Fatal` is invoked after emitting so that the program exits.

---

## Dependencies & Side‑Effects

| Called Function | Purpose |
|-----------------|---------|
| `Default()` | Provides a fallback logger when none is supplied (unused directly in this function). |
| `parseLevel()` | Converts string level to `slog.Level`. |
| `Fatal()` | Terminates the process for fatal logs. |
| `Enabled()` | Checks if a level should be logged given current global settings. |
| `TODO()` | Placeholder used when a log wrapper is missing; not part of normal flow. |
| `Callers(2)` | Captures file/line info two stack frames up (the caller of the wrapper). |
| `NewRecord()` | Creates a `Record` struct with timestamp, level, message, and call stack. |
| `Now()` | Current time in UTC. |
| `Sprintf()` | Formats the user‑supplied message. |
| `Handle()` | Dispatches the record to the logger’s handler. |
| `Handler()` | Retrieves the current handler for a logger. |

**Side‑effects**

* Writes log entries via the active handler (usually stdout or a file).  
* On fatal level, calls `os.Exit(1)` through `Fatal`.  
* Uses global variables (`globalLogLevel`, `globalLogger`) indirectly via helper functions.

---

## Package Context

The `log` package provides a lightweight wrapper around Go’s standard `slog` with:

* Custom log levels (e.g., `CustomLevelFatal`).  
* Global configuration for level, file output, and permissions.  
* Convenience wrappers (`Debug`, `Info`, `Warn`, `Error`, `Fatal`) that internally call `Logf`.

`Logf` is the single point where formatting logic lives, ensuring consistent metadata (timestamp, caller) across all log levels.

---

### Mermaid Diagram (suggested)

```mermaid
flowchart TD
    A[Caller] -->|wrapper| B[Log Wrapper]
    B --> C[Logf]
    C --> D[parseLevel]
    D --> E{valid?}
    E -- no --> F[Fatal(err)]
    E -- yes --> G[Enabled]
    G -- false --> H[return]
    G -- true --> I[NewRecord]
    I --> J[Handle]
    J --> K[Handler -> output]
```

This visualizes the control flow from user code to the final log emission.
