sendRequest` – Internal HTTP helper

**File:** `internal/results/rhconnect.go`  
**Package:** `results`

| Item | Details |
|------|---------|
| **Signature** | `func sendRequest(req *http.Request, client *http.Client) (*http.Response, error)` |
| **Visibility** | Unexported (used only inside the package) |

### Purpose
Centralises the logic for performing an HTTP request to Red Hat Connect (RHConnect).  
It logs the request/response details at debug level and returns the raw `*http.Response` or an error.

### Parameters
| Name | Type | Meaning |
|------|------|---------|
| `req`  | `*http.Request` | The HTTP request to send. It is expected to already contain method, URL, headers, body etc. |
| `client` | `*http.Client` | HTTP client used to execute the request. Allows callers to inject custom transport, timeouts or TLS settings. |

### Return values
| Name | Type | Meaning |
|------|------|---------|
| `(*http.Response)` | The response returned by `client.Do`. May be nil if an error occurs before a response is received. |
| `(error)` | Error describing why the request failed (e.g., network failure, non‑2xx status code handling). |

### Flow
1. **Debug log – request details**  
   ```go
   logger.Debug("Sending RHConnect request", "method", req.Method, "url", req.URL)
   ```
2. **Execute the request** (`client.Do`)  
   - If `Do` returns an error, it is wrapped with `Errorf` and returned.
3. **Debug log – response status**  
   ```go
   logger.Debug("Received RHConnect response", "status", resp.Status)
   ```
4. **Return** the response and any error.

### Dependencies
| Dependency | Usage |
|------------|-------|
| `logger.Debug` | Logs request/response info for troubleshooting. |
| `logger.Errorf` | Formats and logs errors before returning them. |
| `client.Do(req)` | Performs the actual network call. |

### Side‑effects
- Emits log entries (debug or error) – no state changes elsewhere.
- Does **not** close the response body; callers must do so.

### Context within the package
`results` orchestrates communication with RHConnect to upload test results.  
`sendRequest` is a helper used by higher‑level functions that build specific endpoints, set headers (e.g., `Authorization`), and marshal payloads. It isolates HTTP execution so that unit tests can mock the client or validate logging behaviour without duplicating request logic.

```mermaid
flowchart TD
    A[Build *http.Request] -->|call| B(sendRequest)
    B --> C{client.Do(req)}
    C -- success --> D[Return resp]
    C -- error --> E[Log & return err]
```

> **Note:** The function assumes that the caller handles response body consumption and cleanup.  
---
