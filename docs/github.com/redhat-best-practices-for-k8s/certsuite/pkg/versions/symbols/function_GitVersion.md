GitVersion` – Display the current Git build version

| Item | Description |
|------|-------------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/versions` |
| **Signature** | `func GitVersion() string` |
| **Exported** | Yes – available to callers outside the package |

### Purpose
`GitVersion` provides a human‑readable version string that represents the state of the current build:

1. **Released builds**  
   If the binary was built from a tagged release (`GitRelease` is non‑empty), it returns that tag (e.g., `"v1.2.3"`).

2. **Unreleased builds**  
   When `GitRelease` is empty, the function falls back to the *latest* previous release (`GitPreviousRelease`).  
   The returned string contains both the previous release tag and the current commit hash:
   ```
   <previous-release>-<git-commit>
   ```

3. **Display formatting**  
   The output is formatted for display purposes – e.g., `fmt.Println(GitVersion())` yields a concise identifier that can be shown in logs, CLI help text, or UI elements.

### Inputs / Globals Used
The function relies on the following exported globals defined in the same package:

| Global | Type | Role |
|--------|------|------|
| `GitRelease` | `string` | The tag of the current release (empty if not released). |
| `GitPreviousRelease` | `string` | Tag of the most recent previous release. Used when no current release exists. |
| `GitCommit` | `string` | Short commit hash used to identify an unreleased build. |

The globals are typically injected at build time via linker flags (e.g., `-X main.GitRelease=v1.2.3`). If none of these values are set, the function will return `"unknown"`.

### Output
A single string representing:

* The release tag if present; otherwise,
* `<previous-release>-<git‑commit>`.

The format is deterministic and suitable for display or logging.

### Side Effects & Dependencies
- **No side effects** – purely read‑only.
- Depends only on the exported globals listed above.  
  No external packages are imported inside this function (other than `fmt` for formatting, which is standard).

### How It Fits the Package

The `versions` package centralizes all version‑related metadata for CertSuite:

```go
// pkg/versions/versions.go
var (
    GitRelease          string // set by ldflags at build time
    GitPreviousRelease  string
    GitCommit           string
)
```

`GitVersion` is the public API that other parts of CertSuite (CLI, web UI, etc.) call to obtain a stable identifier for the running binary. It abstracts away the logic of determining whether a release tag exists and provides a consistent string format.

---

#### Suggested Mermaid diagram

```mermaid
graph TD
    GitRelease{GitRelease}
    PrevRel[GitPreviousRelease]
    Commit[GitCommit]
    Output[GitVersion()]

    GitRelease -->|non‑empty| Output
    GitRelease -->|empty| PrevRel
    PrevRel -->|exists| Output
    PrevRel -->|empty| Commit
    Commit -->|always used if prev empty| Output
```

This diagram illustrates the decision flow inside `GitVersion`.
