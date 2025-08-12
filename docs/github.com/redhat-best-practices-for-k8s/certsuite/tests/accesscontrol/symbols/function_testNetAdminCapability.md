testNetAdminCapability`

| Item | Detail |
|------|--------|
| **Package** | `accesscontrol` – tests for Kubernetes access‑control rules. |
| **Signature** | `func testNetAdminCapability(c *checksdb.Check, env *provider.TestEnvironment)` |
| **Exported?** | No (unexported helper used only inside the test suite). |

### Purpose
`testNetAdminCapability` verifies that a user lacking the `NET_ADMIN` capability is correctly denied when attempting network‑related operations. It is one of several small “check” functions invoked by the overall Access Control test runner.

The function receives:

* **c** – a pointer to a `checksdb.Check` instance, which represents a single test case in the suite.  
  *It is used only to store the result (`SetResult`) after the check finishes.*

* **env** – a pointer to `provider.TestEnvironment`, providing context such as the Kubernetes client and logger.

### Workflow
1. **Logging** – `GetLogger()` obtains a structured logger from `env` for diagnostic output.
2. **Capability Test** – The core logic is delegated to `checkForbiddenCapability`.  
   *This helper encapsulates the actual test of whether the subject can perform network‑admin actions; it returns a boolean indicating pass/fail and an optional error.*  
3. **Result Handling** –  
   * If `checkForbiddenCapability` reports failure, `SetResult(c, false)` marks the check as failed.  
   * On success, the function leaves the result unchanged (default is true).  
4. **No side‑effects** beyond setting the check status and optional log output.

### Dependencies
| Dependency | Role |
|------------|------|
| `checkForbiddenCapability` | Performs the capability check; this function merely forwards its outcome. |
| `GetLogger` | Provides a logger for debugging; no state mutation. |
| `SetResult` | Persists the pass/fail status on the `Check`. |

### How it fits the package
Within `accesscontrol/suite.go`, a series of such helper functions are defined, each targeting a different Kubernetes privilege (e.g., `NET_ADMIN`, `SYS_TIME`). The test runner iterates over all checks in the database and calls the corresponding function. Thus, `testNetAdminCapability` is a small, focused unit that contributes to the overall validation of the cluster’s RBAC enforcement.

### Example Mermaid Flow
```mermaid
flowchart TD
    A[Start] --> B{Get Logger}
    B --> C[Invoke checkForbiddenCapability]
    C -- Pass --> D[SetResult(true) – implicit]
    C -- Fail --> E[SetResult(false)]
    D & E --> F[End]
```

### Summary
- **What**: Tests that a user cannot gain network‑admin privileges.  
- **Inputs**: A `Check` record and test environment.  
- **Outputs**: Result stored in the `Check`.  
- **Side‑effects**: Logging, no state changes elsewhere.  
- **Package role**: One of many capability checks executed by the Access Control suite.
