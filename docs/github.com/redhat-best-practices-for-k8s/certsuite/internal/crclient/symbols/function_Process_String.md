Process.String` – Human‑readable representation of a containerised process

| Item | Detail |
|------|--------|
| **Package** | `crclient` (`github.com/redhat-best-practices-for-k8s/certsuite/internal/crclient`) |
| **Receiver type** | `Process` (struct defined in the same file) |
| **Signature** | `func (p Process) String() string` |

### Purpose
`Process.String` implements the `fmt.Stringer` interface for the `Process` struct.  
It returns a concise, human‑readable description of the process that is running inside a container image.  
The representation is used throughout the package to log or display process information (e.g., in test reports, debug logs, or UI panels).

### Inputs
* The method receives no explicit parameters – it operates on the receiver `p Process`.

### Outputs
* **String**: A formatted string that typically contains:
  * the process name (`Name`)
  * its executable path (`ExecPath`)  
  * optionally the PID if known (`PID`), or a placeholder such as `<unknown>`.

The exact format is defined by the implementation (likely `fmt.Sprintf("%s (%s)", p.Name, p.ExecPath)` or similar).  

### Key Dependencies
* **Standard library**: `fmt.Sprintf` – used to build the output string.
* No external packages are invoked directly inside this method.  
* The method indirectly depends on the public fields of `Process`, which may be populated by other parts of the `crclient` package (e.g., during container inspection or when parsing `/proc/<pid>/cmdline`).

### Side Effects
None – the function is pure: it only reads the receiver’s fields and returns a string.  
It does **not** modify the `Process`, nor interact with the filesystem, network, or Docker daemon.

### How It Fits the Package
* The `crclient` package orchestrates interactions with container runtimes (Docker, CRI-O, etc.) to introspect processes inside images.
* `Process.String` provides a convenient way to log or display those processes without exposing implementation details of the struct.  
* Other components such as test runners, reporters, or UI handlers import `crclient.Process` and rely on its `String()` method for consistent output.

---

#### Suggested Mermaid diagram (optional)

```mermaid
flowchart TD
    A[crclient.Process] -->|String()| B[String representation]
    B --> C{Used by}
    C --> D[Test runner logs]
    C --> E[UI report panels]
```

This method is a small but essential utility that keeps the rest of the package clean and readable.
