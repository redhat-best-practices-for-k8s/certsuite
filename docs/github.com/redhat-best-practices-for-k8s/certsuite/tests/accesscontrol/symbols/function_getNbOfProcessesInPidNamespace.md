getNbOfProcessesInPidNamespace`

| Aspect | Detail |
|--------|--------|
| **Package** | `accesscontrol` (`github.com/redhat-best-practices-for-k8s/certsuite/tests/accesscontrol`) |
| **Visibility** | Unexported (lower‑case name) – used only within this test package. |
| **Signature** | `func getNbOfProcessesInPidNamespace(ctx clientsholder.Context, pid int, execCmd clientsholder.Command) (int, error)` |

---

### Purpose
The helper returns the number of processes that belong to the same PID namespace as a given process ID (`pid`) inside a container.

1. It builds and runs a shell command that queries `/proc/<pid>/status` for the `NSpid:` field.  
2. The value is parsed into an integer – this represents how many distinct PIDs are in that namespace from the perspective of the host.
3. Errors are wrapped with context if the command fails or the output cannot be parsed.

This function is used by tests that need to assert whether a process runs isolated (e.g., after a namespace change).

---

### Parameters
| Name | Type | Role |
|------|------|------|
| `ctx` | `clientsholder.Context` | Holds client information needed for executing commands inside containers. |
| `pid` | `int` | The PID whose namespace we want to inspect. |
| `execCmd` | `clientsholder.Command` | Function that actually runs a command inside the container; injected for testability. |

---

### Return Values
| Value | Type | Meaning |
|-------|------|---------|
| first return | `int` | Count of processes in the PID namespace (or 0 on error). |
| second return | `error` | Non‑nil if command execution or parsing failed; wrapped with a descriptive message. |

---

### Core Logic
1. **Command Construction**  
   ```go
   cmd := fmt.Sprintf("cat /proc/%s/status | grep NSpid:", strconv.Itoa(pid))
   ```
   The command fetches the `NSpid:` line from `/proc/<pid>/status`.

2. **Execution**  
   `out, err := execCmd(ctx, cmd)` – runs the command inside the target container.

3. **Parsing**  
   * Split the output into fields (`strings.Fields`) and take the last field (the PID count).  
   * Convert that string to an integer with `strconv.Atoi`.  

4. **Error Handling**  
   Each failure point returns a wrapped error using `fmt.Errorf` for clarity.

---

### Dependencies
- Standard library: `strconv`, `strings`, `fmt`.
- Test helper types from the package:
  - `clientsholder.Context`
  - `clientsholder.Command`

No global variables are accessed; all data comes through parameters or local variables, keeping the function pure and test‑friendly.

---

### Usage Context
The function is called by other test helpers that need to verify namespace isolation. For example:

```go
procCount, err := getNbOfProcessesInPidNamespace(ctx, targetPID, execCommand)
```

It provides a quick way to confirm whether a process has been moved into a new PID namespace (expected count = 1) or remains in the host namespace (count > 1).

---

### Diagram (optional)

```mermaid
flowchart TD
    A[Call getNbOfProcessesInPidNamespace] --> B[Build cmd: cat /proc/<pid>/status | grep NSpid:]
    B --> C[execCmd(ctx, cmd)]
    C --> D{Success?}
    D -- Yes --> E[Parse fields]
    E --> F[Atoi(fields[last])]
    F --> G[Return count]
    D -- No --> H[Wrap error & return]
```

---

**Summary:**  
`getNbOfProcessesInPidNamespace` is a lightweight helper that inspects the PID namespace of a process inside a container, returning the number of processes it sees. It relies only on injected command execution and basic string parsing, making it straightforward to unit‑test and safe for use in end‑to‑end tests.
