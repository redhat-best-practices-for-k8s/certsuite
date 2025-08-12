testSYSNiceRealtimeCapability`

**Package**: `accesscontrol`  
**File**: `suite.go` (line 812)  

### Purpose
The function verifies that every container running on a node with a *realtime* kernel has the Linux capability **`SYS_NICE`**.  
This is a compliance check for security hardening – containers without this capability are considered non‑compliant.

### Signature

```go
func testSYSNiceRealtimeCapability(
    chk *checksdb.Check,
    env *provider.TestEnvironment,
) ()
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `chk` | `*checksdb.Check` | The check instance that will receive the result (`PASS/FAIL`) and a list of report objects. |
| `env` | `*provider.TestEnvironment` | Test environment providing access to node information, containers, etc. |

The function does **not** return any value; it mutates the `chk` argument.

### Key Steps

1. **Node kernel check**  
   - Calls `IsRTKernel(env)` to determine if the test node is running a realtime kernel.  
   - If not, the check is skipped (no objects added, result remains unset).

2. **Iterate over containers**  
   For each container on the node:
   * Retrieve its capabilities.
   * Use `isContainerCapabilitySet(container, "SYS_NICE")` to test presence of the capability.

3. **Report construction**  
   - On success: create a `ContainerReportObject` via `NewContainerReportObject`, add it to the *compliant* slice.
   - On failure: log an error (`LogError`) and add a similar object to the *non‑compliant* slice.

4. **Set check result**  
   After processing all containers, `chk.SetResult()` is called with the compiled compliant/non‑compliant lists.  

### Dependencies & Side Effects

| Dependency | Role |
|------------|------|
| `LogInfo`, `LogError` | Emit informational or error logs; side effect: console/structured logging. |
| `IsRTKernel(env)` | Returns a boolean indicating realtime kernel status. |
| `NewContainerReportObject` | Builds a report entry for each container. |
| `isContainerCapabilitySet(container, capability)` | Checks whether a capability is present in a container’s spec. |
| `chk.SetResult()` | Stores the final result and report slices inside the check object. |

No global variables are read or modified; all data flows through the passed arguments.

### How it Fits the Package

`testSYSNiceRealtimeCapability` is one of several *compliance* test functions in the **accesscontrol** suite.  
It follows the same pattern:

```go
func testXYZ(chk *checksdb.Check, env *provider.TestEnvironment) {
    // common logging
    // gather data from env
    // build report objects
    // chk.SetResult(...)
}
```

Thus it integrates seamlessly with the overall testing harness that iterates over all registered checks, collects results, and reports compliance status for a Kubernetes cluster.
