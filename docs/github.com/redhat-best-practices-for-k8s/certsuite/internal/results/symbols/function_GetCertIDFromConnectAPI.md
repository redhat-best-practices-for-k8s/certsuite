GetCertIDFromConnectAPI`

| Feature | Details |
|---------|---------|
| **Package** | `results` (`github.com/redhat-best-practices-for-k8s/certsuite/internal/results`) |
| **Exported?** | ✅ (public) |
| **Signature** | `func GetCertIDFromConnectAPI(certName, certVersion, orgID, apiKey, endpoint string) (string, error)` |

### Purpose
`GetCertIDFromConnectAPI` retrieves the unique certification ID assigned by Red Hat Connect for a specific certification name and version.  
The function constructs a REST call to the Connect API, sends it through any configured proxy, decodes the JSON response, extracts the `id` field, and returns that value.

### Parameters

| Name | Type   | Role |
|------|--------|------|
| `certName`     | `string` | Certification name (e.g., “Red Hat OpenShift Container Platform”) |
| `certVersion`  | `string` | Version string of the certification (e.g., “4.10”) |
| `orgID`        | `string` | Organization identifier used by Connect |
| `apiKey`       | `string` | API key for authentication |
| `endpoint`     | `string` | Base URL of the Connect service |

### Return Values

| Value | Type   | Meaning |
|-------|--------|---------|
| first  | `string` | The certification ID returned by the API. Empty string on failure. |
| second | `error`  | Non‑nil if any step (URL building, request creation, proxy handling, network I/O, JSON decoding) fails. |

### Key Dependencies

| Dependency | How it’s used |
|------------|---------------|
| `log.Info`, `log.Debug` | Log progress and debug information at various stages. |
| `strings.ReplaceAll` | Build the API URL by substituting placeholders with actual values. |
| `fmt.Sprintf` | Format strings for log messages and error formatting. |
| `http.NewRequest` & `bytes.NewBuffer` | Create an HTTP POST request containing JSON payload (`org_id`, `api_key`). |
| `setProxy` (internal helper) | Apply proxy configuration to the request if required. |
| `sendRequest` (internal helper) | Perform the HTTP round‑trip and return the response body or error. |
| `json.NewDecoder` / `Decode` | Parse the JSON payload returned by Connect. |

### Algorithm Overview

```text
1. Log start of operation.
2. Replace placeholders in the endpoint URL with certName, certVersion, orgID, apiKey.
3. Create a POST request with a minimal JSON body containing org_id and api_key.
4. Set headers: Content-Type=application/json, Accept=application/json.
5. Apply proxy settings via setProxy().
6. Send request using sendRequest(); read response body.
7. Decode JSON into map[string]interface{}.
8. Extract "id" field; if missing, return error.
9. Log success and return the ID.
```

### Side Effects

* Logs messages at `Info`/`Debug` level – no state mutation beyond that.
* No global variables are modified or read (aside from configuration accessed by helper functions).
* The function itself is pure with respect to input/output; all I/O occurs through the HTTP client.

### How it Fits the Package

The `results` package focuses on gathering and formatting test results.  
`GetCertIDFromConnectAPI` supplies the unique certification identifier needed when submitting results back to Red Hat Connect, ensuring that subsequent API calls reference the correct certification instance.

--- 

#### Mermaid Flow Diagram (suggested)

```mermaid
flowchart TD
  A[Start] --> B{Build URL}
  B --> C[Create POST request]
  C --> D[Set Headers]
  D --> E[Apply Proxy via setProxy()]
  E --> F[sendRequest() -> Response]
  F --> G{Decode JSON}
  G --> H[Extract "id"]
  H --> I[Return ID / Error]
```

*Use the diagram in README or internal docs to illustrate the request‑response lifecycle.*
