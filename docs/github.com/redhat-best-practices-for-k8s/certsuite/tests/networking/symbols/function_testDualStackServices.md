testDualStackServices`

| Item | Details |
|------|---------|
| **Package** | `networking` – part of the CertSuite test harness for Kubernetes networking |
| **Visibility** | Unexported (`func testDualStackServices`) |
| **Signature** | `func(*checksdb.Check, *provider.TestEnvironment)` |

### Purpose
Runs a suite‑level check that verifies services deployed with both IPv4 and IPv6 addresses can be reached from the control plane node.  
The function is called by the framework’s test runner after all other per‑test setup has been performed.

### Inputs

| Parameter | Type | Meaning |
|-----------|------|---------|
| `check` | `*checksdb.Check` | The current check record (contains ID, status, etc.). The function updates this object with results. |
| `env` | `*provider.TestEnvironment` | Context that provides access to the Kubernetes cluster and helper utilities such as logging and reporting. |

### Core Logic

1. **Logging** – Uses `LogInfo` to announce start of dual‑stack service validation.
2. **IP Version Detection** – Calls `GetServiceIPVersion(env)` to determine whether services are IPv4, IPv6 or both.
3. **Error Handling** – On failure, logs via `LogError`, creates a report object (`NewReportObject`) with relevant fields and records the error in `check`.
4. **Reporting** – For each service:
   - Builds a slice of strings representing IP addresses that should be reachable.
   - Creates a `report` object, attaches metadata (e.g., `"ip_version"`, `"service_name"`).
   - Calls `SetResult` on the check to mark it successful or failed.
5. **Success Path** – If all services respond as expected, logs success and sets the check result accordingly.

### Dependencies

| Dependency | Role |
|------------|------|
| `LogInfo`, `LogError` | Test‑environment logging helpers. |
| `GetServiceIPVersion` | Determines if the cluster exposes dual‑stack endpoints. |
| `NewReportObject` | Constructs a structured report for the check result. |
| `SetResult` | Persists the final outcome in the `check` object. |

### Side Effects

* Mutates the supplied `check` with status and detailed results.
* Emits log messages to the test harness.
* Generates one or more report objects that are attached to the check.

### Integration Context
This function is part of a collection of test helpers that validate Kubernetes networking features.  
It is executed as a **suite‑level** check (hence its signature lacks a `t *testing.T` parameter) and relies on the global `env` variable defined in `suite.go`. The surrounding test suite sets up services with dual‑stack IPs, then invokes `testDualStackServices` to confirm connectivity.

### Suggested Mermaid Flow

```mermaid
flowchart TD
    A[Start] --> B{LogInfo}
    B --> C[GetServiceIPVersion]
    C --> D{Success?}
    D -- Yes --> E[Build IP list]
    E --> F[NewReportObject]
    F --> G[SetResult(Success)]
    D -- No --> H[LogError]
    H --> I[NewReportObject(Error)]
    I --> J[SetResult(Failure)]
    G --> K[End]
    J --> K
```

This diagram captures the decision point around IP‑version detection and shows how reports are created for both success and failure paths.
