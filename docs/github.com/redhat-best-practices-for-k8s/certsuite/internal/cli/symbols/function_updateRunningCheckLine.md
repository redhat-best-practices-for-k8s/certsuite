updateRunningCheckLine`

### Purpose
`updateRunningCheckLine` is a helper that keeps the terminal line showing the status of a currently running check up‑to‑date.  
It runs in its own goroutine and updates the line every second until it receives a stop signal.

### Signature
```go
func updateRunningCheckLine(checkName string, stop <-chan bool) func()
```
* **`checkName`** – the name of the check whose status should be displayed.  
  The function uses this string only when printing the initial message.
* **`stop`** – a read‑only channel that signals the goroutine to terminate.

The returned value is an *initializer* that starts the ticker loop; it can be called immediately after creating the stop channel.

### How It Works
1. **TTY check** – If `isTTY()` returns `false`, the function does nothing and returns a no‑op closure.  
   This prevents terminal manipulation on non‑interactive streams.
2. **Ticker creation** – A `time.NewTicker` fires every `tickerPeriodSeconds` (1 s).  
3. **Loop** – On each tick the goroutine:
   * Calls `printRunningCheckLine(checkName)` to refresh the status line.
   * Uses `Now()` only for timestamping inside that helper.
4. **Termination** – The loop listens on `stop`. When a value is received:
   * The ticker is stopped (`Stop()`).
   * The function prints a final status line indicating completion via `printRunningCheckLine(checkName)`.

### Dependencies
| Dependency | Role |
|------------|------|
| `isTTY` | Detects interactive terminal. |
| `time.NewTicker`, `ticker.Stop` | Drives periodic updates. |
| `Now` | Provides current time for the status line. |
| `printRunningCheckLine` | Renders the status text on the terminal. |

### Side‑Effects
* Writes to standard output (overwrites the same line each second).  
  Uses ANSI escape codes (`ClearLineCode`) to clear and rewrite the line.
* Blocks only while running; it terminates when a message arrives on `stop`.

### Package Context
`updateRunningCheckLine` is part of the internal CLI package that drives certsuite’s interactive command‑line interface.  
It collaborates with:

- **`CliCheckLogSniffer`** – a global logger that may emit additional log lines; the status line remains visible above those logs.
- **`checkLoggerChan` / `stopChan`** – channels used elsewhere to coordinate check execution and termination.

In practice, when a test is launched, the CLI creates a `stopChan`, starts this helper by invoking its returned closure, and closes `stopChan` once the test finishes. The function then cleans up the ticker and prints the final status line.
