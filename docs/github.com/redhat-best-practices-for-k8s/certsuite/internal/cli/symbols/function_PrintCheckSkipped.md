PrintCheckSkipped`

```go
func PrintCheckSkipped(name string, errMsg string) func()
```

### Purpose  
`PrintCheckSkipped` is a helper that logs the fact that a particular check was *skipped* during test execution.  
It is part of the CLI output formatting package (`github.com/redhat-best-practices-for-k8s/certsuite/internal/cli`) and is used whenever a check is skipped because of configuration, environment or prerequisite issues.

### Inputs

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `name`    | `string` | The human‑readable name of the check that was skipped. |
| `errMsg`  | `string` | Optional message explaining why the check was skipped (e.g., “missing cert file”). |

### Output

The function returns a **zero‑argument closure** (`func()`).  
When invoked, this closure performs two actions:

1. **Stops any ongoing line spinner** – it calls `stopCheckLineGoroutine()` which signals the goroutine that prints the live progress indicator to finish.
2. **Prints a formatted skip message** – it delegates to the generic `Print` function (which writes to stdout with colour formatting) passing in the tag constant `CheckResultTagSkip` and the supplied name/message.

Because it returns a closure, callers can defer its execution until after any asynchronous work is finished, ensuring that the spinner stops before the final message appears.

### Key Dependencies

| Dependency | Role |
|------------|------|
| `stopCheckLineGoroutine()` | Signals the goroutine responsible for printing the progress line to terminate. |
| `Print(tag string, name string)` | Handles coloured output; it receives `CheckResultTagSkip` as the tag and concatenates the check name with the optional message. |

Both dependencies are defined in the same package (`cli`) and rely on shared global channels (`checkLoggerChan`, `stopChan`) to coordinate concurrent printing.

### Side Effects

* **Channel communication** – `stopCheckLineGoroutine()` sends a signal over `stopChan` that causes the spinner goroutine to exit.  
* **Stdout output** – The closure writes the skip message to stdout via `Print`. No other state is mutated.

### Relationship to the Package

Within the `cli` package, each check status (pass, fail, error, aborted, running, skip) has a dedicated helper (`PrintCheckPassed`, `PrintCheckFailed`, etc.).  
`PrintCheckSkipped` follows the same pattern:

1. It stops any active progress indicator.
2. It emits a coloured message tagged with `CheckResultTagSkip`.

These helpers are used by the higher‑level test runner to provide real‑time feedback to users in the terminal.

### Example Usage

```go
// When a check is determined to be skipped:
defer PrintCheckSkipped("TLSv1.2", "certificate missing")()
```

The defer ensures that the spinner stops before printing the skip message, keeping the UI tidy and deterministic.
