ShouldRun` – pre‑flight decision helper

| Aspect | Detail |
|--------|--------|
| **Package** | `preflight` (`github.com/redhat-best-practices-for-k8s/certsuite/tests/preflight`) |
| **Signature** | `func ShouldRun(labels string) bool` |
| **Exported** | Yes – intended for use by test suites and helpers. |

### Purpose
`ShouldRun` determines whether the suite should perform pre‑flight checks on a given Kubernetes cluster.  
It short‑circuits expensive operations (`preflight.LoadChecks()`) when they are not needed, saving time in CI runs.

### Decision logic
The function returns `true` only if **both** of the following conditions hold:

1. **Pre‑flight tags present** – The supplied `labels` string must contain any label that matches one of the pre‑flight tag/label names.  
   This is evaluated by calling `labelsAllowTestRun(labels)`.  

2. **Docker configuration available** – The pre‑flight Docker config file must exist on disk.  
   The presence of this file is checked via `GetTestEnvironment()` which reads the environment and verifies that the required config path is set.

If either condition fails, the function returns `false`.

### Side effects
* Logs a warning when the Docker config is missing: `Warn("preflight dockerconfig does not exist")`.  
  This uses the package‑level `env` variable (type `provider.TestEnvironment`) to access test parameters.

No other state is mutated; the function is pure aside from the log side effect.

### Dependencies
| Dependency | Role |
|------------|------|
| `GetTestEnvironment()` | Reads the global `env` and returns the current test environment. |
| `labelsAllowTestRun(labels)` | Checks whether the label string contains any pre‑flight tag. |
| `GetTestParameters()` | Provides configuration values (e.g., Docker config path) used in logging. |
| `Warn(msg)` | Emits a warning log if prerequisites are missing. |

### How it fits into the package
`ShouldRun` is invoked at the start of the test suite (`suite.go`) before attempting to load pre‑flight checks:

```go
if !preflight.ShouldRun(env.Labels) {
    // Skip expensive check loading
}
```

By gating the execution on labels and config availability, the test harness avoids unnecessary work when running in environments that do not require pre‑flight validation. This function therefore acts as a lightweight gatekeeper for the more costly `preflight.LoadChecks()` call.
