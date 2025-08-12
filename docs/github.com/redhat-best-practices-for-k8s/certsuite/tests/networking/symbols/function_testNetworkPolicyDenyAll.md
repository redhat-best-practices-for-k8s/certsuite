testNetworkPolicyDenyAll`

**Location**

`tests/networking/suite.go:333`

---

### Purpose
Verifies that a *deny‑all* NetworkPolicy correctly blocks traffic between pods in the same namespace while still allowing intra‑pod communication (e.g., `localhost`).  
The test is run as part of the broader networking test suite and contributes to the final compliance report stored in a `checksdb.Check` object.

### Inputs

| Parameter | Type                         | Description |
|-----------|------------------------------|-------------|
| `c`       | `*checksdb.Check`            | The check record that will be updated with the results of this test. |
| `env`     | `*provider.TestEnvironment` | Test environment providing access to cluster resources (pods, services, namespaces). |

### Key Steps & Dependencies

1. **Logging**  
   - `LogInfo` is used throughout to describe progress and outcomes.
   - Errors are reported with `LogError`.

2. **Policy Verification**  
   - `IsNetworkPolicyCompliant(env, nodePort)` checks whether the deny‑all policy is applied on the target namespace.  
   - The helper `LabelsMatch` confirms that the selected pods match the intended label selector.

3. **Ping Test Setup**  
   - Pods are created in two stages:
     1. *Local* pod (within the same pod) – verifies intra‑pod connectivity.
     2. *Remote* pod (in a different namespace or with differing labels) – ensures traffic is blocked.
   - Each ping operation is wrapped by `NewPodReportObject`, which captures success/failure and logs details.

4. **Result Aggregation**  
   - Successful pings are appended to the check’s result slice via `append`.
   - After all checks, `SetResult(c)` finalises the outcome in the database record.

### Output

The function does not return a value; it mutates the supplied `*checksdb.Check` object:

- **Success** – If the deny‑all policy is enforced and ping attempts behave as expected (local success, remote failure).
- **Failure** – Any deviation triggers `LogError` entries and leaves the check in an error state.

### Integration

`testNetworkPolicyDenyAll` is invoked by the networking test suite’s `BeforeEach` hook (`beforeEachFn`). It relies on:

- Global variables `env` (the test environment) and `c` (check record).
- The package‑level constants `defaultNumPings` and `nodePort`, which control ping counts and service port selection.

This function is a building block for the overall compliance assessment of NetworkPolicies within CertSuite.
