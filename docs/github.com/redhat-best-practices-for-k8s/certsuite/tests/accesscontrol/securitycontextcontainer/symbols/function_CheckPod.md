CheckPod`

| Aspect | Detail |
|--------|--------|
| **Package** | `securitycontextcontainer` – a test helper that evaluates Kubernetes pod security context rules. |
| **Signature** | `func CheckPod(pod *provider.Pod) []PodListCategory` |
| **Exported?** | Yes |

---

### Purpose
`CheckPod` is the core evaluation routine used by the tests in this package.  
It takes a parsed Kubernetes Pod (from the test suite’s `provider` package), updates an internal representation of the pod’s container security context, and then produces a slice of `PodListCategory`. Each element represents one container inside the pod together with the category that describes whether its security settings are acceptable, partially accepted, or disallowed.

The function performs three main steps:

1. **Propagate Pod‑level defaults** – The security context defined on the pod is applied to all containers unless overridden.
2. **Apply container overrides** – Any per‑container security context values overwrite the pod‑level ones.
3. **Categorise each container** – Using `checkContainerCategory` (internal helper) and global category rules, it classifies the container into one of several categories (`Category1`, `Category2`, …). It also verifies that all requested volumes are allowed via `AllVolumeAllowed`.

---

### Inputs

| Parameter | Type | Description |
|-----------|------|-------------|
| `pod` | `*provider.Pod` | The parsed Kubernetes Pod object. This struct contains the pod spec, containers, and security context fields used by the tests. |

> **Note**: `CheckPod` assumes that `pod` is non‑nil; passing a nil pointer will cause a panic.

---

### Outputs

| Return value | Type | Description |
|--------------|------|-------------|
| `[]PodListCategory` | Slice of `PodListCategory` structs | Each element holds:
  - the container name,
  - its final security context after merging pod and container settings,
  - the category (`OK`, `NOK`, or a specific `CategoryID` value) determined by `checkContainerCategory`. |

The slice order matches the order of containers in the original pod spec.

---

### Key Dependencies

| Dependency | Role |
|------------|------|
| `AllVolumeAllowed(pod)` | Validates that every volume requested by the pod is permitted. If this check fails, all containers are marked as disallowed (`NOK`). |
| `checkContainerCategory(container *provider.Container) CategoryID` | Internal helper that inspects a single container’s security context and returns its category ID. It uses global constants such as `category2AddCapabilities`, `dropAll`, etc., to decide the category. |

These helpers are defined elsewhere in the same package; they rely on the global constants declared near the top of the file (`Category1`, `Category2`, …).

---

### Side Effects

* **No external mutation** – The function does not modify the input pod or any global state.
* **Internal caching** – It may create temporary structures (e.g., a slice of categories) but discards them before returning.

Because all data is passed by value (or pointers that are only read), `CheckPod` is pure from the perspective of callers.

---

### How it Fits the Package

The package implements a simplified *Security Context Constraints* checker for test purposes.  
- **Higher‑level**: Test cases construct `provider.Pod` objects and invoke `CheckPod`.  
- **Middle layer**: `CheckPod` orchestrates merging defaults, applying overrides, and delegating to category logic.  
- **Lower layer**: The helper functions (`checkContainerCategory`, `AllVolumeAllowed`) perform the actual policy checks using the global rule constants.

Thus, `CheckPod` is the public API that test cases use to assert whether a pod’s containers satisfy the expected security constraints.
