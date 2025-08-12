createSendToCollectorPostRequest`

| Item | Details |
|------|---------|
| **Package** | `collector` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/collector`) |
| **Location** | `/Users/deliedit/dev/certsuite/pkg/collector/collector.go:54` |
| **Signature** | `func createSendToCollectorPostRequest(url, token, claimFileName, claimFileContent, varFile string) (*http.Request, error)` |

### Purpose
Builds an HTTP `POST` request that uploads a *claim file* and optionally a *variable file* to the collector service. The resulting request is ready for dispatch by an HTTP client.

- **URL** – Endpoint of the collector service (e.g., `"https://collector.example.com/collect"`).
- **Token** – Bearer token used for authentication; added as `Authorization: Bearer <token>`.
- **claimFileName** – Filename to use when attaching the claim file.
- **claimFileContent** – Raw text of the claim file (YAML/JSON/etc.).
- **varFile** – Optional variable file content. If empty, no variable field is added.

The function returns either a fully‑constructed `*http.Request` or an error if any step fails (e.g., I/O on multipart writer).

### Key Dependencies
| Dependency | Role |
|------------|------|
| `http.NewRequest` | Creates the base request object. |
| `multipart.NewWriter` (`NewWriter`) | Builds a multipart/form‑data body. |
| `addClaimFileToPostRequest` | Adds the claim file part to the multipart writer. |
| `addVarFieldsToPostRequest` | Conditionally adds the variable file part. |
| `writer.Close()` | Finalizes the multipart payload and writes the terminating boundary. |
| `req.Header.Set` | Sets `Content-Type` (to the multipart boundary) and `Authorization`. |
| `FormDataContentType` | Returns the correct `Content-Type` header value for the multipart writer. |

> **Note**: The helper functions (`addClaimFileToPostRequest`, `addVarFieldsToPostRequest`) are defined elsewhere in the same package; they encapsulate the details of writing each part.

### Flow (pseudocode)
```go
func createSendToCollectorPostRequest(url, token, claimName, claimContent, varFile string) (*http.Request, error) {
    // 1. Create multipart writer
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    // 2. Attach the claim file
    if err := addClaimFileToPostRequest(writer, claimName, claimContent); err != nil { return nil, err }

    // 3. Optional variable file
    if varFile != "" {
        if err := addVarFieldsToPostRequest(writer, varFile); err != nil { return nil, err }
    }

    // 4. Close writer to finalize payload
    if err := writer.Close(); err != nil { return nil, err }

    // 5. Create HTTP request
    req, err := http.NewRequest(http.MethodPost, url, body)
    if err != nil { return nil, err }

    // 6. Set headers
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req.Header.Set("Authorization", "Bearer "+token)

    return req, nil
}
```

### Side Effects
- **I/O**: Writes to an in‑memory buffer (`bytes.Buffer`). No files are touched on disk.
- **State**: None. The function is pure aside from constructing the request object.

### How It Fits the Package
The `collector` package orchestrates communication with the CertSuite collector service.  
`createSendToCollectorPostRequest` is a low‑level helper that prepares the HTTP payload; higher‑level functions in the same package will call it, then use an `http.Client` to send the request and handle responses.

A typical usage pattern:

```go
req, err := createSendToCollectorPostRequest(endpoint, token, "claim.yaml", claimYAML, vars)
if err != nil { /* handle error */ }

resp, err := httpClient.Do(req)
```

This separation keeps request construction isolated from networking logic, simplifying testing and maintenance.
