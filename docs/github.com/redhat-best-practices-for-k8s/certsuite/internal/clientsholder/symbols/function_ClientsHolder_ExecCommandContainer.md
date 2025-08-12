# `ExecCommandContainer`

> **Package**: `clientsholder`  
> **File**: `internal/clientsholder/command.go` (line 36)  
> **Receiver**: `ClientsHolder`  

## Purpose

`ExecCommandContainer` runs an arbitrary shell command inside a specific container of a pod and returns the combined stdout/stderr output as a string. It is used by tests and helper utilities that need to interact with a running workload without exposing the underlying Kubernetes API directly.

> *Note*: The method name contains “Container” but the comment in the source says `ExecCommand runs command in the pod …`. In practice it uses the container name supplied via the `ClientsHolder` configuration, so the function is effectively **Execute Command in Container**.

## Signature

```go
func (ch ClientsHolder) ExecCommandContainer(ctx Context, cmd string) (string, error)
```

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `ctx`     | `Context` | Execution context (for cancellation/timeout). |
| `cmd`     | `string`  | The command to run inside the container. |

| Return | Type   | Meaning |
|--------|--------|---------|
| `stdoutStderr string` | Combined output of stdout and stderr from the executed command. |
| `err error` | Non‑nil if any step fails (HTTP request, executor creation, streaming). |

## How it works

1. **Logging** – The method starts by logging the intended execution via `Debug`.

2. **Build REST request**  
   * Uses `ch.RESTClient()` to obtain a client for the CoreV1 API.
   * Constructs an HTTP POST to the sub‑resource `exec` of the target pod:
     ```go
     req := ch.RESTClient().CoreV1().
              Namespace(ch.GetNamespace()).
              Pods(ch.GetPodName()).Exec()
     ```
   * Sets request parameters (`container`, `command`, etc.) via `VersionedParams`.
   * The resulting URL is logged.

3. **Create SPDY executor**  
   Calls `NewSPDYExecutor` with the REST config to obtain an executor capable of streaming over a WebSocket/SPDY tunnel.

4. **Execute and stream**  
   Uses `executor.StreamWithContext(ctx, ...)` to pipe the command’s output into buffers. The same buffer is used for both stdout and stderr (`stdoutStderr`).

5. **Return** – On success returns the captured string; otherwise returns an error with context information.

## Dependencies

| Dependency | Role |
|------------|------|
| `Debug` | Logging helper from `ClientsHolder`. |
| `GetNamespace`, `GetPodName`, `GetContainerName` | Retrieve configuration values. |
| `RESTClient`, `CoreV1` | Build Kubernetes REST request. |
| `NewSPDYExecutor` | Create executor for streaming exec. |
| `StreamWithContext` | Perform the actual command execution. |

All these helpers are defined on the same `ClientsHolder` struct or its embedded configuration, making this method a thin wrapper around the standard client-go exec flow.

## Side‑effects & Notes

* **No state mutation** – The function does not alter any fields of `ClientsHolder`.  
* **Timeout handling** – Uses the provided `ctx`; if no deadline is set it will run until completion.  
* **Error handling** – Errors from each step are wrapped with descriptive messages for easier debugging.

## Usage context

`ExecCommandContainer` is typically invoked by higher‑level tests or utilities that need to:

- Check container readiness (`curl`, `test -f …`).
- Retrieve logs or status files.
- Perform configuration checks inside the pod.

It is part of the **clientsholder** abstraction, which centralizes Kubernetes client logic so that callers can focus on business logic rather than REST plumbing.
