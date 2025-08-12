LoadChecks`

```go
// nolint:funlen
func LoadChecks() func()
```

| Element | Description |
|---------|-------------|
| **Purpose** | Registers a suite of operator‑related tests with the test framework. It creates a group of checks, populates it with concrete test cases, and returns a function that will be executed by the testing harness. |
| **Inputs** | None (no parameters). |
| **Outputs** | A `func()` – the “before‑each” wrapper that Ginkgo/Gomega will invoke before each test run. The returned closure contains all registered checks but performs no logic on its own; it simply triggers the framework’s internal execution pipeline. |
| **Side effects** | * Logs a debug message (`Debug("Loading operator checks")`).<br>* Sets up a global `beforeEachFn` that will be called before each test (via `WithBeforeEachFn`).<br>* Instantiates a `ChecksGroup` and adds multiple `Check` objects to it. <br>* Each `Check` is configured with:<br>  * A descriptive ID/label via `GetTestIDAndLabels`. <br>  * An execution function such as `testOperatorInstallationPhaseSucceeded`. <br>  * Optional skip logic (`GetNoOperatorsSkipFn`, `GetNoOperatorCrdsSkipFn`). |
| **Key dependencies** | • `Debug` – logging helper.<br>• `WithBeforeEachFn` – registers a hook to run before each test.<br>• `NewChecksGroup` / `Add` – constructs the group container.<br>• `WithCheckFn`, `WithSkipCheckFn` – attach behaviour to checks.<br>• `NewCheck` – factory for individual checks.<br>• Test helpers (`GetTestIDAndLabels`, `GetNoOperatorsSkipFn`, etc.) that supply metadata and conditional skip logic.<br>• The actual test functions: `<testOperatorInstallationPhaseSucceeded>`, `<testOperatorInstallationAccessToSCC>`, `<testOperatorOlmSubscription>`, `<testOperatorSemanticVersioning>`, `<testOperatorCrdVersioning>`.<br>• Global variables `env` (type `provider.TestEnvironment`) and `beforeEachFn`. |
| **Package context** | Part of the `operator` test suite. The function is called during test initialization to ensure that all operator‑related checks are available to the framework. It does not perform any runtime assertions itself; it merely registers them. |

### Flow (Mermaid)

```mermaid
flowchart TD
    A[LoadChecks] --> B{Debug}
    B --> C[WithBeforeEachFn(beforeEachFn)]
    C --> D[NewChecksGroup("operator")]
    D --> E1[Add Check: Installation Phase Succeeded]
    D --> E2[Add Check: Access to SCC]
    D --> E3[Add Check: OLM Subscription]
    D --> E4[Add Check: Semantic Versioning]
    D --> E5[Add Check: CRD Versioning]
    E1-E5 --> F[Return func()]
```

### Summary

`LoadChecks` is the central registration point for operator‑related tests. It prepares a suite of checks, each with its own execution logic and optional skip conditions, attaches them to a shared group, and supplies a closure that the test runner will execute before each test iteration. This keeps the test definitions modular while ensuring they are all loaded once at package initialization time.
