cliCheckLogSniffer.Write([]byte) (int, error)`

### Purpose
Implements the **`io.Writer`** interface for the custom slog handler that powers CertSuite’s command‑line check output.  
The method receives log data from the slog logger, formats it if necessary, and forwards a human‑readable string to an internal channel (`checkLoggerChan`). That channel is consumed by the CLI ticker goroutine, which periodically updates the terminal with a status line for each running check.

### Inputs
| Parameter | Type   | Description |
|-----------|--------|-------------|
| `p`       | `[]byte` | Raw log payload emitted by slog. |

The method does **not** expose any other arguments; all state is held in the receiver struct (`cliCheckLogSniffer`).

### Outputs
| Return | Type   | Meaning |
|--------|--------|---------|
| `int`  | Number of bytes consumed from `p`. The implementation always consumes all input, so it returns `len(p)`. |
| `error`| Always `nil`. Errors are not expected in this context. |

### Key Dependencies & Side‑Effects
1. **Terminal detection**  
   Calls the helper function `isTTY()` to decide whether the output stream is attached to an interactive terminal. If not a TTY, the method exits early, performing no further work.

2. **Channel communication**  
   - Converts the byte slice into a string and sends it on the package‑level channel `checkLoggerChan`.  
   - This channel is read by a separate goroutine that updates the CLI status line; thus, `Write` has the side effect of triggering UI refreshes indirectly.

3. **Length handling**  
   Uses `len(p)` to report how many bytes were processed and again when converting to a string (`string(p)`). The conversion is safe because slog guarantees UTF‑8 compliance for log messages.

4. **No external state mutation**  
   Apart from writing to `checkLoggerChan`, the method does not alter any global variables or package state.

### Integration with the Package
- **`cliCheckLogSniffer` struct** – This type is defined in `cli.go` and represents the slog handler used for check logs. By satisfying `io.Writer`, it can be passed to `slog.New()` as a custom output destination.
- **Ticker goroutine** – A ticker (period = `tickerPeriodSeconds`) reads from `checkLoggerChan` and writes formatted status lines to `os.Stdout`. The `Write` method is the only place where log entries are pushed onto that channel.
- **Graceful shutdown** – When the CLI is exiting, a signal is sent on `stopChan`, causing the ticker goroutine to terminate. `Write` remains functional until then.

### Usage Pattern
```go
logger := slog.New(slog.NewTextHandler(os.Stdout))
checkSniffer := cli.CliCheckLogSniffer{}
logger.SetOutput(&checkSniffer) // logger now uses Write()
```

Every call to `logger.Info`, `logger.Error`, etc., will invoke `Write`, which in turn updates the live CLI status.

---

#### Mermaid Diagram (suggested)

```mermaid
flowchart TD
    A[Log Emitted] -->|slog emits| B[slog Writer]
    B --> C{isTTY?}
    C -- no --> D[Return 0, nil]
    C -- yes --> E[Convert []byte → string]
    E --> F[Send on checkLoggerChan]
    F --> G[Ticker Goroutine updates status line]
```

This diagram illustrates the flow from a log event to the terminal display.
