stopCheckLineGoroutine`

| Aspect | Details |
|--------|---------|
| **Package** | `cli` (internal/cli) |
| **Signature** | `func() ()` – returns a function that, when called, stops the goroutine responsible for printing check‑status lines. |
| **Visibility** | Unexported; used only within the package. |

### Purpose
During test execution the CLI starts a background goroutine that reads status updates from `CliCheckLogSniffer` (a custom logger that emits log lines) and writes them to the console in a user‑friendly format.  
`stopCheckLineGoroutine` creates a *cleanup function* that, when invoked, signals this goroutine to terminate gracefully.

### How It Works
1. **Channel Coordination** – The goroutine listens on two channels:
   - `checkLoggerChan`: receives formatted status strings.
   - `stopChan`: receives a boolean indicating the goroutine should exit.
2. **Return‑Closure** – `stopCheckLineGoroutine` simply returns an anonymous function that, when called, sends `true` into `stopChan`.  
   ```go
   return func() {
       stopChan <- true
   }
   ```
3. The caller typically stores this returned closure and defers its execution or calls it after the test run completes.

### Inputs / Outputs
- **Input**: None.
- **Output**: A `func()` that, when executed, sends a termination signal on `stopChan`.

### Dependencies & Side Effects
| Dependency | Role |
|------------|------|
| `stopChan` (`chan bool`) | Channel used to notify the status‑printing goroutine to stop. |
| `checkLoggerChan` | Not directly used in this function but part of the same goroutine’s context. |
| Goroutine launched elsewhere (not shown) | The returned closure affects that goroutine by sending a signal on its stop channel. |

No other global state is modified, and no return value is produced besides the closure.

### Package Context
The `cli` package implements command‑line interactions for CertSuite.  
- **`CliCheckLogSniffer`** captures log output from tests.  
- The background goroutine (started elsewhere in the file) formats these logs into a concise progress line.
- `stopCheckLineGoroutine` provides a clean shutdown mechanism, ensuring no stray goroutines linger after a test run.

```mermaid
flowchart TD
    A[runTests()] --> B{start status goroutine}
    B --> C[goroutine reads from checkLoggerChan]
    C --> D[prints line]
    B --> E[waits for stopChan signal]
    F[stopCheckLineGoroutine] --> G(returned closure)
    G --> H[closure sends true on stopChan]
    H --> E
```

**Use‑case**: In the main test runner, after all tests finish or if an interrupt occurs, the code calls `defer stopFunc()` where `stopFunc := stopCheckLineGoroutine()`. This guarantees that the status goroutine terminates cleanly.
