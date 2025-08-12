SendClaimFileToCollector`

| Feature | Detail |
|---------|--------|
| **Package** | `collector` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/collector`) |
| **Exported?** | ✅ |
| **Signature** | `func SendClaimFileToCollector(collectorURL, claimID, certName, certVersion, certData string) error` |

### Purpose
Sends a single certificate‑claim file to the Collector service via an HTTP POST request.  
The function is used by CertSuite when it needs to persist a new claim or update an existing one on the remote Collector API.

### Parameters

| Name | Type   | Description |
|------|--------|-------------|
| `collectorURL` | `string` | Base URL of the Collector endpoint (e.g. `"https://collector.example.com/api/claims"`). |
| `claimID` | `string` | Unique identifier for the claim record that will be created or updated. |
| `certName` | `string` | Human‑readable name of the certificate being claimed. |
| `certVersion` | `string` | Semantic version of the certificate (e.g., `"1.2"`). |
| `certData` | `string` | Base64‑encoded or raw file contents that represent the claim file to be stored. |

> **Note** – The function does not perform any validation on these inputs; it forwards them verbatim to the Collector API.

### Return Value

- `error`:  
  *`nil`* when the HTTP request succeeds (status code 2xx).  
  Any non‑nil error indicates a failure to construct, send, or close the request, or an HTTP error response from the Collector service.

### Key Dependencies & Flow

```mermaid
flowchart TD
    A[SendClaimFileToCollector] --> B[createSendToCollectorPostRequest]
    B --> C[Do (HTTP client)]
    C --> D[Close()]
```

1. **`createSendToCollectorPostRequest`**  
   Builds an `*http.Request` with method `POST`, target URL constructed from `collectorURL` and `claimID`, and a JSON body containing the certificate details (`certName`, `certVersion`, `certData`). It also sets standard headers such as `Content-Type: application/json`.

2. **`Do`**  
   The function invokes the global HTTP client’s `Do` method to transmit the request. No custom timeout or retry logic is applied here; those are handled elsewhere in the package.

3. **`Close`**  
   After the response is received, its body is closed to free resources. If closing fails, that error is returned.

### Side Effects

- Sends a network request; therefore, it requires network connectivity and proper Collector API credentials (typically set via environment or higher‑level wrappers).
- Does **not** modify local state; all data is transmitted over HTTP.
- May return an error if the Collector responds with a non‑2xx status code.

### Integration Context

`SendClaimFileToCollector` is typically called from higher‑level orchestration logic that iterates over pending claims. It is one of several helper functions in `collector/collector.go` responsible for CRUD operations against the Collector service:

| Function | Responsibility |
|----------|----------------|
| `CreateClaim` | POST a new claim record (no file) |
| `UpdateClaimFile` | PATCH or PUT to add/update the claim file |
| **`SendClaimFileToCollector`** | POST the actual claim file content |

Because it is exported, external modules can use it directly for unit tests or custom integrations that need fine‑grained control over claim‑file uploads.
