adjustLineMaxWidth`

| Aspect | Details |
|--------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/info` |
| **Exported?** | No – it is an internal helper. |
| **Signature** | `func()()` – returns a closure that takes no arguments and has no return value. |

### Purpose
`adjustLineMaxWidth` produces a lightweight, reusable *setter* that updates the package‑level variable `lineMaxWidth`.  
The setter determines how wide a line of output can be when printed to a terminal:

1. It checks whether the current standard output is attached to an interactive terminal (`IsTerminal`).  
2. If so, it queries the terminal’s size (`GetSize`) and subtracts a constant padding value (`linePadding`).  
3. The resulting width is stored in `lineMaxWidth`.  

This logic is executed lazily: the closure can be invoked at any point after the program has started (for example, when the `info` command is run), ensuring that the width reflects the actual terminal size at runtime.

### Inputs / Outputs
- **Input** – None. The function’s closure captures no external parameters; it relies solely on global state (`os.Stdout`, `linePadding`).  
- **Output** – It mutates the package‑level variable `lineMaxWidth`. No value is returned to callers of the closure.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `IsTerminal` (from a terminal handling package) | Determines if the current output stream is a TTY. |
| `GetSize` | Retrieves the current terminal’s width and height. |
| `linePadding` | A constant that reserves space on either side of printed lines. |

### Side Effects
- **Global state mutation** – `lineMaxWidth` is overwritten each time the closure runs.
- **I/O inspection** – The function may read from the OS (via `IsTerminal`/`GetSize`) but does not produce output.

### How It Fits in the Package

The `info` command displays diagnostic information about the CertSuite installation.  
Its output is formatted to fit within a terminal window, and `lineMaxWidth` controls that formatting width.  
By encapsulating the logic in a closure, the code that prints help or status messages can simply call the returned function once (or at the start of execution) to guarantee that subsequent formatting uses an appropriate width.

---

#### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[adjustLineMaxWidth()] --> B{IsTerminal?}
    B -- Yes --> C[GetSize() → (w, h)]
    C --> D[w - linePadding]
    D --> E[lineMaxWidth = w - linePadding]
    B -- No --> E
```

This diagram shows the decision path and how `lineMaxWidth` is ultimately set.
