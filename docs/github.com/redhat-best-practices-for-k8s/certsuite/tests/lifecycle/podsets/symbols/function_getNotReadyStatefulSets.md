getNotReadyStatefulSets`

**Package:** `podsets`  
**File:** `tests/lifecycle/podsets/podsets.go:162`  

### Purpose
Collects all StatefulSet objects that are **not yet ready** in the cluster.  
The function is a helper used by higher‑level waiting logic to determine which
StatefulSets still need to finish scaling, rolling updates, or other readiness
checks.

### Signature
```go
func getNotReadyStatefulSets(statefulSets []*provider.StatefulSet) []*provider.StatefulSet
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `statefulSets` | `[]*provider.StatefulSet` | Slice of StatefulSet objects retrieved from the API. |

| Return | Type | Description |
|--------|------|-------------|
| `[]*provider.StatefulSet` | Slice containing only those StatefulSets that are *not ready*. |

### Core Logic
1. **Iterate** over each StatefulSet in the input slice.
2. For every element, call `isStatefulSetReady(statefulSet)` to evaluate its readiness status.
3. If the function returns `false` (i.e., the set is not ready), append it to a result slice.
4. Log debug information using `Debug()` for each decision, including the stringified
   representation of the StatefulSet (`ToString(statefulSet)`).
5. Return the slice of non‑ready StatefulSets.

### Dependencies & Side Effects
| Dependency | Role |
|------------|------|
| `isStatefulSetReady` | Determines readiness; side‑effect free. |
| `Debug` | Logs internal decision steps (side effect: writes to test logs). |
| `ToString` | Serialises a StatefulSet for logging. |
| `append` | Standard Go slice operation; no external side effects. |

The function itself is **pure** apart from the logging calls – it does not modify
the input slice or any global state.

### Integration in the Package
- Used by higher‑level wait functions (e.g., `WaitForDeploymentSetReady`) to filter out
  StatefulSets that still require attention.
- Helps orchestrate test execution flow by providing a list of objects needing further
  readiness checks before tests can proceed.

### Example Usage
```go
allStatefulSets := fetchAllStatefulSets()
notReady := getNotReadyStatefulSets(allStatefulSets)
if len(notReady) > 0 {
    // Wait or retry logic here
}
```

---

**Note:** The function relies on the `provider.StatefulSet` type defined elsewhere in the test suite; its internal structure is not exposed here.
