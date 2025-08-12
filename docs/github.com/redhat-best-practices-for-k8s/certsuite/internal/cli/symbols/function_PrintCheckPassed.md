PrintCheckPassed`

### Purpose
`PrintCheckPassed` is a helper that produces a **zero‑argument closure** used by the CLI to signal that an individual check has finished successfully.  
The returned function:

1. Stops any visual “checking” animation that may be running for the current line.
2. Prints a green‑colored success message in place of the animated spinner.

This keeps the user interface tidy and provides consistent feedback across all checks.

### Signature
```go
func PrintCheckPassed(msg string) func()
```
* **Input** – `msg` (string): The human‑readable name or description of the check that has passed.
* **Output** – a function with no parameters and no return value.  
  When invoked, it performs the side effects described above.

### Dependencies & Side Effects
| Dependency | Role |
|------------|------|
| `stopCheckLineGoroutine()` | Stops the goroutine that animates a spinner for the current check line. |
| `Print(msg)` | Outputs the final success message to stdout with appropriate colour and formatting. |

Both called functions are defined elsewhere in this package:

* `stopCheckLineGoroutine` is responsible for signalling the animation goroutine (via `stopChan`) to cease.
* `Print` writes coloured text to the terminal, using constants like `Green`, `Reset`, etc.

### Interaction with Package State
The function does **not** read or modify any global variables directly.  
However, it relies on the state that a check line is currently being animated (which is managed by other parts of the package). By stopping that goroutine and printing a final message, it ensures that the terminal output remains coherent.

### Usage Context
Typical flow in the CLI:

```go
// When starting a check:
startCheckLineGoroutine(msg)

// Later, after the check succeeds:
defer PrintCheckPassed(msg)()
```

The deferred call guarantees that the success message is printed exactly once when the surrounding function exits successfully.

---

#### Mermaid diagram (suggested)

```mermaid
graph TD
  A[Start Check] --> B[startCheckLineGoroutine]
  B --> C{Spinner Running}
  C -->|Stop| D[stopCheckLineGoroutine]
  D --> E[Print(msg) with Green]
```

This illustrates how `PrintCheckPassed` ties together the spinner shutdown and final output.
