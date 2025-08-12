testNamespace`

| Aspect | Detail |
|--------|--------|
| **Package** | `accesscontrol` (tests) |
| **Signature** | `func testNamespace(*checksdb.Check, *provider.TestEnvironment)` |
| **Exported?** | No – used only inside the suite file. |

### Purpose
`testNamespace` validates that all Kubernetes namespaces in the test environment are compliant with two rules:

1. **No invalid prefixes** – every namespace name must not start with any string listed in `invalidNamespacePrefixes`.
2. **Correct Custom Resource (CR) mapping** – each CR defined in a namespace must belong to one of the namespaces under test; if a CR references an unknown namespace, it is reported.

The function generates a *report* that can be consumed by the rest of the test framework to determine pass/fail status for this check.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `c` | `*checksdb.Check` | The current check instance; used to record results and add nested reports. |
| `env` | `*provider.TestEnvironment` | Test environment holding all runtime information (namespaces, CRs, logger, etc.). |

### Workflow
1. **Logging** – the function starts by emitting a log message indicating that namespace checks are beginning (`LogInfo`).  
2. **Prefix Validation**  
   * Iterate over every namespace in `env.NamespaceMap`.  
   * For each name, call `HasPrefix(name, invalidNamespacePrefixes)`; if true, record an error with `LogError` and append a namespaced report object via `NewNamespacedReportObject`.  
3. **CR Namespace Validation**  
   * Call the helper `TestCrsNamespaces(env)`, which returns a slice of CRs that reference namespaces not present in `env.NamespaceMap`.  
   * If any such CRs exist, log an error and create report objects for each invalid CR. The number of invalid CRs is retrieved with `GetInvalidCRsNum()`.  
4. **Result Aggregation**  
   * The function counts all namespace‑level errors (`len`) and sets the overall result on the check via `c.SetResult(...)`.  
5. **Reporting** – All individual error reports are attached to the main check report, allowing callers to inspect which namespaces or CRs failed.

### Dependencies
| Dependency | Role |
|------------|------|
| `LogInfo`, `LogError` | Logging utility from the test framework. |
| `HasPrefix` | Helper that checks if a string starts with any of a list of prefixes. |
| `NewNamespacedReportObject`, `NewReportObject` | Factories for creating structured report entries (namespaced or generic). |
| `TestCrsNamespaces` | Computes CRs whose referenced namespace is not in the set under test. |
| `GetInvalidCRsNum` | Returns count of invalid CRs detected by `TestCrsNamespaces`. |

### Side‑Effects
* No mutation to global state – all data are read from `env`.
* Generates log output and updates the passed `Check` instance.
* Creates new report objects but does not modify the test environment.

### Context within Package
The `accesscontrol` package contains a suite of integration tests for Kubernetes cluster security.  
`testNamespace` is one of several check functions that are registered in the test suite (`beforeEachFn`). It focuses on namespace hygiene, ensuring that:

* Namespaces do not carry disallowed prefixes (e.g., system or privileged namespaces).
* Custom resources reference only legitimate namespaces.

Its results feed into the overall compliance report for a cluster. The function is intentionally non‑exported because it is invoked only by the suite’s orchestration logic.
