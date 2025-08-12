CordonHelper` – Package `podrecreation`

**Purpose**

`CordonHelper` toggles the *cordon* state of a Kubernetes node, i.e., it marks a node as
schedulable (`Uncordon`) or unschedulable (`Cordon`).  
It is used in the test suite to simulate node maintenance scenarios: after cordoning a
node, pods are evicted and later recreated by the cluster controller.

**Signature**

```go
func CordonHelper(nodeName string, action string) error
```

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `nodeName` | `string` | The name of the node to modify. |
| `action`   | `string` | One of `"cordon"` or `"uncordon"`. Case‑insensitive. |

The function returns an error if any Kubernetes API call fails or if the supplied
`action` is invalid.

**Key Steps**

1. **Retrieve Kubernetes client**  
   Uses `GetClientsHolder()` to obtain a `kubernetes.Clientset`.

2. **Log intent**  
   Calls `Info(fmt.Sprintf("CordonHelper: %s node %q", action, nodeName))`.

3. **Handle the action**  
   * If `action == "cordon"`:  
     * Wrap the update in `RetryOnConflict` to handle concurrent modifications.  
     * Fetch the node via `client.CoreV1().Nodes().Get(...)`.  
     * Set `node.Spec.Unschedulable = true`.  
     * Call `client.CoreV1().Nodes().Update(...)`.
   * If `action == "uncordon"`:  
     * Same pattern but set `Unscheduled = false`.

4. **Return**  
   Propagates any errors from the API or conflict handling.

**Dependencies**

| Dependency | Role |
|------------|------|
| `GetClientsHolder` | Provides Kubernetes clientset. |
| `RetryOnConflict`  | Handles update conflicts by retrying. |
| `CoreV1().Nodes()` | Node CRUD operations. |
| Logging helpers (`Info`, `Errorf`) | Emit test logs. |

**Side Effects**

* Modifies the node’s schedulable flag in the cluster, affecting pod placement.
* No other objects are touched.

**How It Fits the Package**

`podrecreation.go` contains tests that exercise the recreation of pods when nodes
are cordoned/uncordoned. `CordonHelper` is a small utility that drives those tests,
abstracting the node‑level manipulation away from the test logic.  
It is exported so it can be reused by other test modules if needed.

**Mermaid Diagram (suggested)**

```mermaid
flowchart TD
    A[Call CordonHelper(node, action)] --> B{action}
    B -- cordon --> C[Get node]
    C --> D[node.Spec.Unschedulable = true]
    D --> E[Update node]
    B -- uncordon --> F[Get node]
    F --> G[node.Spec.Unschedulable = false]
    G --> H[Update node]
```

---
