updateTnf` – Web‑Server Utility for Updating Test‑Result Files

```go
func updateTnf(data []byte, r *RequestedData) []byte
```

| Aspect | Description |
|--------|-------------|
| **Purpose** | Transforms a raw byte slice (`data`) that represents the current state of a test‑result JSON file into an updated version incorporating information from `*RequestedData`. The function is invoked by the WebServer when a user submits changes via the UI. |
| **Inputs** | * `data` – raw bytes read from an existing `.tnf.json` (Test‑Result) file.<br>* `r` – pointer to a `RequestedData` struct that contains:<br>  * `OutputFolder` (string): target directory for output files.<br>  * `TnfID` (string): identifier of the test result. |
| **Outputs** | Updated byte slice containing the new JSON representation ready to be written back to disk. If an error occurs during marshaling/unmarshaling, the function logs a fatal message and terminates the process. |
| **Key Dependencies** | * `encoding/json` – used for `json.Unmarshal`/`Marshal`. <br>* `log.Fatal` – for fatal error handling (no recovery). <br>* `bytes.Buffer` (`buf`) – reused buffer to avoid allocations during marshaling. |
| **Side‑Effects** | * No global state is modified.<br>* On failure, the program exits via `log.Fatalf`, which stops the entire web server. |
| **How it fits in the package** | The WebServer exposes an HTTP endpoint that receives user edits (e.g., new test‑result fields). This handler reads the original file into a byte slice and passes it to `updateTnf`. The returned bytes are then written back to disk, ensuring the server always works with up‑to‑date JSON. |

### Internal Flow (pseudo‑code)

```mermaid
flowchart TD
    A[Receive raw data] --> B{Unmarshal into struct}
    B -- success --> C[Modify fields]
    B -- fail --> D[log.Fatal("unmarshal failed")]
    C --> E[Marshal back to JSON]
    E -- success --> F[Return bytes]
    E -- fail --> G[log.Fatal("marshal failed")]
```

1. **Unmarshaling** – `data` is parsed into a temporary struct (`tnfJSON`) that mirrors the expected schema.
2. **Updating** – Fields such as `TnfID`, timestamps, or other metadata are patched using values from `r`.
3. **Marshaling** – The modified struct is written back to JSON. A global buffer (`buf`) is cleared and reused for performance.
4. **Error handling** – Any unmarshaling/marshaling failure triggers a fatal log, halting the server.

### Caveats

- The function assumes `data` is well‑formed JSON; malformed input will terminate the server.
- It does not perform validation beyond the built‑in JSON checks (e.g., no schema enforcement).
- Because it uses a global buffer (`buf`), concurrent invocations may race unless guarded by synchronization mechanisms (not shown in this snippet).

### Summary

`updateTnf` is a focused helper that bridges the user‑submitted UI changes and the underlying `.tnf.json` files, ensuring consistency between front‑end edits and back‑end storage within the `webserver` package.
