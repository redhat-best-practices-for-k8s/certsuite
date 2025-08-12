GetVersionK8s`

| Aspect | Detail |
|--------|--------|
| **Package** | `diagnostics` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/diagnostics`) |
| **Signature** | `func GetVersionK8s() string` |
| **Exported?** | Yes |

### Purpose
`GetVersionK8s` returns the Kubernetes version that is currently being used by the test environment.  
It is a helper for diagnostic and reporting tooling: callers can embed this information in logs, reports or dashboards to identify which Kubernetes release a given test run exercised.

### Inputs / Outputs
- **Inputs**: none (the function takes no arguments).
- **Output**: a `string` containing the Kubernetes version.  
  The string is typically something like `"v1.28.3"` or `"v1.27.0"`. If the environment cannot be determined, the function may return an empty string (the exact behaviour depends on `GetTestEnvironment`).

### Key Dependencies
| Dependency | How it’s used |
|------------|---------------|
| `GetTestEnvironment()` | The only external call made by this function. It is expected to return a struct or value that contains information about the test environment, including the Kubernetes version. `GetVersionK8s` extracts the version field from that result and returns it. |

### Side‑Effects
- None.  
  The function performs read‑only operations: it queries the current test environment but does not modify any state.

### Integration with the Package
Within the `diagnostics` package, several functions gather system information (e.g., CPU, disk layout, PCI devices). `GetVersionK8s` complements those by providing cluster‑level context. Typical usage pattern:

```go
// Gather all diagnostics for a test run
info := diagnostics.GetTestEnvironment()
k8sVer := diagnostics.GetVersionK8s()

report := fmt.Sprintf("Running on Kubernetes %s with env %+v", k8sVer, info)
```

The function is intentionally simple to keep the diagnostic logic straightforward and deterministic. It relies on `GetTestEnvironment` for all heavy lifting, which abstracts how the environment information is obtained (e.g., via in‑cluster APIs or test harness metadata).
