PrintCheckRunning`

**File:** `internal/cli/cli.go`  
**Line:** 181

```go
func PrintCheckRunning(name string) func() {
    // …
}
```

### Purpose
`PrintCheckRunning` creates a *cleanup* function that displays a live “running” message for the test named `name`.  
The returned closure is intended to be called after the check has finished (or aborted). It stops an internal ticker, clears the status line and prints the final result.

Typical usage:

```go
defer PrintCheckRunning(checkName)()
```

### Parameters

| Name | Type   | Description |
|------|--------|-------------|
| `name` | `string` | The name of the check being executed. It is interpolated into the status line shown to the user.|

### Return Value

* A **zero‑argument function** (`func()`) that, when called, stops the progress ticker and renders the final check result on the terminal.

### Key Dependencies & Calls

| Call | Purpose |
|------|---------|
| `make([]byte, len)` (twice) | Builds a byte slice to hold the status line. The first call creates an empty line of spaces; the second writes the formatted running text (`"[name] …"`). |
| `isTTY()` | Detects if stdout is attached to a terminal. If not, no live updates are attempted and the function returns immediately. |
| `Print(string)` | Prints a single line (used for the final status after stopping the ticker). |
| `updateRunningCheckLine([]byte)` | Periodically updates the status line while the check runs. The returned cleanup function stops this goroutine. |

### Side Effects

1. **Terminal Output** – While active, the function writes an in‑place “running” message to stdout using ANSI escape codes (`ClearLineCode`).  
2. **Goroutines & Channels** – Launches a ticker that updates the line every `tickerPeriodSeconds`. The returned cleanup function closes `stopChan`, signaling the goroutine to exit.
3. **State Persistence** – It writes status information into the global channel `checkLoggerChan` (not shown in this snippet but used elsewhere for logging).

### Interaction with Package State

- `CliCheckLogSniffer` and the two channels (`checkLoggerChan`, `stopChan`) are part of the package‑wide log/sniffing system.  
- `PrintCheckRunning` does **not** modify those globals directly; it only reads `isTTY()` to decide whether to engage the live update mechanism.
- The cleanup function returned by `PrintCheckRunning` is responsible for closing `stopChan`, ensuring that the ticker goroutine exits cleanly.

### Summary

`PrintCheckRunning` is a helper that gives users visual feedback while a test runs. It encapsulates the logic of starting a progress ticker, updating a terminal line in place, and providing a convenient defer‑able cleanup function to stop the ticker and print the final status when the check completes.
