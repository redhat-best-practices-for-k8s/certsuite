testUnalteredBootParams`

| Attribute | Value |
|-----------|-------|
| **Package** | `platform` (`github.com/redhat-best-practices-for-k8s/certsuite/tests/platform`) |
| **Signature** | `func (*checksdb.Check, *provider.TestEnvironment)` |
| **Exported** | No (unexported helper used only in tests) |

### Purpose
`testUnalteredBootParams` validates that the boot parameters of a target node remain unchanged after a test run. It does this by:

1. Logging entry and exit points for traceability.
2. Invoking `TestBootParamsHelper`, which performs the actual comparison between expected and observed kernel‑boot arguments.
3. Building two `NodeReportObject`s – one for the **pre‑test** state and one for the **post‑test** state – each containing the boot parameters snapshot from the node.
4. Setting the test result based on whether the snapshots match.

This function is part of the test suite that checks platform‑level compliance, specifically that tests do not inadvertently modify a node’s kernel configuration.

### Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `c` | `*checksdb.Check` | The check record being executed; it receives the result of this helper. |
| `env` | `*provider.TestEnvironment` | Test environment context providing access to logging, node communication, and state persistence. |

### Return Value
None – the function records its outcome directly on the supplied `Check` object.

### Key Dependencies & Calls
| Called Function | Role |
|-----------------|------|
| `GetLogger()` | Retrieves a structured logger scoped to this test. |
| `LogInfo(...)` | Emits informational messages at start/end and during report construction. |
| `LogError(...)` | Logs any error encountered while creating reports. |
| `TestBootParamsHelper(env)` | Performs the actual boot‑parameter comparison; returns a boolean indicating success or failure. |
| `NewNodeReportObject()` | Constructs a new node‑report entry (used twice). |
| `AddField(..., ...)` | Adds key/value pairs to the report (`"boot_params"` and `"result"`). |
| `SetResult(c, result)` | Persists the boolean test outcome on the `Check`. |

### Side Effects
* **Logging** – Emits structured logs that can be traced back to this helper.
* **Report Creation** – Two node‑report objects are added to the test environment’s report list; they include boot‑parameter snapshots for before/after comparison.
* **Result Recording** – The check’s result is updated, affecting downstream test reporting.

### How It Fits Into the Package
The `platform` package contains integration tests that validate Kubernetes platform behavior. This helper is used by higher‑level test functions (e.g., `TestUnalteredBootParams`) to ensure that node boot parameters are not altered by operations performed during a test run. By abstracting the comparison logic and report generation into this function, the test suite keeps its core test flows concise while still providing detailed audit trails.

---

#### Suggested Mermaid Flow Diagram
```mermaid
flowchart TD
    A[Start] --> B{Get Logger}
    B --> C[LogInfo: "Enter testUnalteredBootParams"]
    C --> D[TestBootParamsHelper]
    D --> E{Result?}
    E -- true --> F[Create before report]
    E -- true --> G[Create after report]
    F & G --> H[SetResult(c, true)]
    E -- false --> I[LogError: "Failed to create report"]
    I --> J[SetResult(c, false)]
    H & J --> K[LogInfo: "Exit testUnalteredBootParams"]
    K --> L[End]
```
This diagram illustrates the control flow from logging, through helper execution, to report creation and result setting.
