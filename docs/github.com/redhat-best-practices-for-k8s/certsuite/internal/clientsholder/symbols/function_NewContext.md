NewContext`

```go
func NewContext(namespace string, clusterID string, token string) Context
```

### Purpose
Creates a new **`Context`** value that encapsulates all information required to talk to the Certsuite API in a specific Kubernetes namespace and cluster.

The returned context is used by the rest of the package to:

| Action | What the `Context` provides |
|--------|-----------------------------|
| **HTTP requests** | The base URL, authentication token and request‑timeout settings. |
| **Client construction** | A pre‑configured HTTP client that respects the supplied timeout. |
| **Namespace / cluster identification** | Allows API calls to be scoped correctly (e.g., to a tenant or environment). |

### Parameters

| Name       | Type   | Description |
|------------|--------|-------------|
| `namespace` | `string` | The Kubernetes namespace in which the Certsuite resources live. |
| `clusterID` | `string` | Identifier of the cluster (used for namespacing or logging). |
| `token`     | `string` | Bearer token used for authenticating to the API server. |

> **Note** – All parameters are required; passing an empty string will still create a context, but downstream calls may fail.

### Return Value

* `Context` – A struct that contains:

  * The supplied namespace, cluster ID and token.
  * An HTTP client configured with:
    * `Timeout: DefaultTimeout` (see constant below).
    * A transport that injects the bearer token into every request.
  * Any other package‑internal helpers required by the API.

The function is pure: it has no side effects such as writing to global state or performing I/O.

### Key Dependencies

| Dependency | How it’s used |
|------------|---------------|
| `DefaultTimeout` (exported constant) | Sets the timeout on the HTTP client created in this context. |
| Internal types (`Context`, `Command`) | The returned value is of type `Context`; its definition lives in the same package. |

### Package Integration

* **Where it’s called** – The public API entry points (e.g., `clientsholder.NewClient()`) call `NewContext` to obtain a configured context before performing any request.
* **Why it matters** – Centralising configuration in a single function simplifies testing and guarantees that all clients share the same timeout behaviour.

### Example

```go
ctx := clientsholder.NewContext("certsuite", "cluster-1", "my‑bearer‑token")
client, _ := clientsholder.NewClient(ctx)
// … use client …
```

This snippet shows how a caller builds a context and immediately uses it to create an authenticated HTTP client.

---

**Summary:** `NewContext` is the package’s factory for the `Context` type. It bundles namespace/cluster identity with authentication credentials, sets up a timeout‑aware HTTP client using the exported `DefaultTimeout`, and returns a fully‑ready value that downstream code can use without needing to reconfigure these common settings.
