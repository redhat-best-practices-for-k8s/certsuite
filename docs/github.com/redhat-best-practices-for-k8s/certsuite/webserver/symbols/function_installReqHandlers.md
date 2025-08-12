installReqHandlers`

```
func installReqHandlers() func()
```

> **Purpose**  
`installReqHandlers` registers the HTTP handlers that serve the web‑interface of CertSuite and returns a cleanup function (currently a no‑op).  It is called once during server initialisation to bind URLs to handler functions.

---

### How it works

| Step | What happens | Key calls |
|------|--------------|-----------|
| 1 | Register **root** (`/`) handler | `http.HandleFunc("/", func(...))` |
| 2 | Set the content type header to `text/html; charset=utf-8`.  
The handler writes the embedded `indexHTML` bytes and ends. | `w.Header().Set`, `w.Write` |
| 3 | Register **static file** handlers: `index.js`, `logs.js`, `submit.js`, `toast.js` | `http.HandleFunc("/index.js", ...)`, etc. Each handler writes the corresponding embedded byte slice. |
| 4 | Register **WebSocket** endpoint `/ws` to upgrade the connection using `upgrader`. The actual WebSocket logic is in `outputTestCases`. | `upgrader.Upgrade`, `outputTestCases` |
| 5 | Return a cleanup function that currently does nothing (placeholder for future resource release). | `return func() {}` |

The function uses only the embedded byte slices (`indexHTML`, `index`, `logs`, `submit`, `toast`) and the global `upgrader`. No external state is mutated.

---

### Inputs & Outputs

| Direction | Description |
|-----------|-------------|
| **Input** | None – the handler functions capture all required data from the package globals. |
| **Output** | A zero‑argument function (`func()`). The returned closure is meant to be called when shutting down the server (currently empty). |

---

### Key Dependencies

- `net/http` for `HandleFunc`, header manipulation, and writing responses.
- `gorilla/websocket` (implied by `upgrader`) for WebSocket upgrades.
- Embedded assets (`index.html`, `.js` files) injected at compile time via `//go:embed`.

---

### Side‑Effects & Observations

* Registers global URL patterns – calling this multiple times would overwrite previous handlers.
* Does **not** alter any global state other than the handler registration itself.
* The cleanup function is currently a no‑op; future iterations may close websockets or release resources.

---

### How it fits in the `webserver` package

The package orchestrates a simple HTTP server that serves a single-page UI and exposes a WebSocket endpoint for live test output.  
`installReqHandlers` is the glue that connects URL paths to their respective handler functions, enabling the web interface to load static assets and establish real‑time communication.

```mermaid
flowchart LR
  A[main.go] --> B[initWebServer]
  B --> C{installReqHandlers}
  C --> D[/ -> index.html]
  C --> E[/index.js -> embedded bytes]
  C --> F[/logs.js -> embedded bytes]
  C --> G[/submit.js -> embedded bytes]
  C --> H[/toast.js -> embedded bytes]
  C --> I[/ws -> WebSocket upgrader]
```

*The diagram illustrates the URL routes and their handler destinations.*
