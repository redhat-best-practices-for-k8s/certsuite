GetNoContainersUnderTestSkipFn`

```go
func GetNoContainersUnderTestSkipFn(env *provider.TestEnvironment) func() (bool, string)
```

### Purpose  
`GetNoContainersUnderTestSkipFn` builds a **skip function** that can be passed to the test framework to conditionally skip tests when there are no containers marked for testing in the supplied `TestEnvironment`.  
The returned closure evaluates the environment on every call and reports:

| return value | meaning |
|--------------|---------|
| `true, msg`  | The test should be skipped; `msg` explains why. |
| `false, ""`  | No skip condition – the test may run. |

### Parameters

| Name | Type | Description |
|------|------|-------------|
| `env` | `*provider.TestEnvironment` | The current test environment that holds information about which containers are selected for testing. |

> **Note**  
> The exact fields of `TestEnvironment` used by the function are not visible in the snippet, but it relies on a slice or map that contains the “containers under test”. The code simply checks its length.

### Return Value

A closure with signature `func() (bool, string)` – the standard pattern for skip functions in this project.

### Key Dependencies

* **`len`** – used to inspect the size of the container list.
* **`provider.TestEnvironment`** – type from the `provider` package; provides the data needed for the check.

No external packages are imported beyond the standard library and the local `provider` package.

### Side Effects

None.  
The function only reads state from `env`; it does not modify any global or package‑level variables (e.g., `AbortTrigger` is unrelated).

### How It Fits in the Package

* The `testhelper` package supplies helper utilities for test authors.
* Skip functions are a common pattern used by the test harness to avoid running tests that cannot be evaluated under current conditions.
* This particular helper addresses the case where **no containers were selected** for testing, which would otherwise cause a test to fail or produce misleading results.

```mermaid
graph TD;
    TestFramework -->|calls skipFn| SkipFn;
    SkipFn -->|checks len(env.ContainersUnderTest)| env;
```

> **Usage Example**

```go
skip := GetNoContainersUnderTestSkipFn(env)
if skip() {
    t.Skip("no containers under test")
}
```

This keeps tests clean and readable while delegating the “empty‑environment” check to a reusable helper.
