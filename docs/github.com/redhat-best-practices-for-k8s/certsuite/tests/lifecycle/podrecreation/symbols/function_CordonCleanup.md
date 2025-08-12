CordonCleanup` – Helper for Test‑Suite Tear‑Down

```go
func CordonCleanup(node string, c *checksdb.Check)()
```

| Aspect | Details |
|--------|---------|
| **Purpose** | Undo a node cordon that was applied during the pod‑recreation test.  The function is intended to be used as a cleanup routine (e.g., `t.Cleanup`) so that the test leaves the cluster in its original state. |
| **Parameters** | *`node string`* – name of the Kubernetes node that was cordoned.<br>*`c *checksdb.Check`* – the check context that holds test metadata and logging facilities. |
| **Return value** | None; it performs side‑effects only. |
| **Key dependencies** | • `CordonHelper` – actually toggles the node’s schedulable flag.<br>• `Abort` – aborts the current test with a fatal error if the cleanup fails.<br>• `Sprintf` (from `fmt`) – builds an informative log message. |
| **Side‑effects** | Calls `CordonHelper(node, false)` to uncordon the node.  If that call returns an error it aborts the current test via `Abort(c, err)`. The function also logs the action with a formatted string. |
| **How it fits the package** | In *podrecreation* tests we temporarily cordon a node to prevent new pods from being scheduled while recreating existing ones.  After each test run we must return the node to its original state; `CordonCleanup` is registered as a cleanup routine so that this happens automatically, even if the test panics or fails early. |


### Usage pattern

```go
// Inside a test:
node := "worker-1"
CordonHelper(node, true)          // cordon before recreation
t.Cleanup(func() { CordonCleanup(node, c) })  // ensure uncordon afterwards
```

The function is deliberately simple: it delegates the actual API call to `CordonHelper` and relies on the test harness (`Abort`) for error handling. This keeps cleanup logic consistent across the test suite while centralizing node‑un‑cordoning in a single, documented place.
