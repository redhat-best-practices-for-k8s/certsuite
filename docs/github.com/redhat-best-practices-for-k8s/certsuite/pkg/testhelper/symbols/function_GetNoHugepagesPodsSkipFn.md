GetNoHugepagesPodsSkipFn`

**Package**: `github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper`  
**Exported**: yes

## Purpose
Creates a *skip‑function* used by the test framework to determine whether the **“no‑hugepages”** test should be skipped for a given test environment.  
The returned function returns:

1. `bool` – whether the test should be skipped (`true`) or run (`false`).  
2. `string` – an explanatory message used when skipping.

This helper centralises the logic that checks if any pod in the test environment is using hugepages, which would invalidate the “no‑hugepages” rule.

## Signature
```go
func GetNoHugepagesPodsSkipFn(env *provider.TestEnvironment) func() (bool, string)
```

* `env` – The test environment containing cluster state.  
* Returns a closure that can be invoked during test execution to decide on skipping.

## Implementation Details

1. **Capture the current hugepage pod list**  
   ```go
   hpPods := GetHugepagesPods(env)
   ```
   * `GetHugepagesPods` scans the environment and returns all pods requesting hugepages.

2. **Return a closure that evaluates at test time**  
   The closure uses the captured `hpPods` slice:

   ```go
   return func() (bool, string) {
       if len(hpPods) > 0 {
           return true, fmt.Sprintf("Skipping because %d pod(s) use hugepages", len(hpPods))
       }
       return false, ""
   }
   ```

3. **Side effects** – None.  
   The function only reads from `env`; it does not modify the environment or any global state.

## Dependencies

| Dependency | Role |
|------------|------|
| `GetHugepagesPods` | Retrieves hugepage‑using pods; used to decide skip status. |
| `len` | Determines if any such pods exist. |

No global variables are accessed, and the only exported global defined in this file is unrelated (`AbortTrigger`).

## Usage Context

- The test runner calls `GetNoHugepagesPodsSkipFn(env)` when setting up a **no‑hugepages** compliance check.
- The returned function is invoked during test execution; if it returns `(true, msg)`, the framework records the test as *skipped* with the provided message.

This pattern allows tests to be dynamically skipped based on runtime cluster state without hard‑coding skip conditions.
