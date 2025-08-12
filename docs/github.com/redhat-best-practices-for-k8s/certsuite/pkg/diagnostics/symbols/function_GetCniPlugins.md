GetCniPlugins`

**Package:** `diagnostics`  
**File:** `pkg/diagnostics/diagnostics.go`  
**Signature:**  

```go
func GetCniPlugins() map[string][]interface{}
```

---

### Purpose

Collects the CNI (Container Network Interface) plugins that are installed on every node in a Kubernetes cluster and returns them as a JSON‑compatible Go structure. The function is used by diagnostic tooling to verify network configuration across all nodes.

### Return value

```go
map[string][]interface{}
```

* **Key** – Node name (`string`).  
* **Value** – Slice of decoded CNI plugin objects (`[]interface{}`). Each element corresponds to one JSON object returned by the `cni-plugins` command executed inside a container on that node.

If an error occurs while querying a node, its entry will contain a single element: a map with `"error"` → `<error string>`.

### Key dependencies & workflow

1. **Environment setup**  
   * Calls `GetTestEnvironment()` to obtain the current Kubernetes test environment (cluster name, kubeconfig, etc.).  
   * Uses `GetClientsHolder()` to retrieve a client holder that provides access to the API server and node list.

2. **Node enumeration**  
   * Iterates over all nodes returned by the clients holder.

3. **Command execution**  
   * For each node, creates a new context with a 30‑second timeout via `NewContext()`.  
   * Executes `cni-plugins` inside a container on that node using `ExecCommandContainer(context, clientset, nodeName, cniPluginsCommand)`.  
   * The command string is defined by the package constant `cniPluginsCommand`.

4. **Result processing**  
   * If execution fails: logs error via `Error()` and stores an error map in the result slice.  
   * On success: unmarshals the JSON output (`Unmarshal`) into a slice of empty interfaces, which is then stored under the node’s key.

5. **Return**  
   * After processing all nodes, returns the populated map.

### Side effects

* **Logging** – Errors are logged with `Error()` but otherwise no side‑effects occur on the cluster.  
* **Timeouts** – Each command runs within a 30‑second context; if exceeded, the node’s entry will contain an error.

### Diagram (optional)

```mermaid
flowchart TD
    A[GetTestEnvironment] --> B[GetClientsHolder]
    B --> C{Nodes}
    C --> D[ExecCommandContainer]
    D -->|Success| E[Unmarshal JSON]
    D -->|Error| F[Store error map]
    E --> G[Result slice]
    F --> G
    G --> H[Return map[string][]interface{}]
```

### How it fits the package

`GetCniPlugins` is one of several helper functions in `diagnostics.go` that expose cluster state as JSON‑ready Go structures. It complements other utilities like `GetNodeInfo`, `GetPods`, etc., and is typically invoked by higher‑level diagnostic commands or test suites to validate network plugin installation across the Kubernetes environment.
