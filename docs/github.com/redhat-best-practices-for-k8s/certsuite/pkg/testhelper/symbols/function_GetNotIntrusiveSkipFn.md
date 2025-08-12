GetNotIntrusiveSkipFn`

| Attribute | Value |
|-----------|-------|
| **Package** | `testhelper` (github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper) |
| **Exported** | Yes |
| **Signature** | `func(*provider.TestEnvironment)(func() (bool, string))` |
| **Position in source** | `/Users/deliedit/dev/certsuite/pkg/testhelper/testhelper.go:554` |

### Purpose
`GetNotIntrusiveSkipFn` returns a *closure* that can be used by test cases to decide whether they should skip execution because the environment is “intrusive”.  
The returned function follows the typical signature for a `skipFunc` in this codebase: it returns

1. `bool` – whether the test should be skipped.
2. `string` – a human‑readable reason (used by the test runner to report why the skip happened).

### Inputs
* `env *provider.TestEnvironment`: The current test environment instance.  
  - It contains configuration that indicates whether the tests are running in an intrusive mode or not.

### Outputs
A function with signature `func() (bool, string)` that:

* Calls `IsIntrusive(env)` to determine if the environment is marked as intrusive.
* If **intrusive** → returns `(true, "intrusive test")`.
* Otherwise → returns `(false, "")`.

### Dependencies
| Dependency | Role |
|------------|------|
| `provider.TestEnvironment` | Holds runtime flags and configuration for tests. |
| `IsIntrusive(*provider.TestEnvironment)` | Helper that checks the environment’s intrusive flag. |

The function only reads state; it has no side effects.

### How It Fits in the Package
* **Context** – The `testhelper` package supplies utilities for constructing test environments, generating skip conditions, and creating resource objects.
* **Usage pattern** – Test functions often need to decide at runtime whether they should run.  They call `GetNotIntrusiveSkipFn(env)` to obtain a skip function, then pass it to the testing framework or a helper that evaluates it before executing the test body.

```go
func TestSomething(t *testing.T) {
    env := provider.NewTestEnvironment()
    skipFn := GetNotIntrusiveSkipFn(env)
    if skip, reason := skipFn(); skip {
        t.Skip(reason)
    }
    // …test logic…
}
```

Thus, `GetNotIntrusiveSkipFn` is a small but essential glue that connects the environment configuration to test execution control.
