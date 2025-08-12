runHandler` – Web‑Request Driver for CERTSUITE

### Purpose
`runHandler` is the HTTP handler that powers the **/run** endpoint of the Certsuite web UI.  
When a user submits the form (or uploads a YAML file) through the browser, this function:
1. Parses the incoming request.
2. Persists the submitted test data to a temporary directory.
3. Invokes `updateTnf` to create a *Test‑N‑F* (TNF) configuration file from that data.
4. Starts a Certsuite run inside a fresh process and captures its output.
5. Sends back an HTTP response containing a JSON blob with the test result path.

In short, it bridges the browser UI to the underlying Certsuite engine.

---

### Signature
```go
func runHandler(http.ResponseWriter, *http.Request)()
```
- Returns a closure that will be executed by `http.HandlerFunc` – this is the idiomatic pattern used in the package to keep the handler body readable while still being able to capture local variables (e.g. `buf`, `upgrader`).

---

### Key Inputs

| Variable | Type | Source | Notes |
|----------|------|--------|-------|
| `http.ResponseWriter` (`w`) | `http.ResponseWriter` | Function argument | Used to send JSON and error responses. |
| `*http.Request` (`r`) | `*http.Request` | Function argument | Holds form fields, uploaded files, and context values. |

### Key Outputs

- **HTTP response** – either a 200‑OK with a JSON body containing the path to the generated TNF file or an error status.
- **Side effects** – temporary directory creation, file writes, logger configuration, process launch.

---

### Core Steps & Dependencies

1. **Context Retrieval**
   ```go
   outputFolder := r.Context().Value(outputFolderCtxKey).(string)
   ```
   *Depends on* `outputFolderCtxKey` (a context key defined in the same package).

2. **Form Parsing**
   - Reads `name`, `tags`, `testdata`, and `yamlFile` form values.
   - If a file is uploaded (`yamlFile`) it is read into memory.

3. **Logging Setup**
   ```go
   buf = NewBufferString()
   SetLogger(buf)
   ```
   *Creates* an in‑memory buffer to capture logs, then registers it with the global logger.

4. **Test Data File Creation**
   - A temporary directory is created:
     ```go
     tmpDir, _ := os.CreateTemp(outputFolder, "certsuite")
     ```
   - Test data (either from `testdata` or uploaded file) is written to `tmpDir/TestData.yaml`.

5. **TNF Generation (`updateTnf`)**
   ```go
   updateTnf(tmpDir)
   ```
   The function populates a *Test‑N‑F* configuration that tells Certsuite which tests to run.

6. **Process Execution**
   - A new `exec.Cmd` is spawned with the appropriate arguments.
   - Standard output and error are redirected to temporary files inside `tmpDir`.
   - The command is started and awaited (`cmd.Wait()`).

7. **Result Handling**
   - On success, the path of the generated TNF file (relative to `outputFolder`) is written back in JSON:
     ```go
     fmt.Fprintf(w, "{\"result\":\"%s\"}", tnfPath)
     ```
   - On failure, an error status and message are returned.

8. **Cleanup**
   - The temporary directory and its contents are removed (`os.RemoveAll(tmpDir)`).
   - Logger buffer is reset with `buf.Reset()`.

---

### Side‑Effects & Important Notes

- **Global state**:  
  *`buf`* (a package‑level `*bytes.Buffer`) holds logs for the current request; it is reused per handler invocation.  
  The logger is temporarily redirected to this buffer, then restored after the run completes.
  
- **Temporary files**: All data written during a run lives in a unique subdirectory under the *output folder*. This directory is cleaned up immediately after the process finishes.

- **Error handling**: Any step that fails aborts the handler and writes an appropriate HTTP status (e.g., `400 Bad Request`, `500 Internal Server Error`) with the error message.

---

### Interaction With Other Package Components

| Component | Relationship |
|-----------|--------------|
| `updateTnf` | Generates the TNF file needed for Certsuite. |
| `GetNewClientsHolder` & `LoadChecksDB` | Used indirectly by Certsuite itself during execution; not called directly in this handler. |
| Web assets (`index.html`, `submit.js`, etc.) | Serve the UI that posts to `/run`; the handler is the server‑side counterpart. |

---

### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[Client POST /run] --> B{Parse Form}
    B --> C{Create Temp Dir}
    C --> D[Write TestData.yaml]
    D --> E[updateTnf(tmpDir)]
    E --> F[Launch Certsuite process]
    F --> G{Process Exit?}
    G -- success --> H[Return JSON {result: tnfPath}]
    G -- failure --> I[Return HTTP error]
    H & I --> J[Cleanup tmpDir]
```

This diagram captures the high‑level flow of `runHandler`.
