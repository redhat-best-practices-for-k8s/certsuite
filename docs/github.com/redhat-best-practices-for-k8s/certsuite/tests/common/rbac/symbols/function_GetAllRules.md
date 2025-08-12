GetAllRules`

**Location**

`/Users/deliedit/dev/certsuite/tests/common/rbac/roles.go:57`

```go
func GetAllRules(role *rbacv1.Role) []RoleRule
```

### Purpose

`GetAllRules` extracts every permission rule defined in a Kubernetes `Role`
object and returns them as a slice of the package‑specific type
`RoleRule`.  It is used by tests that need to examine all rules without
respecting the role’s name or namespace – essentially flattening the
role into a simple list.

### Inputs

| Parameter | Type            | Description |
|-----------|-----------------|-------------|
| `role`    | `*rbacv1.Role`  | A pointer to a Kubernetes Role resource. The function expects this value to be non‑nil; if nil, it will panic (normal Go behaviour when dereferencing). |

### Output

| Return value | Type        | Description |
|--------------|-------------|-------------|
| `[]RoleRule` | Slice of `RoleRule` | Each element represents one rule from the role’s policy. The slice is ordered in the same way as the rules appear in the original Role object. |

> **Note**: `RoleRule` is a lightweight struct defined elsewhere in the package that mirrors the fields needed for testing (e.g., `APIGroups`, `Resources`, `Verbs`). It does not include metadata such as `Namespace` or `Name`.

### Key Dependencies

* **`rbacv1.Role`** – The Kubernetes Role type from `k8s.io/api/rbac/v1`.  
  *The function accesses `role.Rules` to iterate over each `PolicyRule`.*
* **Standard library `append`** – Used to build the result slice.

No external packages or global state are touched; the function is pure
with respect to its input.

### Side Effects

None. The function only reads from the supplied `Role`; it does not modify
the role, any global variables, or produce I/O.

### How It Fits in the Package

The `rbac` package provides utilities for working with RBAC objects during
tests.  `GetAllRules` is a helper that normalises a Role into a simple
list of rules so that other test helpers (e.g., rule‑matching or
validation functions) can operate without caring about the role’s
metadata.

Typical usage pattern:

```go
role := &rbacv1.Role{ /* ... */ }
rules := rbac.GetAllRules(role)
// now `rules` can be passed to assertion helpers
```

Because it is exported (`GetAllRules`), it is intended for use by test
code outside the package as well.
