testPodRequests`

**File:** `suite.go` – line 960  
**Package:** `accesscontrol`  
**Visibility:** unexported (used only inside the test suite)

---

### Purpose
`testPodRequests` evaluates every container in all pods that are under test to ensure each one has *resource requests* defined.  
The function collects compliant and non‑compliant containers, logs details, builds a report object for each container, and finally updates the compliance check result.

### Signature
```go
func (*checksdb.Check, *provider.TestEnvironment)()
```
| Parameter | Type                    | Role |
|-----------|------------------------|------|
| `c`       | `*checksdb.Check`      | The current compliance check being evaluated.  It holds the context needed to set a pass/fail status and store detailed results. |
| `env`     | `*provider.TestEnvironment` | Provides access to the test environment, including the list of pods and containers that need inspection. |

> **Note** – Both parameters are passed by reference; the function mutates only the `Check` object (via `SetResult`) and writes log output.

### Key Steps

| Step | Action | Dependencies |
|------|--------|--------------|
| 1 | Log start message using the logger obtained from `GetLogger`. | `LogInfo`, `GetLogger` |
| 2 | Iterate over all containers in the environment (`env`). For each container: <ul><li>Check if resource requests are set via `HasRequestsSet`. </li><li>If compliant, create a `ContainerReportObject` with status “passed” and append to the list of compliant objects.</li></ul> | `HasRequestsSet`, `NewContainerReportObject`, `append` |
| 3 | For containers without requests: create a `ContainerReportObject` with status “failed”, log an error, and add it to the non‑compliant list. | `LogError`, `NewContainerReportObject`, `append` |
| 4 | After scanning all containers, call `SetResult` on the check object to mark the overall test as passed or failed depending on whether any non‑compliant containers were found. | `SetResult` |

### Dependencies

- **Logging utilities** (`LogInfo`, `LogError`) – provide visibility into which containers pass/fail.
- **Resource request checker** (`HasRequestsSet`) – encapsulates the logic for determining if a container’s resource requests are defined.
- **Reporting helpers** (`NewContainerReportObject`) – constructs a structured report entry that can be consumed by the test framework or output tooling.
- **Check object methods** (`SetResult`) – finalizes the result of this particular compliance check.

### Side‑Effects

1. **Logging** – Emits informational and error messages for each container processed.
2. **State mutation** – Updates the supplied `checksdb.Check` with a pass/fail status and attaches detailed report objects to it.
3. No external resources are altered; the function is read‑only regarding the test environment.

### Placement in the Package

`testPodRequests` is one of several internal helper functions used by the `accesscontrol` test suite to validate Kubernetes object configurations.  
It is invoked from higher‑level test orchestration code (likely inside a Ginkgo/Go test runner) where a specific compliance check (e.g., “All containers must have resource requests”) is instantiated, and this function performs the actual inspection logic.

---

#### Suggested Mermaid Flow

```mermaid
flowchart TD
  A[Start] --> B{Iterate over containers}
  B -->|Has requests| C[Create passed report]
  B -->|No requests| D[Log error & create failed report]
  C --> E[Append to compliant list]
  D --> F[Append to non‑compliant list]
  E & F --> G{Any failures?}
  G -- Yes --> H[SetResult(failed)]
  G -- No --> I[SetResult(passed)]
  H & I --> J[End]
```

This diagram illustrates the decision path and how compliance status is determined.
