addClaimFileToPostRequest`

**Package:** `collector`  
**Location:** `pkg/collector/collector.go:11`  

---

#### Purpose
Adds a file to an HTTP multipart POST request body.

The function opens the file whose path is supplied as the second argument, streams its contents into a new form‑data part created by the provided `multipart.Writer`, and then closes the file.  
It is used internally when the collector needs to attach a certificate/claim file to a submission endpoint.

---

#### Signature
```go
func addClaimFileToPostRequest(writer *multipart.Writer, filePath string) error
```

| Parameter | Type              | Description                                |
|-----------|-------------------|--------------------------------------------|
| `writer`  | `*multipart.Writer` | The multipart writer that builds the request body. |
| `filePath`| `string`          | Full path to the file that should be uploaded. |

**Returns**

- `error`:  
  *Non‑nil* if any I/O operation fails (opening, reading, copying, or closing the file).  
  *Nil* when the file has been successfully added.

---

#### Key Dependencies
| Dependency | Role |
|------------|------|
| `os.Open` | Opens the source file for reading. |
| `multipart.Writer.CreateFormFile` | Creates a new form-data part with the appropriate headers. |
| `io.Copy` | Streams the file contents into the multipart part. |
| `file.Close` | Ensures the file descriptor is released. |

---

#### Side Effects
- Writes to the provided `multipart.Writer`, extending the request body.
- Does **not** close the writer; that responsibility lies with the caller.

---

#### Usage Flow (pseudo‑code)
```go
var buf bytes.Buffer
w := multipart.NewWriter(&buf)

if err := addClaimFileToPostRequest(w, "/path/to/claim.pem"); err != nil {
    // handle error
}
w.Close()          // finish writing multipart body
req, _ := http.NewRequest("POST", url, &buf)
req.Header.Set("Content-Type", w.FormDataContentType())
```

---

#### Integration with the Package

`collector` orchestrates uploading claims to a remote service.  
This helper is invoked by higher‑level functions that build multipart requests (e.g., `submitClaims`).  
By abstracting file attachment logic, it keeps request construction concise and testable.

---
