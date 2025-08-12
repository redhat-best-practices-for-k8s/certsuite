GetNoOperatorsSkipFn`

**Signature**

```go
func GetNoOperatorsSkipFn(env *provider.TestEnvironment) func() (bool, string)
```

> **Returns** a function that can be used as the `skip` callback for a test.  
> The returned closure evaluates whether the current test environment has any operators installed.  
> If no operators are present it signals to skip the test and supplies an explanatory message.

---

### Purpose

Many tests in *certsuite* are only meaningful when at least one operator is running in the cluster (e.g., tests that validate operator status, CRDs, or operator‑specific resources).  
`GetNoOperatorsSkipFn` provides a reusable way to short‑circuit such tests when an empty operator list is detected.

---

### Inputs

| Parameter | Type | Description |
|-----------|------|-------------|
| `env` | `*provider.TestEnvironment` | A pointer to the test environment object that holds the current state of the cluster, including the slice of discovered operators. |

> **Note**: The function does not modify `env`; it only reads from it.

---

### Returned Closure

The returned function has the signature required by the testing framework:

```go
func() (bool, string)
```

* **First return value (`bool`)** – indicates whether the test should be skipped.
  * `true`  → skip the test.  
  * `false` → run the test.

* **Second return value (`string`)** – a message that explains why the test was skipped (or can be empty).

The closure captures the `env` variable, so it operates on the environment instance that existed when `GetNoOperatorsSkipFn` was called.

---

### Implementation Details

```go
func GetNoOperatorsSkipFn(env *provider.TestEnvironment) func() (bool, string) {
    return func() (bool, string) {
        // If no operators were discovered, skip the test.
        if len(env.Operators) == 0 {
            return true, "no operators present in the environment"
        }
        return false, ""
    }
}
```

* The only external dependency is `len`, used to count elements of the slice `env.Operators`.
* No global variables or side‑effects are involved.
* The function is **pure**: it depends solely on its input and returns deterministic output.

---

### Usage in Tests

```go
func TestOperatorCRDs(t *testing.T) {
    env := provider.NewTestEnvironment()
    // ... environment setup ...

    t.Run("Validate CRDs", func(t *testing.T) {
        if skip, msg := GetNoOperatorsSkipFn(env)(); skip {
            t.Skip(msg)
        }
        // test body that expects operators to exist
    })
}
```

This pattern keeps the test code clean and makes the intent explicit: “skip unless at least one operator is present.”

---

### Side‑Effects & Dependencies

| Category | Details |
|----------|---------|
| **Side‑effects** | None – only reads `env`. |
| **External packages** | Only standard library (`len`). |
| **Global state** | No globals are accessed. |

---

### Where It Fits the Package

`GetNoOperatorsSkipFn` lives in the `testhelper` package, which supplies utilities for building and executing tests against a Kubernetes environment.  
It complements other helpers that query the test environment (e.g., operator discovery, resource checks) by providing a simple skip‑logic abstraction.

---

### Mermaid Diagram (Optional)

```mermaid
flowchart TD
    A[Call GetNoOperatorsSkipFn(env)] --> B{Return closure}
    B --> C[Closure uses env.Operators]
    C --> D[If len==0 -> skip=true, msg="no operators present"]
    C --> E[Else skip=false, msg="" ]
```

---
