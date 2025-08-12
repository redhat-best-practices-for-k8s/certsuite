logStreamHandler` – WebSocket log streaming endpoint

| Item | Details |
|------|---------|
| **Package** | `webserver` (github.com/redhat-best-practices-for-k8s/certsuite/webserver) |
| **Signature** | `func(http.ResponseWriter, *http.Request)()` |
| **Exported?** | No – internal helper used by the HTTP router |

### Purpose
`logStreamHandler` is a **WebSocket endpoint** that streams log data to an HTML page in real time.  
When the browser connects to `/logs`, this handler:

1. Upgrades the HTTP connection to WebSocket.
2. Reads logs from the file system (via a `bufio.Scanner`).
3. Sends each new line as a JSON‑encoded WebSocket message.
4. Handles disconnections and errors gracefully.

It is used by the front‑end JavaScript in `logs.js`, which opens a WebSocket to `/logs` and renders received lines in a scrolling view.

### Parameters & Return
| Parameter | Type | Role |
|-----------|------|------|
| `w http.ResponseWriter` | writer for the HTTP response | passed to `upgrader.Upgrade` to create the WebSocket connection |
| `r *http.Request` | incoming request | used by the upgrader and later to read log data |

The function returns an empty closure (`func()`).  
That closure is intended to be executed **after** the request completes, e.g., as a deferred cleanup or logging step. It contains no state of its own beyond what was captured in the outer scope.

### Key Dependencies
| Dependency | How it’s used |
|------------|---------------|
| `upgrader` (`*websocket.Upgrader`) | Calls `Upgrade(w, r)` to switch protocol to WebSocket. |
| `bufio.NewScanner` | Scans the log file line by line. |
| `scanner.Scan()` / `scanner.Bytes()` | Retrieve each log line as a byte slice. |
| `bytes.Buffer` (`buf`) | Temporary buffer used when converting lines to strings for printing/logging. |
| `ConvertToHTML` (internal helper) | Escapes log text for safe HTML rendering before sending over WebSocket. |
| `websocket.WriteMessage` | Sends the escaped line as a text frame (`TextMessage`). |
| `time.Sleep` | Throttles loop to avoid busy‑waiting when no new lines are available. |
| `log.Info`, `log.Err` | Structured logging of connection events and errors. |

### Side Effects & Error Handling
* **WebSocket lifecycle** – The handler immediately upgrades the connection; if upgrade fails, it logs the error and returns.
* **File reading loop** – Continuously scans a log file until EOF or an error occurs. On any scanning error, it logs via `log.Err` and breaks out of the loop.
* **Cleanup** – Calls `conn.Close()` at function exit to free resources.
* **Output** – Sends each log line as an HTML‑escaped string to the client; no direct file or network I/O besides the WebSocket.

### Package Context
The `webserver` package serves a single‑page application that:

1. Serves static assets (`index.html`, `index.js`, `logs.js`, etc.) from embedded files.
2. Provides endpoints for log viewing and test submission.
3. Uses the Gorilla WebSocket library to push live logs.

`logStreamHandler` is one of several handlers registered on the HTTP router (e.g., via `http.HandleFunc("/logs", logStreamHandler)`). It works in concert with the front‑end JavaScript that establishes a WebSocket connection and renders messages inside an `<output>` element.

### Diagram (Mermaid)

```mermaid
flowchart TD
  A[Client] -->|HTTP GET /logs| B[logStreamHandler]
  B -->|Upgrade to WS| C[WebSocket conn]
  subgraph Server loop
    C --> D{Read next log line}
    D -->|new line| E[ConvertToHTML]
    E --> F[WriteMessage (Text)]
    D -->|no new line| G[Sleep(200ms)]
    G --> D
  end
  C -->|Close| H[Cleanup]
```

---

**TL;DR:** `logStreamHandler` upgrades an HTTP request to WebSocket and streams log file lines in real time, escaping them for safe HTML rendering before sending them to the browser. It is a core part of the live‑logging feature in the CertSuite web UI.
