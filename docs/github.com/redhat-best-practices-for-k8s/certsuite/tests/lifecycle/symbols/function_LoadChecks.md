LoadChecks` – Lifecycle Test Registry

| Aspect | Detail |
|--------|--------|
| **Package** | `lifecycle` (`github.com/redhat-best-practices-for-k8s/certsuite/tests/lifecycle`) |
| **Signature** | `func()()` |
| **Exported?** | Yes |
| **Linting directive** | `nolint:funlen` – function is intentionally long because it registers many checks. |

### Purpose
`LoadChecks` builds the *lifecycle* test group used by CertSuite’s e2e engine.  
It:

1. Instantiates a new `ChecksGroup` for the “lifecycle” category.
2. Adds a set of checks that exercise pod lifecycle hooks (`preStop`, `postStart`) and container behaviour (image policy, readiness probe).
3. Configures global *before‑each* logic to run once per test group.
4. Supplies skip functions that prevent running checks in unsuitable environments (e.g., when there are no containers or CRDs under test).

The function is called by the test harness during initialization; it returns a closure with no arguments, which is then invoked to register all checks.

### Inputs / Outputs
- **Input**: None – the function uses package‑level globals (`env`, `beforeEachFn`) and constants defined in this file.
- **Output**: A zero‑argument function that, when executed, registers a set of checks with the CertSuite engine.  
  The closure is expected to be called by the test runner; it does not return any value.

### Key Dependencies

| Dependency | Role |
|------------|------|
| `Debug` | Logs the start of the check registration for debugging purposes. |
| `WithBeforeEachFn` | Associates a function (`beforeEachFn`) that runs before each individual check in the group. |
| `NewChecksGroup` | Creates a new `ChecksGroup` instance for grouping related checks. |
| `Add`, `WithCheckFn`, `WithSkipCheckFn` | Builder methods on `ChecksGroup` used to attach checks and skip logic. |
| `NewCheck` | Instantiates a concrete check with metadata (ID, labels). |
| `GetTestIDAndLabels` | Provides standardized ID/label pairs for each test based on the function name and package context. |
| Skip helpers (`GetNoContainersUnderTestSkipFn`, `GetNoCrdsUnderTestSkipFn`, `GetNotIntrusiveSkipFn`) | Return predicates that skip checks when prerequisites are not met (e.g., no containers, no CRDs, or non‑intrusive mode). |
| Check functions (`testContainersPreStop`, `testScaleCrd`, etc.) | The actual test logic executed by each check. |

### Side Effects

- Registers the lifecycle checks with CertSuite’s global registry; this is a one‑time operation per test run.
- Sets a *before‑each* function that may perform environment preparation (e.g., resetting state) before each check.
- Uses package globals (`env`, `skipIfNoPodSetsetsUnderTest`) to influence skip logic.

### How It Fits the Package

The `lifecycle` package contains all tests related to Kubernetes pod lifecycle hooks and container behaviour.  
`LoadChecks` is the entry point that wires together:

1. **Metadata** – IDs, labels, and categories.
2. **Execution** – The actual test functions.
3. **Control flow** – Skipping logic based on the current test environment.

When CertSuite starts, it calls `LoadChecks()` for each test package to gather all checks before running them. This function keeps the registration logic in a single place, enabling maintainers to add or remove lifecycle tests by editing this file without touching the rest of the framework.
