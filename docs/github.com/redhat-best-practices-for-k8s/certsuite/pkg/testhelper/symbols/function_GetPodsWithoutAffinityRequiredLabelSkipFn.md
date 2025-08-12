GetPodsWithoutAffinityRequiredLabelSkipFn`

| | |
|-|-|
| **Package** | `testhelper` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper`) |
| **Signature** | `func(*provider.TestEnvironment) (func() (bool, string))` |
| **Exported?** | Yes |

### Purpose

`GetPodsWithoutAffinityRequiredLabelSkipFn` returns a *skip‑function* that can be used by the test harness to conditionally skip tests when pods are missing the required affinity label.  
The function is typically passed to `provider.TestEnvironment.SkipIf()` (or similar) so that the surrounding test can decide at runtime whether it should continue or abort.

### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `env` | `*provider.TestEnvironment` | The current test environment. It is passed to the internal call to `GetPodsWithoutAffinityRequiredLabel`, which enumerates pods that lack the necessary affinity label in this environment. |

### Returned value

A **closure** of type `func() (bool, string)`:

- When invoked, it returns a tuple:
  - `bool`: indicates whether the test should be skipped (`true`) or not (`false`).
  - `string`: an explanatory message that will be logged when skipping.

The closure captures the environment and defers the expensive check until execution time.

### Implementation details

```go
func GetPodsWithoutAffinityRequiredLabelSkipFn(env *provider.TestEnvironment) func() (bool, string) {
    return func() (bool, string) {
        // Obtain a list of pods that are missing the required affinity label.
        pods := GetPodsWithoutAffinityRequiredLabel(env)

        // Skip only if there is at least one such pod.
        if len(pods) > 0 {
            return true, fmt.Sprintf("Skipping test: %d pod(s) lack affinity-required label", len(pods))
        }
        return false, ""
    }
}
```

* Calls:
  * `len` – to check whether any pods were returned.
  * `GetPodsWithoutAffinityRequiredLabel(env)` – helper that queries the environment’s API server for pods missing the label.

### Side‑effects

- None. The function merely builds a closure; it does not modify the test environment or external state.
- The returned closure will query the Kubernetes API when called, so it may incur network I/O and latency.

### How it fits the package

`testhelper` provides utilities for building and executing compliance tests against a Kubernetes cluster.  
The *skip‑function* pattern is used throughout the package to defer test decisions until runtime (e.g., depending on the current cluster state).  

- `GetPodsWithoutAffinityRequiredLabelSkipFn` is part of that infrastructure, specifically handling the scenario where pods lack the mandatory affinity label.
- It is typically used in tests that validate pod scheduling policies; if the necessary label is missing, the test is deemed irrelevant and is skipped rather than failing.

### Usage example

```go
env := provider.NewTestEnvironment(...)
skipFn := GetPodsWithoutAffinityRequiredLabelSkipFn(env)
if skip, msg := skipFn(); skip {
    t.Skip(msg) // or env.SkipIf(...), depending on harness
}
```

This pattern keeps test code concise while allowing dynamic adaptation to the cluster’s current configuration.
