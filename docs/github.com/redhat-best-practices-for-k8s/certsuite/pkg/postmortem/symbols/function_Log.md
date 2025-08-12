Log` – Collect and format a post‑mortem diagnostic snapshot

| Aspect | Details |
|--------|---------|
| **Signature** | `func Log() string` |
| **Visibility** | Exported (`Log`) – public API of the `postmortem` package. |
| **Purpose** | Generate a human‑readable log describing the current test environment and any errors that occurred during a run, then return it as a single string. It is intended to be invoked at the end of a test or when an error needs to be reported in a post‑mortem report. |

### Inputs / Outputs

| Parameter | Type | Notes |
|-----------|------|-------|
| None | – | The function pulls its data from the global *test environment* state via `GetTestEnvironment()`. |

| Return value | Type | Notes |
|--------------|------|-------|
| `string` | – | A formatted diagnostic string. It contains: <br>• A header with the test name and timestamp.<br>• The current error (if any).<br>• All key/value pairs from the environment that were set during the test run. |

### Key Dependencies

* **`GetTestEnvironment()`** – Reads the shared `testEnvironment` structure that holds per‑run data such as the test name, timestamp, errors, and custom fields.
* **`SetNeedsRefresh(true)`** – Signals that the cached environment snapshot must be refreshed before the next read. This is called immediately after obtaining the current environment to ensure subsequent reads reflect the latest state.
* **`Sprintf`, `String` (from Go's `fmt`)** – Used to format the final log string.

### Side Effects

1. Calls `SetNeedsRefresh(true)`, marking that the cached environment data should be refreshed on next access.  
2. No mutation of any global state other than the refresh flag; it purely reads and formats data.

### How It Fits the Package

The `postmortem` package provides tooling for generating diagnostic information after a test run or when an error occurs. `Log()` is the central helper that packages all relevant environment details into a single, readable string. Other parts of the suite (e.g., reporters or CI integrations) can call `postmortem.Log()` to obtain this snapshot and include it in logs, artifacts, or alert payloads.

```mermaid
flowchart TD
    A[Test Run] --> B{Error?}
    B -- Yes --> C[Collect Env via GetTestEnvironment]
    B -- No  --> D[Collect Env via GetTestEnvironment]
    C & D --> E[SetNeedsRefresh(true)]
    E --> F[Sprintf + String]
    F --> G[Return Log String]
```

> **Note:** The function assumes that `GetTestEnvironment()` and `SetNeedsRefresh()` are correctly implemented elsewhere in the package; otherwise, it will simply return an empty string.
