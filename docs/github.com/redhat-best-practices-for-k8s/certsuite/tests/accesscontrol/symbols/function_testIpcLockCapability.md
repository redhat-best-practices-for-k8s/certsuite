testIpcLockCapability`

### Purpose
`testIpcLockCapability` is a test helper that verifies whether the **IPC_LOCK** capability is correctly denied for unprivileged containers.  
It runs as part of the *access‑control* test suite and ensures that any container lacking explicit privilege will not be able to acquire the `IPC_LOCK` capability.

### Signature
```go
func testIpcLockCapability(check *checksdb.Check, env *provider.TestEnvironment) ()
```
- **check** – a pointer to a `Check` object from the internal database.  
  The function uses it to record the outcome of the test.
- **env** – the current test environment (`*provider.TestEnvironment`) which holds utilities such as logging and result setters.

### Key Operations
1. **Capability Check**  
   Calls `checkForbiddenCapability` (a helper defined elsewhere in the suite) with the string `"IPC_LOCK"`.  
   This function attempts to run a pod that requests the capability; it expects the pod creation to fail or the container to be killed, indicating that the capability is indeed forbidden.

2. **Logging**  
   Uses `GetLogger` from the environment to emit debug information about the test execution.  

3. **Result Recording**  
   Invokes `SetResult` on the passed `check` object to mark the check as *passed* or *failed* based on the outcome of `checkForbiddenCapability`.

### Dependencies
| Dependency | Role |
|------------|------|
| `checksdb.Check` | Holds test metadata and result state. |
| `provider.TestEnvironment` | Supplies logging (`GetLogger`) and result handling (`SetResult`). |
| `checkForbiddenCapability` | Core logic that attempts to use the capability and returns a boolean indicating success/failure. |

### Side‑Effects
- No global state is modified.
- The only observable side effect is updating the `Check.Result` field via `SetResult`.
- Logs are written through the environment’s logger.

### Integration into the Package
Within the **accesscontrol** package, tests are structured as functions with the signature `(check *checksdb.Check, env *provider.TestEnvironment)`.  
`testIpcLockCapability` is one such function registered in the suite (see `suite.go`).  
It fits the overall pattern of validating that each privileged capability listed in the policy is correctly blocked for unprivileged workloads.  

---

#### Suggested Mermaid Flow
```mermaid
flowchart TD
    A[Start] --> B{Get Logger}
    B --> C[Run checkForbiddenCapability("IPC_LOCK")]
    C --> D{Result?}
    D -- Pass --> E[SetResult(Passed)]
    D -- Fail --> F[SetResult(Failed)]
    E & F --> G[End]
```

This diagram illustrates the linear decision flow of `testIpcLockCapability`.
