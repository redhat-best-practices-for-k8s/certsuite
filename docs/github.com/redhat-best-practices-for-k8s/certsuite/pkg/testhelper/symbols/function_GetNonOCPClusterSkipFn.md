GetNonOCPClusterSkipFn`

| Item | Detail |
|------|--------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper` |
| **Exported** | Yes |
| **Signature** | `func() (func() (bool, string))` |

### Purpose
`GetNonOCPClusterSkipFn` returns a *factory* that produces a function used by the test suite to decide whether a given test should be skipped on non‑OpenShift clusters.

The returned inner function implements the standard “skip if not OCP” logic. It is intended for use in tests that are only relevant when running against an OpenShift (OCP) control plane.

### Return value
A **closure** with signature `func() (bool, string)`:

- **`bool`** – indicates whether the test should be skipped (`true`) or run (`false`).
- **`string`** – a message explaining why the test was skipped.  
  When `true`, the message will typically state that the cluster is not OCP.

The closure captures any state required to perform the check (currently none beyond what `IsOCPCluster` provides).

### Key dependencies
| Dependency | Role |
|------------|------|
| `IsOCPCluster()` | Called inside the returned function to determine if the current cluster is an OpenShift cluster. If it returns `false`, the test will be skipped. |

No other global variables or types are used.

### Side‑effects
- None. The function only reads from the environment via `IsOCPCluster` and constructs a closure; no state is mutated.

### How it fits the package
The `testhelper` package contains utilities for writing integration tests against Kubernetes / OpenShift clusters.  
Other test helpers (e.g., `GetOCPClusterSkipFn`) provide similar factories that skip tests on non‑OpenShift environments.  

`GetNonOCPClusterSkipFn` is the counterpart: it returns a skip function that **skips** when *not* running on OCP, allowing tests to be gated on the presence of OpenShift features.

### Typical usage

```go
// In a test file:
skip := GetNonOCPClusterSkipFn()
if skip() {
    t.Skip(skipMessage)
}

// ...rest of test logic that assumes an OCP cluster...
```

The returned closure is lightweight and can be reused across multiple tests, ensuring consistent skip behaviour.

### Summary

- **What it does**: Provides a reusable “skip if not OpenShift” function.
- **Inputs/outputs**: No inputs; outputs a closure returning `(bool, string)`.
- **Dependencies**: Relies on `IsOCPCluster()` to detect the cluster type.
- **Side‑effects**: None.

This helper centralises skip logic for non‑OCP clusters, improving test maintainability and clarity.
