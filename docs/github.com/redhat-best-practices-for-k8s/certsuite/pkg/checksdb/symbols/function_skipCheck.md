skipCheck` – Internal helper for skipping a check

```go
func skipCheck(c *Check, reason string) func()
```

| Element | Description |
|---------|-------------|
| **Purpose** | Creates a deferred function that marks a check as *skipped* and logs the skip reason. The returned closure is meant to be executed via `defer` in a check’s execution routine so that skipping is recorded even if the check panics or returns early. |
| **Parameters** | • `c *Check` – the check instance being processed.<br>• `reason string` – human‑readable message explaining why the check was skipped (e.g., “missing required label”). |
| **Return value** | A zero‑argument function (`func()`) that, when called, performs three actions:<ol><li>Logs the skip reason via `LogInfo`. <li>Sets the check’s result to *Skipped* using `SetResultSkipped(c)`. <li>Prints a concise status line with `printCheckResult(c)` (the same helper used for passed/failed checks).</ol> |
| **Key dependencies** | • `LogInfo` – logs informational messages to the console or log file.<br>• `SetResultSkipped` – updates the check’s internal state and result enumeration (`CheckResultSkipped`).<br>• `printCheckResult` – formats and outputs a single line summarizing the check outcome. |
| **Side effects** | 1. Mutates the passed `*Check`, setting its status to *Skipped*. <br>2. Emits log output (via `LogInfo`) and console output (`printCheckResult`). No other global state is modified. |
| **Package context** | The function lives in `checksdb/checksgroup.go` and is used by the execution engine when a check’s pre‑conditions fail or an external skip condition is triggered. It centralises the “skip” logic so that all skips have consistent logging, status setting, and output formatting. |

### How it fits into the workflow

1. **Check execution** – A `ChecksGroup` runs its checks in sequence.  
2. **Pre‑condition check** – If a check’s pre‑conditions (labels, runtime env, etc.) are not met, the group code calls `skipCheck(c, reason)` and defers the returned function.  
3. **Deferred execution** – At the end of the check’s wrapper (whether it panics or returns), Go runs the deferred closure, ensuring that the skip is recorded even if the check failed to complete normally.  

### Example usage

```go
if !labelsExprEvaluator.Eval(c.Labels) {
    defer skipCheck(c, "label expression not satisfied")()
}
// … rest of check logic …
```

This pattern guarantees that a skipped check never appears as *passed* or *failed* in the final results and that users receive clear feedback on why it was omitted.
