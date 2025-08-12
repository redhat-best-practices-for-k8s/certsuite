LoadChecks` – Networking Test Suite Registration

## Purpose
`LoadChecks` registers a set of **network‑connectivity** checks with the CertSuite testing framework.  
It is called once during test suite initialization to create a group of checks, configure per‑check metadata (IDs, labels, skip logic) and provide a common *BeforeEach* hook that runs before each check.

> **Note:** The function signature `func()()` is intentionally empty – it returns an anonymous cleanup function that does nothing.  This matches the pattern used by other suites in CertSuite.

## Inputs / Outputs
| Direction | Type | Description |
|-----------|------|-------------|
| **Input** | *none* | No parameters; relies on package‑level globals (`env`, `beforeEachFn`). |
| **Output** | `func()` | A no‑op cleanup function that satisfies the framework’s expected return type. |

## Key Steps

1. **Logging & Hook Setup**
   ```go
   Debug("LoadChecks")
   WithBeforeEachFn(beforeEachFn)
   ```
   - Logs entry into the function.
   - Registers a *BeforeEach* hook (`beforeEachFn`) that will run before every check in this group.

2. **Create Check Group**
   ```go
   grp := NewChecksGroup()
   ```
   A new `ChecksGroup` is instantiated to hold all checks for this suite.

3. **Add Connectivity Checks (Repeated 4×)**
   Each iteration creates a *check* that calls the shared helper `testNetworkConnectivity`.  
   The same helper is reused for four distinct test scenarios:

   | Scenario | Skip Logic |
   |----------|------------|
   | No containers under test | `GetNoContainersUnderTestSkipFn()` |
   | DaemonSet failed to spawn | `GetDaemonSetFailedToSpawnSkipFn()` |
   | No pods under test | `GetNoPodsUnderTestSkipFn()` |
   | Normal connectivity | none |

   For each scenario the check is constructed as:
   ```go
   grp.Add(
       WithCheckFn(testNetworkConnectivity),
       WithSkipCheckFn(skipFn),
       NewCheck(GetTestIDAndLabels()),
   )
   ```

4. **Return**
   The function returns an empty closure: `return func() {}`.

## Dependencies

| Called Function | Purpose |
|-----------------|---------|
| `Debug` | Framework logger. |
| `WithBeforeEachFn` | Register global *BeforeEach* hook. |
| `NewChecksGroup`, `Add`, `WithCheckFn`, `WithSkipCheckFn`, `NewCheck`, `GetTestIDAndLabels` | Build and register checks in the group. |
| Skip helpers (`GetNoContainersUnderTestSkipFn`, etc.) | Provide conditions under which a check should be skipped. |
| `testNetworkConnectivity` | The actual test logic executed by each check. |

These functions are part of CertSuite’s *testing* package and orchestrate how checks are discovered, configured, and executed.

## Package Context

The file lives in the **networking** test package (`github.com/redhat-best-practices-for-k8s/certsuite/tests/networking`).  
`LoadChecks` is invoked by the suite bootstrap code (typically `init()` or a `TestMain`) to register all networking checks.  The returned cleanup function is ignored because no teardown logic is required for these stateless tests.

## Mermaid Diagram (Optional)

```mermaid
graph TD
    A[LoadChecks] --> B[Debug]
    A --> C[WithBeforeEachFn]
    A --> D[NewChecksGroup] --> E[grp]
    E --> F{Add 4 checks}
    F --> G[Check1: testNetworkConnectivity, skipNoContainers]
    F --> H[Check2: testNetworkConnectivity, skipDaemonSetFail]
    F --> I[Check3: testNetworkConnectivity, skipNoPods]
    F --> J[Check4: testNetworkConnectivity, noSkip]
    E --> K[Return empty func()]
```

This visual shows the flow from function entry to check registration and final return.
