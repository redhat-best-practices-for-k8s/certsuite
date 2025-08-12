deletePod` – Internal helper for pod lifecycle tests

| Item | Description |
|------|-------------|
| **Signature** | `func deletePod(pod *corev1.Pod, deletionPropagation string, wg *sync.WaitGroup) error` |
| **Visibility** | Unexported (used only inside the `podrecreation` package). |
| **Purpose** | Removes a given Pod from the cluster while optionally controlling the Kubernetes garbage‑collection behaviour. It also waits for the pod to disappear before returning, and records any errors that occur during deletion or waiting. |

### Parameters

| Name | Type | Meaning |
|------|------|---------|
| `pod` | `*corev1.Pod` | The Pod object that should be deleted. |
| `deletionPropagation` | `string` | One of the constants defined in this package: `DeleteBackground`, `DeleteForeground`, or `NoDelete`.  These map to Kubernetes propagation policies (`background`, `foreground`, or no deletion). |
| `wg` | `*sync.WaitGroup` | A WaitGroup used by the test harness to coordinate concurrent pod deletions. The function calls `wg.Done()` when finished, regardless of success or failure. |

### Return Value

- `error`:  
  * `nil` if the pod was deleted successfully and confirmed gone.  
  * non‑`nil` otherwise (e.g., API call failures, timeout while waiting).

### High‑level workflow

1. **Log debug info** – The function starts by logging the name of the pod being removed.
2. **Set up a watch** – A Kubernetes `Watch` is created on the namespace to observe Pod events. This allows the test to react immediately when the pod disappears.
3. **Delete request** – It calls the client’s `Pods().Delete()` method with the specified propagation policy.
4. **Wait for deletion** – Using the helper `waitPodDeleted`, it blocks until either:
   * The watched event indicates that the pod is gone, or
   * A timeout occurs (the test harness defines a reasonable timeout).
5. **Error handling & finalization** – Any error from deletion or waiting is wrapped with context and returned. Finally, `wg.Done()` signals completion.

### Key dependencies

| Dependency | Role |
|------------|------|
| `GetClientsHolder` | Provides the Kubernetes clientset used for API calls. |
| `CoreV1().Pods(namespace)` | Access to pod operations in the target namespace. |
| `Watch` | Streams Pod events for real‑time deletion confirmation. |
| `waitPodDeleted` | Encapsulates logic that polls or listens until the pod is confirmed removed. |
| `sync.WaitGroup` | Synchronizes test goroutines. |

### Side effects

* The function performs a **real deletion** in the cluster – it does not mock the API.
* It updates shared state via the WaitGroup; callers must ensure the WaitGroup is correctly initialised.
* Logs are emitted through the package’s `Debug` and `Errorf` helpers, which may be routed to test output.

### How it fits into the package

The `podrecreation` tests exercise scenarios where Pods created by Deployments, ReplicaSets, StatefulSets, etc., need to be removed and recreated.  
`deletePod` is the low‑level primitive that:

* removes a pod according to a chosen propagation policy,
* guarantees that the test only proceeds once the cluster has fully acknowledged deletion,
* reports any problems back to the caller.

Higher‑level functions in this package (e.g., `restartPods`, `testPodRecreation`) invoke `deletePod` inside goroutines, passing the same WaitGroup so all deletions finish before the test asserts final state. This design keeps the delete logic isolated and reusable across different lifecycle tests.
