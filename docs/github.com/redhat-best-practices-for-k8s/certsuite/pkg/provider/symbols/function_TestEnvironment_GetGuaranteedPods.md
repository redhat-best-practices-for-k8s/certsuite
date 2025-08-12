## `GetGuaranteedPods`

| Item | Detail |
|------|--------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider` |
| **Exported** | ✅ (public API) |
| **Receiver type** | `TestEnvironment` – the struct that holds all state for a certsuite test run. |
| **Signature** | `func (env *TestEnvironment) GetGuaranteedPods() []*Pod` |

### Purpose

`GetGuaranteedPods` extracts from the current test environment only those pods that satisfy the *guaranteed* criteria defined by the method `IsPodGuaranteed`.  
In certsuite, a *guaranteed pod* is one whose resource requests and limits are set in such a way that it receives exclusive CPU and memory allocation (typically `requests == limits`). This function is used by other filters and metrics collectors to focus on pods that can be trusted to not be evicted or throttled during the test.

### Inputs / State Used

* **`env.Pods`** – the slice of all pods discovered in the environment.  
  (`TestEnvironment` stores a `[]*Pod` field, populated elsewhere by the discovery logic.)

No external arguments are required; the function relies solely on the receiver’s state.

### Output

A new slice of pointers to `Pod` containing only those pods for which `IsPodGuaranteed(pod)` returns `true`.  
The order of pods is preserved from the original list.

### Key Dependencies & Calls

| Dependency | Role |
|------------|------|
| `env.IsPodGuaranteed(*Pod) bool` | Predicate used to decide whether a pod qualifies. The actual logic lives in another file (`filters.go`) and may inspect container resource requests/limits, annotations, etc. |
| Go built‑in `append` | Used to accumulate the qualifying pods into a new slice. |

No other global variables or side effects are involved.

### Side Effects

* None – the function is pure with respect to its input state; it does not modify `env.Pods` or any other field.
* It may allocate a new slice (but reuses existing pod pointers).

### How it Fits in the Package

1. **Discovery → Filtering**: After all pods are collected, certsuite applies various filters. `GetGuaranteedPods` is one of those early filters that narrows down the set to high‑priority workloads.
2. **Metrics & Tests**: Subsequent components (e.g., metrics collectors or specific tests) call this method to obtain a list of pods they can safely assume will not be throttled or evicted, ensuring deterministic test results.
3. **Extensibility**: The `IsPodGuaranteed` predicate is central; changing its logic automatically changes what `GetGuaranteedPods` returns without touching this helper.

---

#### Example Usage

```go
env := NewTestEnvironment(...)
allPods := env.GetAllPods()          // hypothetical method that populates env.Pods
guaranteed := env.GetGuaranteedPods()
fmt.Printf("Found %d guaranteed pods out of %d total\n", len(guaranteed), len(allPods))
```

The function is intentionally straightforward to keep the filter chain fast and maintainable.
