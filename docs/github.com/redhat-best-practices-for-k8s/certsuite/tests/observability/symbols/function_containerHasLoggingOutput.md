containerHasLoggingOutput`

| | |
|---|---|
| **Package** | `observability` |
| **Location** | `suite.go:95` |
| **Signature** | `func(container *provider.Container) (bool, error)` |
| **Exported** | No |

### Purpose
`containerHasLoggingOutput` is a helper used in the test suite to verify that a specific container produced at least one line of log output during its run.  
It returns:

* `true` – if any logs were retrieved from the container.
* `false` – if no logs were found (or an empty stream).
* `error` – if anything went wrong while trying to fetch or read the logs.

### Inputs

| Parameter | Type | Description |
|-----------|------|-------------|
| `container` | `*provider.Container` | The container whose logs are being inspected. This struct contains fields such as the pod name, namespace and the container name itself (used when constructing the log request). |

### Workflow & Key Dependencies

1. **Kubernetes client acquisition**  
   ```go
   clients := GetClientsHolder(env)
   ```
   * `GetClientsHolder` pulls a cached set of Kubernetes client interfaces from the test environment (`env`).  
   * This provides access to the CoreV1 API needed for log streaming.

2. **Log request construction**  
   ```go
   podLogs, err := clients.CoreV1().Pods(container.Namespace).GetLogs(
       container.PodName,
       &v1.PodLogOptions{Container: container.Name},
   ).Stream(context.TODO())
   ```
   * `CoreV1()` → returns the Core V1 client.  
   * `.Pods(namespace)` scopes to the correct namespace.  
   * `GetLogs(...).Stream(...)` opens a streaming HTTP connection that yields the pod’s log output.

3. **Read & evaluate logs**  
   ```go
   defer podLogs.Close()
   var lastLine string
   buf := new(bytes.Buffer)
   _, err = io.Copy(buf, podLogs)
   if err != nil { ... }
   if buf.Len() == 0 {
       return false, nil
   }
   ```
   * The entire stream is copied into a buffer.  
   * If the buffer length is zero, no output was produced → `false`.  
   * Otherwise the function returns `true`.

4. **Error handling**  
   All errors from client retrieval, log streaming, or copying are wrapped with contextual messages (`fmt.Errorf`) and returned to the caller.

### Side Effects & Assumptions

* The function **does not modify** the container or its pod; it only reads logs.  
* It assumes that `env` (the test environment) is properly initialized and that the Kubernetes client can connect to the cluster.  
* The log stream is closed in a `defer` statement, ensuring no resource leaks.

### Integration with the Package

Within the `observability` test suite this helper is used by higher‑level tests that:

1. **Launch a workload** (e.g., a sidecar or an application pod).  
2. **Wait for the container to finish** or reach a desired state.  
3. Call `containerHasLoggingOutput` to confirm that the container emitted logs, which is a key observable metric in many tests.

By abstracting the log‑streaming logic into this small helper, test authors can focus on orchestrating workloads and asserting outcomes without re‑implementing Kubernetes API interactions.
