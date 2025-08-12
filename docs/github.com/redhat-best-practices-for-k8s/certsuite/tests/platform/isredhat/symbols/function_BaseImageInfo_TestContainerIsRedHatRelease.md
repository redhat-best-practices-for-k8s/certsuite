BaseImageInfo.TestContainerIsRedHatRelease`

| Item | Details |
|------|---------|
| **Package** | `isredhat` – test helpers for detecting Red‑Hat based images |
| **Receiver** | `b BaseImageInfo` – holds context (e.g. container ID, command runner) |
| **Signature** | `func() (bool, error)` |
| **Exported** | Yes |

---

### Purpose
`TestContainerIsRedHatRelease` determines whether the Docker image running in the test container is a Red‑Hat Enterprise Linux (RHEL) release.  
It does so by:

1. Executing a command inside the container to read `/etc/os-release`.
2. Checking the resulting output against regular expressions that identify RHEL families.
3. Returning `true` if it matches, otherwise `false`.

The function is used in integration tests that need to conditionally skip or alter behavior when running on non‑RHEL images.

---

### Inputs
| Parameter | Type | Description |
|-----------|------|-------------|
| *None* | — | The method operates solely on the state of its receiver (`BaseImageInfo`). |

The receiver holds:
- `ContainerID` – ID of the container to inspect.
- `RunCommand` – a function (via `runCommand`) that executes commands inside the container.

---

### Outputs
| Return | Type | Description |
|--------|------|-------------|
| `bool` | Indicates whether the image is RHEL. |
| `error` | Non‑nil if command execution fails or parsing errors occur. |

A successful call returns `true/false` with a nil error; failures return an appropriate error and a zero value for the boolean.

---

### Key Dependencies
| Called Function | Role |
|-----------------|------|
| `runCommand` (internal) | Executes `cat /etc/os-release` inside the container. |
| `Info` (log helper) | Logs diagnostic information about command execution. |
| `IsRHEL` (helper) | Applies regex checks (`NotRedHatBasedRegex`, `VersionRegex`) to the output and returns a boolean. |

The function relies on two exported regular expressions defined in this package:

```go
const (
    NotRedHatBasedRegex = ...
    VersionRegex        = ...
)
```

These patterns detect non‑RHEL distributions and specific RHEL versions, respectively.

---

### Side Effects
- No state mutation: the method only reads container data.
- Emits log messages via `Info` for debugging purposes.

---

### Flow Diagram (Mermaid)

```mermaid
flowchart TD
    A[TestContainerIsRedHatRelease] --> B[runCommand("cat /etc/os-release")]
    B --> C{command succeeded?}
    C -- Yes --> D[output string]
    D --> E[IsRHEL(output)]
    E --> F{is RHEL?}
    F -- True --> G[return true, nil]
    F -- False --> H[return false, nil]
    C -- No --> I[log error via Info]
    I --> J[return false, err]
```

---

### Package Context
`isredhat` is a lightweight test helper package inside the Certsuite project.  
Its primary job is to provide runtime checks for container base images, enabling tests to adapt their expectations based on whether they run on RHEL or another distribution.

`TestContainerIsRedHatRelease` is one of the public helpers that other test suites import when they need a quick boolean flag about the underlying OS.
