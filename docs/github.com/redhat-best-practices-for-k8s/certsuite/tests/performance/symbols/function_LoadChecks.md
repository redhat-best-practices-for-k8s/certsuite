LoadChecks` – Central Test‑Group Builder

### Purpose
`LoadChecks` is the public entry point that constructs and registers all performance checks used by the **performance** test suite.  
It returns a function of type `func()` which, when invoked, will:

1. Configure a common *before‑each* hook (`beforeEachFn`) for every check.
2. Create logical groups of checks (`NewChecksGroup`).
3. Populate those groups with individual checks (`NewCheck` + `Add`).

The returned closure is typically called by the test harness (e.g., Ginkgo’s `BeforeSuite` or a custom runner) to register the suite.

### Signature
```go
func LoadChecks() func()
```
* **Input** – none.
* **Output** – a zero‑argument function that, when executed, performs the registration described above.

### Key Dependencies & Calls

| Called Function | Role |
|-----------------|------|
| `Debug` | Emits debug information about the current test environment (`env`). |
| `WithBeforeEachFn` | Attaches `beforeEachFn` to a check or group. |
| `NewChecksGroup` | Creates a named collection of checks that can be nested or executed together. |
| `Add` | Adds a `Check` to a group. |
| `WithCheckFn` / `WithSkipCheckFn` | Wraps the core check logic and optional skip condition. |
| `NewCheck` | Instantiates an individual test check with its ID, label set, and functions. |
| `GetTestIDAndLabels` | Provides metadata (ID & labels) for each check. |
| `GetNoPodsUnderTestSkipFn` | Skip‑logic when no pods are available under test. |
| `testExclusiveCPUPool`, `testRtAppsNoExecProbes`, etc. | Core logic that implements the specific performance checks. |

### How it Works – Step‑by‑Step

1. **Debug Logging**  
   The function first calls `Debug` to log the current environment (`env`). This is purely informational.

2. **Group Creation**  
   For each category of tests (e.g., CPU pool checks, runtime application checks) a new group is created with `NewChecksGroup`. Groups are named descriptively (not shown in the snippet but inferred from context).

3. **Check Registration**  
   Within each group, one or more checks are added:
   * A `Check` object is built via `NewCheck`, which requires:
     - An ID and label set (`GetTestIDAndLabels`).
     - The actual test function (e.g., `testExclusiveCPUPool`).
     - Optional skip logic (e.g., `skipIfNoGuaranteedPodContainersWithExclusiveCPUs`).
   * Each check is wrapped with `WithBeforeEachFn`, ensuring the shared setup (`beforeEachFn`) runs before it.

4. **Return Closure**  
   After all groups and checks are registered, `LoadChecks` returns a closure that will execute this entire process when called by the test framework.

### Side Effects

* **Global state mutation** – Registers checks with the global test registry; subsequent calls to the returned function would duplicate registrations unless guarded.
* **No I/O** – Aside from debug logging, it does not interact with external systems.  
* **Dependency on `env`** – The environment is read but not modified.

### Integration in the Package

`LoadChecks` lives in `suite.go` of the `performance` package and is the only exported function that builds the test suite. It is used by higher‑level orchestrators (e.g., a test runner or Ginkgo’s `BeforeSuite`) to load all performance checks into the execution pipeline.

```go
// Example usage in a test harness
func TestPerformance(t *testing.T) {
    t.Run("performance", LoadChecks())
}
```

This design keeps check definitions declarative and centralized, enabling easy addition of new tests or modification of existing ones without touching the runner code.
