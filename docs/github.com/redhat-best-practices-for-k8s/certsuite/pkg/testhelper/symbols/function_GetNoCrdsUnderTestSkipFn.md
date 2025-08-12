GetNoCrdsUnderTestSkipFn`

```go
func GetNoCrdsUnderTestSkipFn(env *provider.TestEnvironment) func() (bool, string)
```

### Purpose
Creates a **skip‑function** that can be supplied to the testing framework to decide whether a test should be skipped when no Custom Resource Definitions (CRDs) are present in the current `TestEnvironment`.  
The function is used by tests that rely on CRDs being available; if the environment reports zero CRDs, the test is automatically skipped with an explanatory message.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `env` | `*provider.TestEnvironment` | The runtime environment that contains a list of CRDs currently deployed or expected. |

> **Note**: `TestEnvironment` (from the `provider` package) is expected to expose a field/attribute that holds the number of CRDs under test – typically a slice such as `CrdsUnderTest []string`.  
> The implementation only relies on the length of this collection (`len(env.CrdsUnderTest)`).

### Return Value
A **closure** with signature `func() (bool, string)`.  
When invoked:

| Return | Meaning |
|--------|---------|
| `true` | Indicates that the test should be skipped. |
| `false`| The test may proceed. |
| `string`| A human‑readable message explaining why the skip occurred.  When skipping, this is `"No CRDs under test"`; otherwise it is an empty string. |

### Implementation Details
```go
func GetNoCrdsUnderTestSkipFn(env *provider.TestEnvironment) func() (bool, string) {
    return func() (bool, string) {
        if len(env.CrdsUnderTest) == 0 {          // <‑ length check
            return true, "No CRDs under test"     // skip with reason
        }
        return false, ""                          // proceed normally
    }
}
```

* The function uses only the built‑in `len` to inspect the slice of CRDs.
* No side effects occur; it simply reads from `env`.
* It is safe for concurrent use because it captures a read‑only reference.

### How It Fits the Package

- **Package**: `github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper`
- **Role**: Provides helper utilities for test suites.  
  This particular helper allows tests that depend on CRDs to gracefully skip themselves in environments where those resources are not available, avoiding false failures.
- **Typical Usage**:
  ```go
  func TestMyCRD(t *testing.T) {
      env := provider.NewTestEnvironment()
      t.SkipIf(GetNoCrdsUnderTestSkipFn(env))
      // … rest of the test …
  }
  ```

The helper centralises the skip logic, making individual tests cleaner and ensuring consistent behaviour across the suite.
