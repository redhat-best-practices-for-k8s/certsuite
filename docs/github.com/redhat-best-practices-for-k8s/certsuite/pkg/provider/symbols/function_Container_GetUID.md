Container.GetUID`

| Item | Detail |
|------|--------|
| **Receiver** | `c Container` – the container instance whose UID is being queried |
| **Signature** | `func (c Container) GetUID() (string, error)` |
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider` |

### Purpose
`GetUID` extracts the **container ID** from a container's `ContainerID` field and returns it as a plain UID string.  
The function is used by higher‑level helpers that need to identify a running container, e.g., when checking logs or performing health probes.

### Inputs & Outputs
| Input | Description |
|-------|-------------|
| `c.ContainerID` | A string that typically looks like `"docker://<hash>"` or `"containerd://<hash>"`. |

| Output | Type | Meaning |
|--------|------|---------|
| `string` | The UID part of the container ID (the `<hash>` after the scheme). |
| `error` | Returned when `ContainerID` is empty, malformed, or cannot be parsed. |

### Algorithm
1. **Split** the `ContainerID` on `":"`.  
   *The split yields `[scheme, uid]` for valid IDs.*  
2. **Validate** the result:  
   * If no colon or more than two parts → error.  
3. **Trim** any leading/trailing whitespace from the UID part.  
4. Return the trimmed UID and `nil` error.

The function logs a debug message when it succeeds, using the package‑level `Debug` logger (`logrus`). It also creates a new logger instance with `New("Container.GetUID")` for contextual logging.

### Dependencies
| Dependency | Role |
|------------|------|
| `strings.Split` | Splits the container ID string. |
| `len` | Validates split result length. |
| `Debug`, `New` (from package’s logger) | Produce debug logs. |

No global variables or external packages are accessed.

### Side Effects
- Produces a log entry on success.
- Does **not** modify the container instance or any global state.

### Package Fit
Within the `provider` package, containers are represented by the `Container` struct (defined in `containers.go`).  
Other functions such as `GetContainerName`, `IsIgnored`, and status checks rely on `GetUID` to uniquely identify a container when querying Kubernetes APIs or inspecting runtime data.  
Thus, `GetUID` is a small but essential helper that bridges raw container IDs from the Kubernetes API to a clean UID usable by the rest of the provider logic.
