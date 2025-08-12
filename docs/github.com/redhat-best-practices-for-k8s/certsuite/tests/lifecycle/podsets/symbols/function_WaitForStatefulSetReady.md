WaitForStatefulSetReady`

```go
func WaitForStatefulSetReady(namespace, name string, timeout time.Duration, log *log.Logger) bool
```

## Purpose

`WaitForStatefulSetReady` blocks until a StatefulSet in the given Kubernetes namespace reaches the **ready** state or the supplied `timeout` expires.  
It is used by tests that need to guarantee that all pods of a StatefulSet are running and ready before proceeding.

The function returns:

* **true** – when the StatefulSet becomes ready within the timeout.
* **false** – if the timeout elapses without readiness, or an unrecoverable error occurs.

## Parameters

| Name       | Type          | Description |
|------------|---------------|-------------|
| `namespace`| `string`      | Namespace containing the StatefulSet. |
| `name`     | `string`      | The name of the StatefulSet to watch. |
| `timeout`  | `time.Duration` | Maximum time to wait for readiness. |
| `log`      | `*log.Logger` | Logger used for debug/diagnostic output (may be nil). |

## Key Dependencies

1. **Kubernetes client**  
   Obtained via `GetClientsHolder()`, which returns a holder that exposes the AppsV1 client.

2. **StatefulSet helpers**  
   * `GetUpdatedStatefulset` – fetches the latest StatefulSet object from the API.  
   * `IsStatefulSetReady` – evaluates whether the retrieved StatefulSet is ready (all replicas are up and ready).

3. **Time utilities**  
   * `time.Now()` and `time.Since()` for measuring elapsed time.

4. **Logging helpers**  
   * `Debug`, `Info`, `Error` – wrapper functions that log messages if a logger is supplied.

5. **Utility conversions**  
   * `ToString` – converts the StatefulSet object to a human‑readable string (used in logs).

## Control Flow

1. Record start time (`start := Now()`).
2. Loop until elapsed time exceeds `timeout`:
   1. Log debug entry.
   2. Fetch current StatefulSet via `GetUpdatedStatefulset`.
   3. If fetch fails → log error and return `false`.
   4. Log info with the current state.
   5. Check readiness using `IsStatefulSetReady`.  
      *If ready* → return `true`.
   6. Sleep for a short interval (`time.Sleep(1s)`).
3. Timeout reached → log error and return `false`.

## Side Effects

* **Logging** – emits debug, info, or error messages to the supplied logger.
* **Kubernetes API calls** – repeatedly polls the API server until timeout.

No state is mutated in the caller’s context; all changes are read‑only checks against the cluster.

## Package Context

The `podsets` package contains helpers for managing Kubernetes pod sets (Deployments, StatefulSets, ReplicaSets) during integration tests.  
Other exported functions include:

* `WaitForDeploymentSetReady` – similar logic for Deployments.
* `WaitForScalingToComplete` – waits until scaling actions finish.

`WaitForStatefulSetReady` is the StatefulSet counterpart and provides a reusable waiting mechanism used by various test scenarios that require a fully operational StatefulSet before advancing.
