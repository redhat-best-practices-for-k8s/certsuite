PermissionsHaveBadRule`

| Item | Details |
|------|---------|
| **Purpose** | Determines whether any of the supplied *StrategyDeploymentPermissions* contain an invalid or disallowed rule. In the context of the CertSuite operator tests, this helper is used to assert that generated RBAC objects adhere to the policy defined by the operator. |
| **Signature** | `func PermissionsHaveBadRule([]v1alpha1.StrategyDeploymentPermissions) bool` |
| **Parameters** | - `permissions []v1alpha1.StrategyDeploymentPermissions` – a slice of permissions objects, each representing the rules that a strategy deployment is allowed to use. The type originates from the operator’s API package (`github.com/redhat-best-practices-for-k8s/certsuite/api/v1alpha1`). |
| **Return value** | `bool` – `true` if at least one permission contains a rule that violates the expected policy; otherwise `false`. |
| **Key dependencies** | *None* – the function performs only in‑memory checks on the provided slice. It relies solely on the exported fields of `StrategyDeploymentPermissions`; no external packages are imported beyond the type definition itself. |
| **Side effects** | None. The function is pure: it does not modify its arguments or interact with global state. |
| **Typical usage pattern** | ```go\nif access.PermissionsHaveBadRule(perms) {\n    t.Fatalf(\"operator produced disallowed permissions\")\n}\n``` This guard ensures that test cases fail early when the operator generates RBAC rules that do not match the expected policy. |
| **Where it fits in the package** | The `access` package contains utilities used by the operator tests to validate access control configurations. `PermissionsHaveBadRule` is a small, focused helper that encapsulates the logic for detecting policy violations, keeping test code concise and readable. |

### Suggested Mermaid diagram (optional)

```mermaid
flowchart TD
    A[Operator Test] --> B{Calls}
    B --> C[access.PermissionsHaveBadRule(perms)]
    C --> D{Result}
    D -->|true| E[Fail Test]
    D -->|false| F[Continue]
```

This diagram visualises the typical call flow from a test case to the helper and back.
