LoadChecks` – Overview

| Item | Detail |
|------|--------|
| **Package** | `accesscontrol` (github.com/redhat-best-practices-for-k8s/certsuite/tests/accesscontrol) |
| **Signature** | `func LoadChecks() func()` |
| **Exported?** | Yes |
| **Purpose** | Builds a collection of *checks* that will be executed against Kubernetes pods/containers to validate access‑control rules. The function returns a closure that, when invoked by the test harness, registers all checks with the global test environment (`env`). |

---

### 1. How it works

`LoadChecks` is a **factory** for the test suite’s check registry:

1. **Debugging hook**  
   `Debug()` is called to log that check registration has started (no side‑effects on the logic itself).

2. **Before‑each setup**  
   The returned closure will call `WithBeforeEachFn(beforeEachFn)` which attaches a common pre‑condition function (`beforeEachFn`) to every check group. This function runs before each individual test, usually preparing the environment or resetting state.

3. **Check groups** – `NewChecksGroup`  
   A *group* is created for each distinct type of access‑control rule (e.g., SCC checks, capability checks).  
   Each group receives:
   - a unique name (`"SCC"` / `"Capabilities"` etc.)
   - optional skip logic via `WithSkipCheckFn`.

4. **Adding individual checks** – `Add(NewCheck(...))`  
   For each check we call:
   ```go
   Add(
       NewCheck(
           GetTestIDAndLabels("some-test-id"),
           WithCheckFn(actualCheckFunc),
           WithSkipCheckFn(skipLogicFunc),
       ),
   )
   ```
   * `GetTestIDAndLabels` resolves a test identifier and attaches any labels (e.g., `"severity:high"`).
   * `WithCheckFn` supplies the actual function that will be executed during testing.
   * `WithSkipCheckFn` can prevent a check from running if prerequisites are not met (e.g., no containers under test).

5. **Special skip logic**  
   Functions like `GetNoContainersUnderTestSkipFn()` are used to avoid executing checks when the test context does not contain any relevant pods/containers.

6. **Return value**  
   The function returns a closure that, when called by the test harness, registers all these groups and checks with the global `env`. This design keeps registration separate from execution, allowing tests to control *when* checks are loaded.

---

### 2. Key dependencies

| Dependency | Role |
|------------|------|
| `Debug` | Logging during registration. |
| `WithBeforeEachFn`, `NewChecksGroup`, `Add`, `WithCheckFn`, `WithSkipCheckFn`, `NewCheck`, `GetTestIDAndLabels` | Builder pattern for test groups and checks. |
| `GetNoContainersUnderTestSkipFn` | Conditional skip logic. |
| **Specific check functions** (`testContainerSCC`, `testSysAdminCapability`, etc.) | The actual validation logic executed by each check. |

---

### 3. Side effects

* Registers check groups with the global test environment (`env`).  
* No state is mutated outside of this registration (e.g., it does not modify containers or Kubernetes resources).  
* Uses `beforeEachFn` to attach per‑test setup logic.

---

### 4. Where it fits in the package

The `accesscontrol` package implements end‑to‑end tests for Kubernetes security policies.  
`LoadChecks` is invoked by the test runner (likely through a `suite.go` bootstrap routine) to build the full set of checks that will be run against a cluster. Once loaded, the checks are executed as part of each test case, ensuring containers adhere to SCCs, capabilities, and other access controls.

---

### 5. Suggested Mermaid diagram

```mermaid
flowchart TD
    A[LoadChecks] --> B[Debug]
    A --> C{Return closure}
    C --> D[WithBeforeEachFn(beforeEachFn)]
    C --> E[NewChecksGroup("SCC")]
    E --> F[Add(NewCheck(...))]
    E --> G[WithSkipCheckFn(skipLogic)]
    C --> H[NewChecksGroup("Capabilities")]
    H --> I[Add(NewCheck(...))]
    subgraph Check Functions
        testContainerSCC()
        testSysAdminCapability()
        testNetAdminCapability()
        testNetRawCapability()
        testIpcLockCapability()
    end
```

This diagram shows the flow from `LoadChecks` to group creation and check registration.
