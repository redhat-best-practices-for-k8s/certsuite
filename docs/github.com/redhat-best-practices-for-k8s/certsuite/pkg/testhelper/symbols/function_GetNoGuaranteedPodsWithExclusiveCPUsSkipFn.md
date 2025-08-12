GetNoGuaranteedPodsWithExclusiveCPUsSkipFn`

**Package:** `github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper`  
**Location:** `testhelper.go:594`  

## Purpose
Returns a *skip function* that can be passed to the test framework (e.g., Ginkgo/Go test) to skip tests when **no non‑guaranteed pods are using exclusive CPUs** in the current environment.  
The function encapsulates logic for determining whether a test should run based on the presence of such pods.

## Signature
```go
func GetNoGuaranteedPodsWithExclusiveCPUsSkipFn(env *provider.TestEnvironment) func() (bool, string)
```
* **Input**
  - `env`: a pointer to `provider.TestEnvironment`. The environment provides access to cluster state and utilities used by the helper functions.
* **Output**
  - A function with signature `func() (bool, string)` that:
    * returns `(true, <reason>)` when the test should be skipped,
    * returns `(false, "")` otherwise.

## How it works
1. **Retrieve pods**  
   Calls `GetGuaranteedPodsWithExclusiveCPUs(env)`, which scans the cluster for all pods annotated as *guaranteed* (i.e., CPU requests equal to limits) that also request exclusive CPUs via the `cpuExclusive` annotation.
2. **Count pods**  
   Uses Go’s built‑in `len()` on the slice returned by `GetGuaranteedPodsWithExclusiveCPUs`.
3. **Decision logic**  
   - If the count is **zero**, it means there are *no* guaranteed pods with exclusive CPUs, so the test that requires such a pod cannot run. The skip function returns `(true, "No non‑guaranteed pods using exclusive CPUs")`.
   - Otherwise, it returns `(false, "")`, allowing the test to proceed.

## Dependencies
| Dependency | Role |
|------------|------|
| `GetGuaranteedPodsWithExclusiveCPUs` | Retrieves the relevant pod list. |
| `len()` | Counts the slice length. |

No other globals or side‑effects are used; the function is pure except for reading the environment.

## Usage Example

```go
// In a Ginkgo test:
SkipIf := GetNoGuaranteedPodsWithExclusiveCPUsSkipFn(env)
It("should do X", func() {
    skip, reason := SkipIf()
    if skip {
        Skip(reason) // from Ginkgo
    }
    // … test body …
})
```

## Integration Context
- **Test Helper**: Part of a suite that dynamically decides whether to run or skip tests based on the current cluster state.
- **No‑guaranteed pods**: In this context, “no guaranteed pods” refers to pods whose CPU requests are *not* equal to their limits. The helper specifically checks for exclusive CPUs; if none exist, tests that rely on such pods are skipped.

---

**Key takeaway:**  
`GetNoGuaranteedPodsWithExclusiveCPUsSkipFn` is a convenience factory that produces a skip predicate used in test suites to conditionally bypass tests when the cluster lacks non‑guaranteed pods with exclusive CPU annotations.
