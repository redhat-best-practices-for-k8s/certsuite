GetNoIstioSkipFn`

| Item | Detail |
|------|--------|
| **Package** | `testhelper` (github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper) |
| **Exported?** | Yes |
| **Signature** | `func(*provider.TestEnvironment) func() (bool, string)` |

### Purpose
Creates a *skip function* that decides whether the current test should be skipped because **Istio is not installed** in the environment.  
The returned closure can be invoked repeatedly during a test run; it returns:

| Return value | Meaning |
|--------------|---------|
| `true`  | Skip the test. |
| `false` | Do not skip – proceed with the test. |

When skipping, the string explains why (currently `"Istio not found"`).

### Inputs
* `env *provider.TestEnvironment`: a reference to the current test environment.
  * The function uses only the `TestEnvironment` pointer; it does **not** inspect any fields of it in this code snippet.

### Output
A closure `func() (bool, string)` that can be called by the test harness.  
The closure holds no state beyond what is captured from the argument (`env`), and therefore has no side‑effects.

### Key Dependencies & Side Effects
* **No external calls** – the function body contains only a literal closure definition.
* **Globals** – none of the globals (e.g., `AbortTrigger`) are used here.
* **Return value** – always returns `true, "Istio not found"`; no mutation occurs.

### How it fits the package
`testhelper` supplies helper utilities for test execution.  
`GetNoIstioSkipFn` is one such utility that allows a test to be conditionally skipped when Istio is missing from the cluster.  
It can be used in tests as:

```go
skip := GetNoIstioSkipFn(env)
if skip, msg := skip(); skip {
    t.Skip(msg)
}
```

This keeps the test logic simple while delegating the environment check to a reusable helper.

### Diagram (Mermaid)

```mermaid
flowchart TD
  A[Test] -->|calls| B[GetNoIstioSkipFn(env)]
  B --> C{Return closure}
  C --> D[Closure called]
  D --> E{Skip?}
  E -- true --> F[t.Skip("Istio not found")]
  E -- false --> G[Continue test]
```

*The function never returns `false`; it is a static “always skip” helper for the specific condition of missing Istio.*
