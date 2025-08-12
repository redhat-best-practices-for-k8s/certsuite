RunNetworkingTests`

### Purpose
`RunNetworkingTests` orchestrates a collection of ICMP ping tests across multiple network attachments in the certsuite testing framework.

- It receives a **map** that associates each *network name* with a `NetTestContext`, which encapsulates the source and destination container references and any required metadata.
- For every entry, it runs a single‑target ping via the exported helper `TestPing`.
- The function aggregates failures per network, builds structured log reports, and returns two values:
  1. A map of network names to slices of IP addresses that failed to respond (`testhelper.FailureReasonOut`).
  2. A boolean indicating whether any failure was observed.

### Signature
```go
func RunNetworkingTests(
    netTestCtxMap map[string]netcommons.NetTestContext,
    maxRetries int,
    ipVersion netcommons.IPVersion,
    logger *log.Logger,
) (testhelper.FailureReasonOut, bool)
```

| Parameter | Type                               | Description |
|-----------|------------------------------------|-------------|
| `netTestCtxMap` | `map[string]NetTestContext` | Each key is a network name; the value contains source/destination containers and any required configuration. |
| `maxRetries`    | `int` | How many times to retry a ping on failure (currently unused but reserved for future logic). |
| `ipVersion`     | `netcommons.IPVersion` | Indicates IPv4 or IPv6 usage when invoking `TestPing`. |
| `logger`        | `*log.Logger` | Logger used for structured debug/info/error messages. |

### Key Steps & Dependencies

1. **Logging Setup**  
   Uses the supplied logger to emit structured debug entries (`Debug`, `Info`, `Error`) throughout execution.

2. **Iterate Over Networks**  
   For each network in `netTestCtxMap`:
   - Calls `PrintNetTestContextMap()` (utility for debugging).
   - Executes `TestPing(ctx, ipVersion)` once.
     - `TestPing` performs the actual ICMP ping between source and destination containers on that network.

3. **Collect Results**  
   - On success: logs an info message, appends the successful container pair to a report object (`NewContainerReportObject`) and continues.
   - On failure:
     - Records the failed IP address in `failedIPs`.
     - Adds an error entry to a new report object (`NewReportObject`).
     - Marks overall test status as failed.

4. **Return Values**  
   - The first return value (`FailureReasonOut`) is a map of network names → slice of failing IPs.
   - The second boolean indicates if any failure occurred.

### Side Effects
- Emits structured logs and reports but does not modify the input `netTestCtxMap`.
- Does **not** alter container state; only performs ping operations.

### Package Context
`RunNetworkingTests` is part of the `icmp` test package, which validates network connectivity in a Kubernetes environment.  
It relies on:

- `netcommons.NetTestContext`: container and networking metadata.
- `testhelper.FailureReasonOut`: failure aggregation type used across certsuite tests.
- Logging helpers (`Debug`, `Info`, `Error`) and report constructors.

Its output feeds into higher‑level test orchestration logic that aggregates failures across multiple test suites.
