HasExclusiveCPUsAssigned`

| Item | Detail |
|------|--------|
| **Signature** | `func HasExclusiveCPUsAssigned(c *provider.Container, logger *log.Logger) bool` |
| **Exported?** | Yes |

### Purpose
Determines whether a Kubernetes container has been allocated *exclusive* CPUs.  
The function parses the CPU and memory limits set on the container’s pod spec and checks if the CPU limit is an integer value that is not shared with other containers (i.e., it matches the “cpu‑exclusive” policy described in <https://kubernetes.io/docs/tasks/administer-cluster/cpu-management-policies/>).

### Inputs
| Parameter | Type | Role |
|-----------|------|------|
| `c` | `*provider.Container` | The container instance whose CPU allocation is inspected. |
| `logger` | `*log.Logger` | Logger used for debug‑level diagnostics; has no effect on the return value. |

### Output
| Return | Meaning |
|--------|---------|
| `bool` | `true` if the container’s CPU limit is a non‑zero integer and matches the exclusive CPU policy, otherwise `false`. |

### Key Dependencies
* **CPU & Memory extraction** – Uses helper functions `Cpu(c)` and `Memory(c)` to pull resource limits from the pod spec.
* **Zero checks** – Calls `IsZero()` on the extracted values to ensure a limit is actually set.
* **Conversion** – Uses `AsInt64()` to convert the quantity objects to plain integers for comparison.
* **Logging** – Emits debug messages via `logger.Debug(...)` at several stages, but never alters state.

### Side‑Effects
* No modification of the container or any global state.
* Only writes log entries; the function is pure from a functional perspective aside from logging.

### How It Fits in the Package
The `resources` package provides helpers for inspecting pod resource requests/limits.  
`HasExclusiveCPUsAssigned` is a predicate used by higher‑level tests (e.g., access‑control checks) to assert that a container conforms to CPU exclusivity requirements.  It acts as an isolated, reusable check that can be composed with other predicates in the test suite.

---

#### Suggested Mermaid Diagram
```mermaid
flowchart TD
    A[Container c] -->|Cpu(c)| B(CPU limit)
    A -->|Memory(c)| C(Memory limit)
    B --> D{IsZero?}
    C --> E{IsZero?}
    D -->|Yes| F(false)
    E -->|Yes| G(false)
    D -->|No| H[AsInt64()]
    E -->|No| I[AsInt64()]
    H & I --> J{Exclusive?}
    J -->|Yes| K(true)
    J -->|No| L(false)
```

This diagram illustrates the decision path that `HasExclusiveCPUsAssigned` follows to determine exclusivity.
