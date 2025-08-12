SliceDifference`

| Aspect | Detail |
|--------|--------|
| **Package** | `rbac` (github.com/redhat‑best‑practices‑for‑k8s/certsuite/tests/common/rbac) |
| **Signature** | `func SliceDifference(s1 []RoleRule, s2 []RoleRule) []RoleRule` |
| **Exported?** | Yes |

### Purpose
`SliceDifference` computes the set difference between two slices of `RoleRule`.  
Given two input slices `s1` and `s2`, it returns a new slice containing all elements that appear in `s1` but *do not* appear in `s2`.

> **Why this is useful** – In RBAC tests, the function allows callers to identify role rules that have been added or removed between two policy states.

### Inputs
| Parameter | Type | Description |
|-----------|------|-------------|
| `s1` | `[]RoleRule` | The source slice from which differences are extracted. |
| `s2` | `[]RoleRule` | The reference slice against which membership is checked. |

> *Note:* `RoleRule` is a custom type defined elsewhere in the package; its equality semantics are those of Go’s struct comparison.

### Output
- A new slice `[]RoleRule` containing all elements present in `s1` but absent from `s2`.  
  The order matches the original appearance in `s1`.

### Key Dependencies & Operations
| Operation | Function | Description |
|-----------|----------|-------------|
| Length checks | `len(s1)`, `len(s2)` | Determine if early exit is possible. |
| Element comparison | `<==` (struct equality) | Used to decide membership in `s2`. |
| Result construction | `append(result, rule)` | Builds the difference slice incrementally. |

No external packages are imported beyond Go’s standard library; all logic is contained within this file.

### Side‑Effects
- **None** – The function only reads from its arguments and returns a new slice; it does not modify the input slices or any package globals.

### Integration in the Package
The `rbac` package provides utilities for handling RBAC objects in tests.  
`SliceDifference` is used by higher‑level test helpers to:

1. Detect changes between expected and actual role bindings.
2. Report added/removed rules when a policy drift occurs.

Because it operates purely on slices of `RoleRule`, the function can be reused across various test scenarios without side effects, making it an ideal helper for test assertions.
