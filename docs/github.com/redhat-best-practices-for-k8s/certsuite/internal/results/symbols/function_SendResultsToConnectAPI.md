SendResultsToConnectAPI`

**Package:** `github.com/redhat-best-practices-for-k8s/certsuite/internal/results`  
**Visibility:** exported (`func SendResultsToConnectAPI(...) error`)  

---

## Purpose
Uploads a set of test results to the Red Hat Connect API.  
The function takes raw data and metadata, builds a multipart/form‑data HTTP request, sends it through an optional proxy, and processes the JSON response.

> The routine is intentionally long (`nolint:funlen`) because it orchestrates several low‑level steps (file handling, form building, network I/O) that are not reusable elsewhere in the package.

---

## Signature

```go
func SendResultsToConnectAPI(
    url string,
    token string,
    name string,
    version string,
    certName string,
    filePath string,
) error
```

| Parameter | Type   | Meaning |
|-----------|--------|---------|
| `url`     | `string` | Target Connect API endpoint (e.g. `https://connect.redhat.com/api/v1/results`). |
| `token`   | `string` | Bearer token for authentication. |
| `name`    | `string` | Name of the test suite / result set. |
| `version` | `string` | Version string of the test suite. |
| `certName`| `string` | The name of the certificate being tested (used in form field). |
| `filePath`| `string` | Path to a local file that will be uploaded as part of the payload (usually a tar‑gz archive). |

**Return value**

* `error`:  
  * `nil` on success.  
  * Non‑nil if any step fails: opening the file, building the request, sending it, or parsing the response.

---

## High‑level Flow

1. **Log start** – `Info("Sending results to Connect API…")`.
2. **Sanitise inputs** – all string parameters are run through `strings.ReplaceAll` to escape problematic characters (e.g., newlines).
3. **Create multipart writer** – `multipart.NewWriter`.  
   * Adds a hidden form field `"type"` set to `"results"`.
4. **Attach files**
   * Reads the file at `filePath` with `os.Open`, copies it into a form file part via `CreateFormFile("data")` and `io.Copy`.
5. **Add metadata fields** – name, version, certName using helper `createFormField`.
6. **Close writer** – finalises the multipart body.
7. **Build HTTP request**
   * `http.NewRequest("POST", url, &body)`  
   * Sets headers:  
     - `Authorization: Bearer <token>`  
     - `Content-Type` from `writer.FormDataContentType()`  
     - `Accept: application/json`
8. **Proxy handling** – optional proxy is set via `setProxy(req)`.
9. **Send request** – helper `sendRequest(req)` performs the actual `http.Client.Do`.
10. **Handle response**
    * Reads body, decodes JSON into a generic map with `json.NewDecoder`.  
    * Checks for HTTP status codes ≥ 400 and returns an error containing the server’s message.
11. **Log success** – prints the returned `"message"` field from the API if present.

---

## Dependencies & Helpers

| Helper | Purpose |
|--------|---------|
| `createFormField(writer, key, value)` | Adds a simple text form field to the multipart writer. |
| `setProxy(req)` | Configures a proxy for the request if an environment variable is set. |
| `sendRequest(req)` | Executes the HTTP request using a shared client and returns the response. |

The function also relies on package‑level constants:

* `writeFilePerms` – file permissions used when writing temporary files (not directly in this routine but part of the same package).
* `htmlResultsFileContent` – an embedded HTML template used elsewhere; not referenced here.

---

## Side Effects

* **Logging** – Uses `logrus.Info`, `Debug`, and `Errorf` to emit runtime information.  
* **Network I/O** – Sends a POST request over the network.
* **Filesystem** – Reads from `filePath`; does **not** modify or delete it.

---

## Where It Fits

Within the *results* package, this routine is the final step of the test‑reporting pipeline:

1. Tests run → results are generated locally (tar‑gz archives).  
2. These files are uploaded to the Connect API via `SendResultsToConnectAPI`.  
3. The API returns a status message that is logged for audit purposes.

It ties together the data generation code with external services, making it the critical integration point for reporting test outcomes.
