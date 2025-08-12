getK8sClientsConfigFileNames`

| Item | Details |
|------|---------|
| **Package** | `certsuite` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/certsuite`) |
| **Visibility** | unexported (private to the package) |
| **Signature** | `func() []string` |

### Purpose
Collects a list of Kubernetes client configuration file paths that will be used by the test harness.  
The function first tries to read an environment variable (`KUBECONFIG`) and then falls back to the default
configuration files defined in the package constants.

### Inputs / Environment
| Source | How it’s accessed |
|--------|-------------------|
| `GetTestParameters()` | Called at runtime; returns a map of test parameters. If the key `"kubeconfig"` is present, its value(s) are used as config paths. |
| `os.Getenv("KUBECONFIG")` | Reads the standard environment variable that can contain one or more kube‑config files separated by `:` on Unix or `;` on Windows. |

### Output
A slice of strings (`[]string`) containing absolute paths to each discovered configuration file.

The function guarantees that:

1. If the test parameters specify a kube‑config, those are used exclusively.
2. Otherwise, if the environment variable is set, its entries are added.
3. Finally, the package default (`kubeConfigDefaultValue` – defined elsewhere in the same file) is appended as a fallback.

If any path cannot be resolved to an existing file, it is omitted and a log entry is emitted via `Info()`.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `GetTestParameters` | Supplies user‑provided configuration paths. |
| `Getenv`, `Join`, `Stat` | Resolve environment variable values and validate existence of files. |
| `append` | Builds the result slice incrementally. |
| `Info` | Logs informational messages about missing or invalid config files. |

### Side Effects
- **Logging** – Calls to `Info()` emit messages if a file cannot be found, aiding debugging.
- **No state mutation** – The function does not alter any global variables; it only reads from them.

### How It Fits the Package
`certsuite` orchestrates end‑to‑end certificate tests against Kubernetes clusters.  
Before any test runs, a client must be created with valid credentials and cluster details.  
This helper is invoked during client construction to determine which kube‑config files should be loaded.
It abstracts away the logic of environment resolution and defaults, keeping the higher‑level code focused on test execution rather than configuration plumbing.

---

#### Suggested Mermaid diagram

```mermaid
flowchart TD
    A[Start] --> B{Test parameters contain "kubeconfig"?}
    B -- Yes --> C[Use values from GetTestParameters]
    B -- No --> D{KUBECONFIG env set?}
    D -- Yes --> E[Split and add env paths]
    D -- No --> F[Skip to default]
    E & C --> G[Validate each path with Stat()]
    G --> H{Path exists?}
    H -- Yes --> I[Append to result slice]
    H -- No --> J[Log via Info() and skip]
    I & J --> K[Continue loop]
    K --> L[After loop, append default config]
    L --> M[Return []string]
```

This diagram visualises the decision flow: test‑parameter override → env variable fallback → default file.
