waitPodDeleted`

`waitPodDeleted` is an internal helper that blocks until a particular pod has been removed from the cluster (or a timeout occurs).  
It is used by the pod‑recreation tests to assert that a controller has successfully cleaned up a pod before attempting to recreate it.

```go
func waitPodDeleted(namespace, podName string, timeoutSec int64, watcher watch.Interface) func()
```

| Parameter | Type   | Meaning |
|-----------|--------|---------|
| `namespace` | `string` | Kubernetes namespace containing the target pod. |
| `podName`    | `string` | Name of the pod that should be deleted. |
| `timeoutSec` | `int64`  | Maximum number of seconds to wait before giving up. |
| `watcher`     | `watch.Interface` | A Kubernetes watch that streams pod events for the given namespace. |

### Return value

A **closure** (`func()`) that, when called, will block until one of two conditions is met:

1. The watched stream reports a deletion event for the specified pod.
2. The timeout expires.

The closure does not return any value; it simply performs logging and signals completion via `Stop()` on the watch channel.

### Key logic flow

```text
+-----------------------+
| Start watching events |
+----------+------------+
           |
           | (event, ok := <-watcher.ResultChan())
           v
+---------------------------+
| Event received?          |
+------+--------------------+
       | yes
       v
+-----------------------------+
| Is it the target pod?      |
+------+----------------------+
       | no
       v
+------------------------+
| Ignore, keep waiting  |
+------------------------+
```

When the closure receives a `DELETE` event for the target pod, it logs a debug message and stops the watcher by calling `watcher.Stop()`.

If the specified timeout elapses before such an event is observed, the function logs a warning (`Info`) and also stops the watcher.

### Dependencies & side‑effects

| Call | Purpose |
|------|---------|
| `Debug` | Logs detailed debugging information (e.g., waiting status). |
| `Stop`  | Stops the Kubernetes watch to free resources. |
| `ResultChan` | Receives events from the watch. |
| `After`, `Duration` | Measure elapsed time for timeout handling. |
| `Info` | Emits a warning if the pod was not deleted within the allotted time. |

The function has no external side‑effects other than logging and terminating the provided watch.

### How it fits the package

- **Package**: `podrecreation` – tests that verify controller behavior during pod recreation cycles.
- **Usage pattern**: After triggering a deletion or a controller action, the test creates a watcher on the relevant namespace and passes it to `waitPodDeleted`. The returned closure is then executed in a goroutine or via `Eventually(...).Should(BeTrue())` (common in Ginkgo tests) to wait for pod removal before proceeding.

```go
watcher, _ := client.CoreV1().Pods(namespace).Watch(...)
defer watcher.Stop()
waitFunc := waitPodDeleted(namespace, podName, 60, watcher)
```

Thus `waitPodDeleted` centralizes the logic of polling a watch stream for pod deletion while respecting timeouts and logging, keeping test code concise and expressive.
