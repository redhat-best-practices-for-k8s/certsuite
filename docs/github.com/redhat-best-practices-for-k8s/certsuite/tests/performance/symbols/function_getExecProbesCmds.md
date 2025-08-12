getExecProbesCmds`

| Aspect | Detail |
|--------|--------|
| **Package** | `performance` (github.com/redhat-best-practices-for-k8s/certsuite/tests/performance) |
| **Signature** | `func(*provider.Container) map[string]bool` |
| **Exported** | No |

---

### Purpose
`getExecProbesCmds` extracts the *exec* probe commands from a Kubernetes container specification and returns them as a set of unique strings.  
The function is used by performance test helpers to identify which exec probes should be run (or skipped) for each pod/container.

> **Why a map?**  
> The returned `map[string]bool` acts like a *set*: the keys are the probe commands and the value is always `true`. This makes it trivial to check whether a particular command exists without caring about duplicates.

---

### Inputs
| Parameter | Type | Meaning |
|-----------|------|---------|
| `c` | `*provider.Container` | A pointer to a container definition that contains a slice of `ExecProbe` objects (typically from a pod spec). |

> The function assumes that the `container.ExecProbes` field is populated; if it is nil or empty, an empty map will be returned.

---

### Outputs
| Return value | Type | Meaning |
|--------------|------|---------|
| `map[string]bool` | A set of all exec‑probe commands found in `c`. | Each key is the full command string that would be executed by Kubernetes for the probe. |

If no probes are present, the map will be empty.

---

### Key Dependencies & Helpers

The function relies only on standard library string helpers:

| Helper | Role |
|--------|------|
| `strings.Join` | Concatenates command parts (`Command`) into a single space‑separated string. |
| `strings.Fields` | Splits a command string back into words; used to normalize whitespace and extract the executable name. |

> No global variables or other package functions are referenced.

---

### Typical Flow (pseudo)

```go
func getExecProbesCmds(c *provider.Container) map[string]bool {
    cmds := make(map[string]bool)

    for _, p := range c.ExecProbes {          // iterate over exec probes
        cmdStr := strings.Join(p.Command, " ") // e.g. ["curl", "-f", "http://localhost"]
        baseCmd := strings.Fields(cmdStr)[0]   // take first word ("curl")
        cmds[baseCmd] = true                   // store as set element
    }

    return cmds
}
```

> *Note*: The actual implementation may contain additional logic (e.g., handling `InitialDelaySeconds`, or filtering by probe type), but the essence is to collect unique command strings.

---

### Side‑Effects

- **No mutation** of the input container (`c`) occurs.
- Only local variables are created; no state is persisted elsewhere.

---

### How It Fits the Package

The `performance` test suite uses this helper when:

1. Determining which exec probes should be executed as part of a performance check.
2. Skipping or isolating tests that rely on specific probe commands (e.g., those requiring network access).

By abstracting the extraction logic into its own function, the test code stays clean and reusable across multiple test cases.

---

### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[Container] -->|has| B[ExecProbes]
    B --> C{Loop over probes}
    C --> D[Join Command parts]
    D --> E[Split into fields]
    E --> F[Take first word (base cmd)]
    F --> G[Add to map set]
```

This diagram visualises the linear transformation from container spec → probe command string → unique set entry.
