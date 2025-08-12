GetNoNodesWithRealtimeKernelSkipFn`

**Package:** `github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper`  
**Signature**

```go
func GetNoNodesWithRealtimeKernelSkipFn(env *provider.TestEnvironment) func() (bool, string)
```

---

### Purpose

Creates a **skip function** that test frameworks can use to conditionally skip tests which require at least one node running a *realtime kernel*.  
When the returned closure is invoked it:

1. Checks whether any node in the current environment uses a realtime kernel.
2. If none are found, returns `true` (meaning “skip”) together with a human‑readable message explaining why.

This helper keeps test code clean: instead of embedding kernel checks everywhere, a single line can decide if a test should run.

---

### Inputs

| Parameter | Type                            | Description |
|-----------|---------------------------------|-------------|
| `env`     | `*provider.TestEnvironment`    | The test environment object that contains information about the cluster nodes. It is used to query node details via `IsRTKernel`. |

> **Note:** The function does *not* modify `env`; it only reads from it.

---

### Output

A closure with signature `func() (bool, string)`:

| Return | Meaning |
|--------|---------|
| `true` | Skip the test. |
| `false`| Run the test. |
| string | If skipped, a message such as `"No node using realtime kernel"`; otherwise an empty string. |

---

### Key Dependencies

- **`IsRTKernel`** – A helper that inspects a single node and returns whether it is running a realtime kernel.  
  The skip function calls this for each node in `env`.

```go
func IsRTKernel(node *v1.Node) bool { … }
```

---

### Side Effects

- None.  
  It only reads from the environment; no state changes or external calls occur.

---

### Usage Pattern

```go
// In a test file
skip := testhelper.GetNoNodesWithRealtimeKernelSkipFn(env)
if skip() {
    t.Skip(skipMessage) // e.g., t.Skip("Skipping: No node using realtime kernel")
}
```

The closure is inexpensive to create and cheap to call, making it suitable for many tests.

---

### Where It Fits the Package

`testhelper` provides utilities that simplify writing and executing tests against a Kubernetes environment.  
This function belongs in the “conditional skip” family, enabling tests to be aware of cluster capabilities without duplicating logic across test files.  

It is typically used by higher‑level test suites (e.g., compliance checks) that only make sense when at least one node has a realtime kernel.

---
