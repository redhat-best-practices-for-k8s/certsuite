GetNoOperatorCrdsSkipFn`

```go
func GetNoOperatorCrdsSkipFn(env *provider.TestEnvironment) func() (bool, string)
```

## Purpose

`GetNoOperatorCrdsSkipFn` is a helper that produces a **skip‑function** for tests that should only run when the test environment contains **no Operator Custom Resource Definitions (CRDs)**.  
The returned function can be passed to `t.SkipNow()` or similar mechanisms in the test suite.

## Parameters

| Name | Type | Description |
|------|------|-------------|
| `env` | `*provider.TestEnvironment` | A pointer to the current test environment. The environment holds a list of CRDs that are present on the cluster (via `env.CustomResourceDefinitions`). |

## Return Value

A closure with signature `func() (bool, string)`:

- **First return value (`bool`)** – `true` indicates the test should be skipped.
- **Second return value (`string`)** – a human‑readable message that explains why the skip is happening.

The function checks if `env.CustomResourceDefinitions` contains any entries.  
If the list is empty, the test should be skipped because it relies on operator CRDs that are not present.

## Key Implementation Details

```go
func GetNoOperatorCrdsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
    return func() (bool, string) {
        if len(env.CustomResourceDefinitions) == 0 {
            return true,
                "Skipping test: no operator CRDs present in the environment."
        }
        return false, ""
    }
}
```

- The closure captures `env` from its surrounding scope.
- It uses Go’s built‑in `len()` to determine if any CRDs are defined.
- No external packages or global variables are referenced.

## Side Effects

The function itself has no side effects.  
Its only effect is to provide a predicate that can be used by test code to decide whether to skip the current test.

## Usage Context in the Package

`GetNoOperatorCrdsSkipFn` lives in `pkg/testhelper`.  
Other test helpers (e.g., `GetMissingOperatorCRDsSkipFn`) generate similar predicates for different conditions.  
Typical usage:

```go
skipFn := GetNoOperatorCrdsSkipFn(env)
if skip, msg := skipFn(); skip {
    t.Skip(msg)
}
```

This pattern keeps tests concise and centralizes the logic for determining when operator‑dependent tests should be omitted.

---

**Summary:**  
`GetNoOperatorCrdsSkipFn` returns a closure that tells test runners to skip tests if the current environment contains no Operator CRDs. It’s a small, dependency‑free utility used throughout the test suite to guard against missing operator resources.
