testCrdRoles`

```go
func testCrdRoles(check *checksdb.Check, env *provider.TestEnvironment)
```

### Purpose  
`testCrdRoles` validates that each Kubernetes **role** (ClusterRole/Role) defined in the current environment applies only to Custom Resource Definitions (CRDs) that are part of the test suite. It ensures no role grants permissions on unrelated resources.

The function:

1. Retrieves all CRD resource names present in the cluster (`GetCrdResources`).
2. Enumerates every rule from every role (`GetAllRules`).
3. Filters out rules whose `APIGroups`, `Resources`, or verbs do **not** match any of the known CRDs.
4. Generates a compliance report that lists:
   - *Non‑compliant* rules (those referencing unknown resources)
   - *Compliant* rules (those only touching known CRDs)

If any non‑compliant rules are found, the check is marked `Failed`; otherwise it passes.

### Inputs  
| Parameter | Type | Description |
|-----------|------|-------------|
| `check` | `*checksdb.Check` | The compliance check instance being evaluated. The function will set its result and attach report objects to this instance. |
| `env`   | `*provider.TestEnvironment` | Context holding the test environment (e.g., Kubernetes client). It is used indirectly by helper functions (`GetCrdResources`, `GetAllRules`) which query the cluster via this env. |

### Output  
The function does **not** return a value. Its side effect is to mutate the supplied `check`:

- `SetResult(...)` is called with either `checksdb.CheckResultPass` or `checksdb.CheckResultFail`.
- Report objects (`NewNamespacedReportObject`, `NewNamespacedNamedReportObject`) are added via `AddField` calls, providing detailed lists of compliant/non‑compliant rules.

### Key Dependencies  

| Dependency | Role |
|------------|------|
| `GetCrdResources(env)` | Returns a slice of CRD names currently installed. |
| `GetAllRules()` | Retrieves all role rules across the cluster. |
| `FilterRulesNonMatchingResources(all, crds)` | Filters out rules that reference resources not in `crds`. |
| Report helpers (`NewNamespacedReportObject`, `AddField`) | Build structured report data for the check. |
| Logging (`LogInfo`, `LogError`) | Emit diagnostic messages during execution. |

### Side Effects  

* Mutates the passed `check` object.
* Generates and attaches report entries to `check`.
* Logs progress and errors via the package’s logger.

### Package Context  
The function resides in the **accesscontrol** test suite, which validates Kubernetes RBAC configurations against a set of expected rules. `testCrdRoles` is one of several internal helper functions invoked by higher‑level tests that orchestrate the overall compliance evaluation workflow. It ensures that roles are tightly scoped to the CRDs relevant for the certification process.
