testOnlySingleOrMultiNamespacedOperatorsAllowedInTenantNamespaces`

**File**: `suite.go` (line 142)  
**Package**: `operator` – a test suite for validating operator deployments in a multi‑tenant environment.

### Purpose
The function validates that **single‑namespace** or **multi‑namespace** operators are installed **only** within the tenant’s dedicated “operator” namespace.  
In other words, it ensures that no such operator appears in any other namespace belonging to the same tenant.

### Signature
```go
func testOnlySingleOrMultiNamespacedOperatorsAllowedInTenantNamespaces(
    check *checksdb.Check,
    env   *provider.TestEnvironment,
)
```
* `check` – a record describing what is being validated; used to store the result.
* `env` – the runtime environment that exposes all namespaces, operators, and utility helpers needed for validation.

### High‑level workflow

| Step | Action |
|------|--------|
| 1 | Log start of the test (`LogInfo`). |
| 2 | Create a **map** `operatorNamespacesByTenant` to collect namespaces where operators are installed per tenant. |
| 3 | For each operator (single or multi‑namespace) found in the cluster: |
|   | * Verify it is installed in its dedicated namespace using `checkValidOperatorInstallation`. |
|   | * Record the namespace under the tenant’s entry in the map. |
| 4 | After scanning all operators, iterate over every tenant to check that **no** other namespaces contain operators for that tenant. |
|   | For each offending namespace: create a report object (`NewNamespacedReportObject`) and add it to `check`. |
| 5 | If any violations are found, set the overall test result to *failed* using `SetResult`; otherwise leave it as *passed*. |

### Key dependencies

| Dependency | Role |
|------------|------|
| `LogInfo` / `LogError` | Structured logging for debugging and audit trails. |
| `checkValidOperatorInstallation` | Helper that confirms an operator’s namespace matches the expected tenant‑dedicated “operator” namespace. |
| `NewNamespacedReportObject` | Builds a detailed report entry tied to a specific namespace, used when violations are detected. |
| `SetResult` | Finalizes the test status (pass/fail). |

### Interaction with other parts of the package

* The function is part of the **operator** test suite (`suite.go`).  
  It is invoked by the generic test runner that iterates over all checks defined in `checksdb`.
* It relies on the global variable `env` to enumerate tenant namespaces and operators; this environment is set up during `beforeEachFn`.
* The function contributes to a **report** that will be emitted at the end of the test run, indicating which tenants violated the namespace rule.

### Side‑effects

* Mutates the supplied `check` object by adding namespaced report objects for each violation.
* Sets the overall result flag on `check`.  
  No other global state is altered; all operations are read‑only with respect to the cluster.

---

#### Suggested Mermaid diagram (optional)

```mermaid
flowchart TD
    A[Start Test] --> B[Create map tenant→namespaces]
    B --> C{For each operator}
    C -->|valid| D[Record namespace]
    C -->|invalid| E[Log error, continue]
    D --> C
    C --> F[Check tenants for extra namespaces]
    F --> G{Violation?}
    G -- Yes --> H[Add report object]
    G -- No  --> I[Mark check passed]
    H --> F
    I --> J[SetResult(passed)]
```

This function is a focused validator ensuring namespace isolation for operators within a multi‑tenant Kubernetes cluster.
