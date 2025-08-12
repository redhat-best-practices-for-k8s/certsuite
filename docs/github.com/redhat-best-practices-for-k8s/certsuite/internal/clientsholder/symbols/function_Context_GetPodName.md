## `GetPodName` – Retrieve the current pod name

| Aspect | Detail |
|--------|--------|
| **Signature** | `func (ctx Context) GetPodName() string` |
| **Receiver** | `Context` – a struct that holds client‑side state for interacting with a Kubernetes cluster. |
| **Return value** | The name of the pod in which the current process is running, or an empty string if it cannot be determined. |

### Purpose
`GetPodName` provides a convenient way to discover the pod identity from within a certsuite test environment.  
The function is used by other helpers that need to:

* Tag logs with the pod name,
* Store temporary files under a pod‑specific directory, or
* Resolve resources scoped to the current pod.

### How it works

1. **Read the `POD_NAME` environment variable** – This is the canonical way Kubernetes injects the pod name into a container.
2. **If unset, fall back to the hostname** – A pod’s hostname defaults to its name in most deployments.
3. The chosen value is returned as a plain string.

No external calls are made (e.g., no API server request), so the function is fast and side‑effect free.

### Dependencies & Side Effects

| Dependency | Notes |
|------------|-------|
| `os.Getenv` | Reads environment variables; no global state change. |
| `os.Hostname` | System call that may error; errors are ignored in favor of an empty string. |

There are **no side effects** beyond reading from the OS and returning a value.

### Role within the package

The `clientsholder` package centralises Kubernetes client configuration and helper utilities for certsuite tests.  
`GetPodName` is one such utility; it complements functions that:

* Acquire clientsets (`NewClientSet`, `GetClientsHolder`),
* Handle timeouts (`DefaultTimeout`), or
* Mock commands (`CommandMock`).

By abstracting pod‑name retrieval, the rest of the package can remain agnostic to deployment specifics and simply rely on this helper when needed.

### Example usage

```go
ctx := clientsholder.Context{}
pod := ctx.GetPodName()
log.Infof("Running in pod %s", pod)
```

This call will log the current pod name or an empty string if it cannot be determined.
