GetNoAffinityRequiredPodsSkipFn`

**Purpose**

`GetNoAffinityRequiredPodsSkipFn` is a higher‑order helper that produces a *skip function* for the CertSuite test framework.  
The returned closure can be used in tests to skip an entire check when **no pods exist that have no node‑affinity or pod‑affinity rules**.

In other words, if the environment contains only pods that declare affinity constraints, the test that relies on *pods without any affinity* is irrelevant and should be skipped. The skip function returns a boolean indicating whether to skip and an optional message explaining why.

---

## Signature

```go
func GetNoAffinityRequiredPodsSkipFn(env *provider.TestEnvironment) func() (bool, string)
```

| Parameter | Type                 |
|-----------|----------------------|
| `env`     | `*provider.TestEnvironment` – the current test environment, containing the cluster state. |

| Return type | Description |
|-------------|-------------|
| `func() (bool, string)` | A closure that when invoked returns: <br>• `true` if there are **zero** affinity‑free pods and the test should be skipped.<br>• `false` otherwise. The second return value is an explanatory message used in the skip log. |

---

## Dependencies

* **`GetAffinityRequiredPods(env)`** – a helper that returns all pods in the cluster which declare any node‑ or pod‑affinity rule.
* **`len()`** – standard library function to count elements.

No global variables are accessed directly; the function relies solely on its argument and the two helpers above.

---

## Implementation Flow

```go
func GetNoAffinityRequiredPodsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
    return func() (bool, string) {
        // 1. Retrieve all pods that *do* have affinity rules.
        affinityPods := GetAffinityRequiredPods(env)

        // 2. If the list is empty → no pods without affinity exist.
        if len(affinityPods) == 0 {
            return true, "No pods in the cluster are free of affinity constraints"
        }

        // 3. Otherwise proceed with the test.
        return false, ""
    }
}
```

1. **Collect** all affinity‑required pods from the environment.  
2. **Count** them.  
   * If the count is zero → there is no pod without affinity; the skip function signals a skip.  
   * If at least one such pod exists → the test should run normally.

---

## Usage Context

In CertSuite tests, you often want to guard checks that require pods without any affinity rules:

```go
skipFn := GetNoAffinityRequiredPodsSkipFn(env)
if skip, msg := skipFn(); skip {
    t.Skip(msg)          // Skip the entire check
}
```

This pattern keeps test suites concise and avoids false negatives when the cluster topology does not satisfy the precondition.

---

## Side‑Effects

* None – the function only reads from `env` and computes a count; it does not mutate state.
* The returned closure can be called multiple times with no side effects.

---
