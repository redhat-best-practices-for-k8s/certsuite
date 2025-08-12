main()` – Package entry point

### Purpose
`main()` is the *only* exported function in the `cmd/certsuite` package and serves as the **program entry point** for the CertSuite CLI.  
It orchestrates command‑line parsing, execution, and error handling by delegating to Cobra’s root command infrastructure.

### Signature
```go
func main() func()
```
- **Returns**: a function with no parameters and no return value.  
  The returned closure is intended for use in tests or other callers that wish to invoke the CLI logic without starting an actual process (e.g., when embedding CertSuite as a library).

### Key Dependencies & Calls
| Call | Purpose |
|------|---------|
| `newRootCmd()` | Builds and returns the top‑level Cobra command tree configured for CertSuite. |
| `Execute()` | Executes the command tree, parsing flags/args from `os.Args`. It blocks until completion or an error occurs. |
| `Error(err)` | Logs a fatal error using Go’s standard `log` package (or the project's logger). |
| `Exit(code int)` | Terminates the process with the given exit status (typically 0 for success, non‑zero for failure). |

These calls are all synchronous; after `Execute()` returns an error, the function logs it and exits immediately.

### Input / Output
- **Input**: The function reads the global `os.Args` slice implicitly via Cobra’s `Execute()`. No other external input is required.
- **Output**:  
  - On success: returns a no‑op closure (the caller may ignore it).  
  - On failure: logs the error and exits with status 1; the returned closure will never be reached.

### Side Effects
- Writes to standard output/error via Cobra’s logging mechanisms.  
- Calls `os.Exit`, which terminates the running process, so any deferred or goroutine work after this point is not guaranteed to run.

### Package Context
`cmd/certsuite` contains the CLI implementation of CertSuite. The `main()` function is the bridge between the Go binary and the Cobra command framework. All user-facing commands are registered in `newRootCmd()`, making `main()` a thin wrapper that ensures proper startup, error handling, and graceful termination.

---

#### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[main()] --> B[newRootCmd()]
    B --> C[Cobra Command Tree]
    C --> D[Execute()]
    D -->|error? | E[Error(err)]
    E --> F[Exit(1)]
    D -->|success| G[return closure]
```

This diagram visualises the linear control flow from program start to termination.
