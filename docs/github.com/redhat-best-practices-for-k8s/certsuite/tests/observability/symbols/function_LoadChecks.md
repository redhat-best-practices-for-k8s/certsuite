LoadChecks` – Observability test loader

### Purpose
`LoadChecks` registers a suite of observability checks with the testing framework.
It is called once per test run and builds a **check group** that will be executed
by the underlying test harness.

The function has no parameters and returns nothing (`func()`).  
Its only side‑effect is adding checks to the global check registry, which is
consumed by the `suite.go` entry point of the package.

### Key steps

| Step | Action |
|------|--------|
| 1 | Log a debug message that the checks are being loaded. |
| 2 | Create a new `ChecksGroup`. This group collects all individual checks and will be executed in order. |
| 3 | Add the **global before‑each** hook (`beforeEachFn`) to the group. It runs once before each check. |
| 4 | For every *observability* test (containers, CRDs, termination message policy, etc.) do: <br>• Create a `Check` with its name and description.<br>• Attach the specific test function (`testContainersLogging`, `testCrds`, …).<br>• Supply a skip function that decides if the check should run in the current environment (e.g. `GetNoContainersUnderTestSkipFn`).<br>• Tag the check with its TestID and labels via `GetTestIDAndLabels`. |
| 5 | Add the fully configured check to the group. |
| 6 | Return – the group is now registered for execution. |

### Dependencies

| Dependency | Role |
|------------|------|
| `Debug` | Logs debug information. |
| `WithBeforeEachFn` | Wraps a function so it runs before each check. |
| `NewChecksGroup` | Creates a container for multiple checks. |
| `Add` | Registers a check or hook to a group. |
| `WithCheckFn` | Sets the test function that implements the actual logic. |
| `WithSkipCheckFn` / `WithSkipModeAll` | Supplies conditions under which a check should be skipped. |
| `NewCheck` | Instantiates a new check object. |
| `GetTestIDAndLabels` | Adds metadata (IDs, labels) to a check for reporting and filtering. |
| Skip functions (`GetNoContainersUnderTestSkipFn`, `GetNoCrdsUnderTestSkipFn`, …) | Determine if the environment supports the required resources; if not, skip the test. |

### Global state used

* `beforeEachFn` – A function that is executed before each check (e.g., resetting shared data).
* `env` – The current testing environment (`provider.TestEnvironment`) used by skip functions to make decisions.

### Package context

The **observability** package implements a set of integration tests that verify the observability features of the CertSuite deployment.  
`LoadChecks` is the central registry function; it is called from the package’s `suite.go` file during test initialization, ensuring that all checks are available to the test runner.

---

#### Suggested Mermaid diagram (optional)

```mermaid
flowchart TD
    A[Start LoadChecks] --> B{Create ChecksGroup}
    B --> C[Add beforeEachFn]
    C --> D{For each check type}
    D --> E[NewCheck(name, desc)]
    E --> F[WithCheckFn(testFunc)]
    F --> G[WithSkipCheckFn(skipFunc)]
    G --> H[GetTestIDAndLabels()]
    H --> I[Add to group]
    I --> D
    D --> J[Finish LoadChecks]
```

This diagram visualizes the loop over all check types and how each is built and registered.
