testContainerPortNameFormat`

| Aspect | Detail |
|--------|--------|
| **Location** | `tests/manageability/suite.go:101` |
| **Signature** | `func(*checksdb.Check, *provider.TestEnvironment)()` |
| **Visibility** | Unexported (used only within the test suite) |

### Purpose
Validates that every container in the examined Kubernetes objects declares its ports using a name that follows the partner‑specific naming convention:

```
<protocol>[-<suffix>]
```

where `<protocol>` must be one of `grpc`, `grpc-web`, `http`, `http2`, `tcp`, or `udp`.  
The function updates the compliance check result to **pass** if all ports are compliant, otherwise it records each non‑compliant port and marks the check as **fail**.

### Inputs

| Parameter | Type | Description |
|-----------|------|-------------|
| `chk` | `*checksdb.Check` | The check record that will be updated with the outcome. |
| `env` | `*provider.TestEnvironment` | Test environment containing all Kubernetes objects under test (via `env.Objects`). |

### Workflow

1. **Log Debug** – records start of validation.
2. **Run Core Validation** – calls `containerPortNameFormatCheck(env)` which returns:
   * `compliant []string` – list of compliant container/port identifiers
   * `nonCompliant []string` – list of non‑compliant container/port identifiers
3. **Error Handling** – if the check function errors, log it and return early.
4. **Populate Report**  
   * For each compliant item: create a `ContainerReportObject`, add a field `"status":"PASS"`, and append to `chk.Compliance.Report`.
   * For each non‑compliant item: similarly create an object but with `"status":"FAIL"` and include the offending port name.
5. **Set Result** – call `chk.SetResult()` which finalises the check status (PASS if no failures, otherwise FAIL).

### Key Dependencies

| Dependency | Role |
|------------|------|
| `LogDebug`, `LogError`, `LogInfo` | Structured logging for debugging and error reporting. |
| `containerPortNameFormatCheck` | Core logic that inspects all containers’ ports against the naming rule. |
| `NewContainerReportObject`, `AddField` | Build per‑container compliance report entries. |
| `chk.SetResult()` | Finalise the check status based on collected data. |

### Side Effects

* Mutates the passed `Check` object by adding detailed reports and setting its overall result.
* No external state is altered; only logs are emitted.

### Package Context

The function lives in the **manageability** test package, which contains a suite of compliance checks for Kubernetes manifests. It is invoked as part of the test harness (likely within a `BeforeEach` or similar hook) to verify that container port naming conventions adhere to the partner specification before a cluster is considered compliant.

---  

**Mermaid diagram suggestion**

```mermaid
flowchart TD
    A[Start] --> B{containerPortNameFormatCheck(env)}
    B -->|ok| C[compliant, nonCompliant lists]
    B -->|err| D[LogError & exit]
    C --> E[Populate Report (PASS/FAIL)]
    E --> F[chk.SetResult()]
    F --> G[End]
```

This diagram visually captures the decision path and report construction performed by `testContainerPortNameFormat`.
