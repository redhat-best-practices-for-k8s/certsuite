getHWTextOutput`

| Feature | Details |
|---------|---------|
| **Package** | `diagnostics` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/diagnostics`) |
| **Visibility** | Unexported (private to the package) |
| **Signature** | `func getHWTextOutput(pod *corev1.Pod, cmd clientsholder.Command, arg string) ([]string, error)` |

### Purpose
`getHWTextOutput` runs a single command inside a pod’s container and returns the output split into individual lines.  
It is used by higher‑level diagnostic functions that need raw text (e.g., CPU info, PCI devices, block devices). The function keeps the plumbing of executing commands isolated from business logic.

### Parameters
| Name | Type | Meaning |
|------|------|---------|
| `pod` | `*corev1.Pod` | Target pod in which to execute the command. |
| `cmd` | `clientsholder.Command` | Wrapper that knows how to run a command inside a container (`ExecCommandContainer`). |
| `arg` | `string` | The argument string passed to the command (e.g., `"cpuinfo"`). |

### Return Values
| Value | Type | Description |
|-------|------|-------------|
| `[]string` | slice of strings | Each element is a line from the command’s stdout. Empty slice if the command produced no output. |
| `error` | error | Non‑nil if any step fails (context creation, execution, or parsing). The returned error contains a stack trace and context information. |

### Key Dependencies & Calls
1. **Context** – `NewContext()` creates a cancellable context for the exec call.
2. **Command Execution** – `ExecCommandContainer` runs the actual command inside the container and streams stdout/stderr back to Go.
3. **Error Handling** – Errors are wrapped with `fmt.Errorf("%w", err)` (or similar) to preserve stack information.
4. **Output Parsing** – The raw byte slice is converted to a string and split on newline characters via `strings.Split`.

### Side‑Effects
* No state in the package or pod is mutated; it only performs read operations.
* The function logs nothing; errors are returned for callers to decide how to handle them.

### How It Fits the Package
The diagnostics package aggregates various hardware and environment checks. Each check typically:
1. Constructs a command (e.g., `lscpu`, `lsblk`).
2. Calls `getHWTextOutput` to obtain its plaintext output.
3. Parses that output into structured data or uses it directly for reporting.

Thus, `getHWTextOutput` is the common “run‑command‑and‑split” helper used across multiple diagnostic routines.

```mermaid
flowchart TD
    A[Caller (diagnostic function)] -->|passes pod, cmd, arg| B[getHWTextOutput]
    B --> C{Create Context}
    B --> D{ExecCommandContainer}
    D --> E[Read stdout bytes]
    E --> F[strings.Split on '\n']
    F --> G[Return []string, nil]
    D --> H{Error?}
    H -- yes --> I[Wrap error with fmt.Errorf]
    I --> J[Return nil, err]
```

---

**Note:** The function is intentionally simple and stateless; any additional parsing or formatting is left to the caller.
