testPodClusterRoleBindings`

| Aspect | Detail |
|--------|--------|
| **Location** | `tests/accesscontrol/suite.go:654` |
| **Signature** | `func testPodClusterRoleBindings(c *checksdb.Check, env *provider.TestEnvironment)` |
| **Visibility** | Unexported – used only inside the test suite. |

### Purpose

The function verifies that a pod does **not** use any Cluster‑Level RoleBinding (i.e., it should be confined to namespace‑scoped RBAC). It is executed as part of a larger set of access‑control checks during a CertSuite run.

It performs three distinct validation steps:

1. **ClusterRoleBinding usage check** – uses `IsUsingClusterRoleBinding` to detect if the pod’s ServiceAccount has any cluster‑wide bindings.
2. **Ownership by a cluster‑wide operator** – determines whether the pod is owned by an operator that runs with cluster‑scoped privileges (`ownedByClusterWideOperator`).  
3. **Special case for default ServiceAccount** – pods using the `default` SA are exempted from the check.

Each step creates a *report object* via `NewPodReportObject`, logs relevant information, and sets the result on the supplied `checksdb.Check`.

### Inputs & Outputs

| Parameter | Type | Description |
|-----------|------|-------------|
| `c` | `*checksdb.Check` | The check instance to which results will be attached. |
| `env` | `*provider.TestEnvironment` | Provides access to pod metadata (e.g., owner references, namespace). |

The function **does not return** a value; it mutates the supplied `Check` by adding report entries and setting the final result.

### Key Dependencies

| Dependency | Role |
|------------|------|
| `GetLogger()` | Retrieves a logger scoped to the current check. |
| `LogInfo`, `LogError` | Emit informational or error logs. |
| `IsUsingClusterRoleBinding(c, env)` | Returns a bool indicating if any cluster‑role bindings are used. |
| `GetTopOwner(env)` | Fetches the top‑level owner reference for the pod. |
| `ownedByClusterWideOperator(owner)` | Determines if that owner is a known cluster‑wide operator. |
| `NewPodReportObject(name, description)` | Builds a report object to be appended to the check’s results. |
| `SetResult(result)` | Finalizes the check result (e.g., `ResultPass`, `ResultFail`). |

### Side Effects

* Adds multiple `report.Pod` entries to the `Check` for each validation step.
* Logs information at various levels, potentially affecting test output verbosity.
* Calls `SetResult` only once – the last evaluation in the function dictates the overall outcome.

### Integration with the Package

In the `accesscontrol` test suite, this function is invoked as part of a broader check routine that iterates over all pods in the cluster. It complements other checks such as:

* **Pod security policy compliance** (`testPodSecurityPolicies`)
* **ServiceAccount usage** (`testPodDefaultSAUsage`)

Together they form the *RBAC‑related validation* set ensuring that workloads run with minimal privileges.

### Suggested Mermaid Diagram

```mermaid
flowchart TD
  A[Start] --> B{Is Using ClusterRoleBinding?}
  B -- Yes --> C[Report Fail]
  B -- No --> D{Owned by cluster‑wide operator?}
  D -- Yes --> E[Report Pass (exempt)]
  D -- No --> F[Check default ServiceAccount]
  F -- Default SA --> G[Report Pass]
  F -- Other SA --> H[Report Fail]
```

This diagram captures the decision flow of `testPodClusterRoleBindings`.
