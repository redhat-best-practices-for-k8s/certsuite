FsDiff.runPodmanDiff`

| Item | Details |
|------|---------|
| **Receiver** | `FsDiff` – the struct that holds state for a file‑system diff operation. |
| **Signature** | `func (f *FsDiff) runPodmanDiff(containerID string) (string, error)` |
| **Visibility** | Unexported; used internally by the package. |

### Purpose

`runPodmanDiff` runs a `podman diff` command against a running container identified by `containerID`.  
It extracts the list of file‑system changes made inside that container compared to its base image and returns them as a single string.

The function is part of **cnffsdiff**, a test helper that validates configuration files on Kubernetes nodes.  
During tests, after a container has been started (via `podman run`), this method gathers the diff output so that it can be parsed later to verify expected changes.

### Inputs

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `containerID` | `string` | The Podman container identifier returned by a previous `ExecCommandContainer("podman", "run"... )`. |

### Outputs

| Return value | Type    | Meaning |
|--------------|---------|---------|
| `stdout string` | `string` | Raw output of the `podman diff` command.  It contains one line per change (e.g., `A /path/to/file`). |
| `err error` | `error` | Non‑nil if any step fails (command execution, parsing, etc.). |

### Key Steps & Dependencies

1. **Command Construction**  
   ```go
   cmd := fmt.Sprintf("podman diff %s", containerID)
   ```
   Builds the shell command string.

2. **Execution via `ExecCommandContainer`**  
   Calls the helper that runs a command in the host’s Podman environment and captures its output.  
   This function is defined elsewhere in the same package (not shown) but is responsible for:
   - Spawning a process
   - Capturing stdout/stderr
   - Returning combined output or an error.

3. **Error Handling**  
   * If execution fails, returns `fmt.Errorf("podman diff %s failed: %w", containerID, err)`.*

4. **Return Result**  
   On success, simply returns the command’s standard output.

### Side‑Effects

- No modification of package globals or state; purely reads from the host.
- The command may touch the host file system indirectly (Podman diff inspects container layers), but this is considered a read‑only side effect.

### Placement in the Package

`runPodmanDiff` sits inside `FsDiff`, which orchestrates a series of operations:
1. Prepare temporary directories (`tmpMountDestFolder`, `nodeTmpMountFolder`).
2. Mount host paths into containers.
3. Execute tests and gather diffs via this method.
4. Clean up.

Thus, `runPodmanDiff` is the bridge between container runtime introspection and the test harness’s diff‑parsing logic.

---

#### Suggested Mermaid Flow (optional)

```mermaid
flowchart TD
    A[FsDiff.runPodmanDiff(containerID)] --> B["Build 'podman diff <id>'"]
    B --> C[ExecCommandContainer]
    C -- success --> D[Return stdout]
    C -- failure --> E[Return error]
```

This diagram visualises the single‑step execution path and its possible outcomes.
