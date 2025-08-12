isResourceInRoleRule`

| Aspect | Detail |
|--------|--------|
| **Package** | `rbac` – part of the CertSuite test utilities for Role/ClusterRole analysis |
| **Signature** | `func isResourceInRoleRule(resource CrdResource, rule RoleRule) bool` |
| **Visibility** | Unexported – used only within this package’s RBAC helper logic |

### Purpose

Determines whether a given Kubernetes Custom Resource Definition (CRD) resource matches the scope of a *role rule*.

A `RoleRule` contains:
- A list of API groups (`Groups []string`)
- A list of plural resource names (`Resources []string`)

The function checks that **both** the group and plural form of the supplied `CrdResource` appear in the corresponding lists of the `RoleRule`. If they do, the CRD is considered covered by that rule.

### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `resource` | `CrdResource` | The resource to test. It must expose at least two fields: `Group string` and `Plural string`. |
| `rule` | `RoleRule` | The role rule against which the resource is evaluated. |

### Return Value

- `bool`:  
  * `true` – the CRD’s group **and** plural are present in the rule.  
  * `false` – at least one of them is missing.

### Key Operations & Dependencies

1. **String Splitting** – uses Go’s standard library `strings.Split` to split the rule’s `Resources` list into individual resource names if they were comma‑separated (this helper call is hidden inside the function body).
2. **Membership Checks** – simple linear scans (`for _, g := range rule.Groups`) to see whether `resource.Group` and `resource.Plural` are present.

No external packages or global state are involved; the function is pure aside from reading its arguments.

### Side‑Effects

None. The function only inspects its inputs and returns a boolean.

### How It Fits in the Package

The RBAC package provides utilities for examining Kubernetes Role/ClusterRole definitions during tests.  
`isResourceInRoleRule` is used by higher‑level helpers that:

1. Iterate over all rules of a role.
2. For each rule, check if a CRD is governed by it.
3. Aggregate results to determine whether the role grants access to particular custom resources.

Because it is unexported, callers remain within the package’s RBAC logic; external code relies on exported wrappers or higher‑level functions that internally invoke this helper.
