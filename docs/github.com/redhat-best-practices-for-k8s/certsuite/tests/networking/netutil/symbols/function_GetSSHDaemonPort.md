GetSSHDaemonPort`

> **Purpose**  
> Retrieve the TCP port on which the SSH daemon inside a given container is listening.

### Signature
```go
func GetSSHDaemonPort(c *provider.Container) (string, error)
```
* **Input** – A pointer to a `provider.Container` that represents a running container instance.
* **Output**
  * `port string` – The port number (as a string) on which the SSH service is bound inside the container.
  * `err error` – Non‑nil if the command cannot be executed or the expected output cannot be parsed.

### How it works

1. **Command construction**  
   The function builds an exec command that runs inside the container:
   ```bash
   ss -tuln | grep sshd | awk '{print $4}'
   ```
   * `ss -tuln` lists all listening TCP sockets.
   * `grep sshd` filters for entries belonging to the SSH daemon.
   * `awk '{print $4}'` extracts the 4th field, which contains the local address and port (e.g. `0.0.0.0:22`).

2. **Execution**  
   It delegates to `ExecCommandContainerNSEnter(c, cmd)` – a helper that executes a shell command inside the container using its namespace entry point.

3. **Parsing the result**  
   The function expects the output to be a single line with the format `<addr>:<port>`.  
   * It trims whitespace.
   * Splits on `:` and returns the part after the colon as the port string.

4. **Error handling**  
   * If command execution fails, it wraps the error with context using `fmt.Errorf`.
   * If parsing does not yield a valid port (e.g., missing colon), an error is returned.

### Dependencies

| Dependency | Role |
|------------|------|
| `ExecCommandContainerNSEnter` | Executes shell commands inside the container. |
| `Errorf` (`fmt.Errorf`) | Formats and returns detailed errors. |

### Side effects

* None beyond reading container state; it does **not** modify the container or its configuration.

### Package context

The `netutil` package provides utilities for inspecting networking aspects of containers used in Certsuite tests.  
`GetSSHDaemonPort` is a helper that lets test code discover which port the SSH daemon is exposed on, enabling subsequent network connectivity checks (e.g., verifying that an external client can reach the container via SSH).  

```mermaid
flowchart TD
    A[Container] -->|ExecCommandContainerNSEnter| B{ss -tuln}
    B --> C[grep sshd]
    C --> D[awk {print $4}]
    D --> E{Parse port}
    E --> F[Return port or error]
```

---
