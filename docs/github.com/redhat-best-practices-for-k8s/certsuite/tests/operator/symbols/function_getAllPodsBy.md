getAllPodsBy`

| Aspect | Details |
|--------|---------|
| **File** | `tests/operator/helper.go` (line 91) |
| **Visibility** | Unexported – used only inside the *operator* test suite. |
| **Signature** | `func getAllPodsBy(label string, pods []*provider.Pod) []*provider.Pod` |

### Purpose
Collects all Pods that contain a specific label key/value pair from an input slice.

The function is a tiny helper used by the operator‑tests to filter Kubernetes Pod objects based on a label selector (e.g. `"app.kubernetes.io/component": "cert-manager"`).  
It does **not** query the API server; it simply scans the slice that was previously obtained via `env.GetPods()` or similar.

### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `label`   | `string` | The full label key (including namespace prefix) to look for, e.g. `"app.kubernetes.io/name"` |
| `pods`    | `[]*provider.Pod` | A slice of Pod pointers that represent the current cluster state in the test environment. |

### Return Value

| Type | Description |
|------|-------------|
| `[]*provider.Pod` | A new slice containing only those Pods whose metadata contains the supplied label key with a non‑empty value. The order is preserved from the input slice. If no Pod matches, an empty slice is returned (never `nil`). |

### Key Dependencies

| Dependency | Role |
|------------|------|
| `append`   | Used to build the result slice; no other external calls are made. |
| `provider.Pod` | The test‑specific representation of a Kubernetes Pod used throughout the operator tests. |

### Side Effects & Invariants

* **No state mutation** – the input slice and its elements are read‑only.
* **Deterministic output** – given the same inputs, it always returns the same slice (order preserved).
* **Performance** – linear in the number of Pods; trivial overhead for test execution.

### How It Fits the Package

The `operator` package implements end‑to‑end tests for the CertSuite operator.  
During a test run, after creating or updating resources, the suite queries the test environment (`env`) to list all relevant Pods.  
`getAllPodsBy` is then used to isolate subsets of those Pods (e.g., cert-manager controller pods, webhook pods) before performing assertions on their status or logs.

A typical usage pattern:

```go
allPods := env.GetPods()                     // fetch every Pod in the test cluster
cmPods  := getAllPodsBy("app.kubernetes.io/name", allPods)
// now assert that cmPods contain expected cert‑manager components
```

This helper keeps the test code concise and focused on logic rather than label filtering boilerplate.
