StartServer`

```go
func StartServer(addr string) func()
```

| Item | Detail |
|------|--------|
| **Purpose** | Starts an HTTP server that serves the CertSuite web UI and WebSocket endpoints for live logs. The function returns a cleanup closure that can be called to stop the server gracefully (not shown in the snippet). |
| **Parameters** | `addr string` – Address on which the server will listen (`":8080"`, `"127.0.0.1:8443"`, etc.). |
| **Return value** | A zero‑argument function that performs any necessary shutdown logic when invoked. (The actual body of this closure is not present in the snippet.) |
| **Key dependencies** | • `installReqHandlers()` – registers custom request handlers for static assets and API routes.<br>• `http.HandleFunc` – associates paths with handler functions.<br>• `logTimeout`, `readTimeoutSeconds` – constants controlling timeouts (see file level).<br>• `WithValue` – attaches context values to requests, notably the output folder via `outputFolderCtxKey`. |
| **Side effects** | • Writes an informational log line when the server starts (`Info("Starting webserver on %s")`).<br>• Calls `http.ListenAndServe(addr, nil)` which blocks until the process is terminated or a fatal error occurs.<br>• Panics if the underlying listener cannot be created. |
| **How it fits the package** | The `webserver` package bundles static assets (`index.html`, `index.js`, etc.) and WebSocket support for live logs. `StartServer` is the entry point that brings everything together: it installs request handlers, configures logging context, and launches the HTTP listener. Other parts of CertSuite can call this function to expose the UI during a test run or in a standalone mode. |

---

### Mermaid diagram (suggested)

```mermaid
flowchart TD
    A[StartServer(addr)] --> B[installReqHandlers()]
    B --> C{Register Routes}
    C -->|/index.html| D[index.html handler]
    C -->|/api/logs| E[WebSocket log handler]
    A --> F[http.ListenAndServe(addr, nil)]
```

*The diagram illustrates the high‑level flow: installation of handlers followed by blocking on `ListenAndServe`.*
