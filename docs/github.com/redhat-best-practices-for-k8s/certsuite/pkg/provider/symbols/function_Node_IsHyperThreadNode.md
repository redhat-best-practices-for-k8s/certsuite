Node.IsHyperThreadNode`

| Aspect | Detail |
|--------|--------|
| **Package** | `provider` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider`) |
| **Receiver** | `node *Node` – the node instance on which the check is performed. |
| **Signature** | `func (node *Node) IsHyperThreadNode(env *TestEnvironment) (bool, error)` |

### Purpose
Determines whether the Kubernetes node identified by `node.Name` has hyper‑threading enabled.

The function runs a shell command inside the `cns-csi-driver` pod that is running on the target node.  
It parses the output of `/sys/devices/system/cpu/online` to count CPU cores and logical processors, then infers hyper‑threading from the ratio of logical CPUs to physical cores.

### Inputs

| Parameter | Type | Role |
|-----------|------|------|
| `env *TestEnvironment` | *TestEnvironment | Provides access to Kubernetes clients and test context. |

The function does not rely on any global variables; it obtains everything it needs through the environment passed in.

### Outputs

| Return | Type | Meaning |
|--------|------|---------|
| `bool` | true if hyper‑threading is detected (more logical CPUs than cores), false otherwise |
| `error` | non‑nil when any step fails: client acquisition, pod lookup, command execution, or parsing |

### Key Steps & Dependencies

1. **Client Acquisition**  
   ```go
   c := env.GetClientsHolder()
   ```
   Uses the test environment to obtain a holder of Kubernetes clients.

2. **Pod Identification**  
   The function searches for the `cns-csi-driver` pod scheduled on the node:
   ```go
   pods, err := c.CoreV1().Pods("").List(ctx, metav1.ListOptions{FieldSelector: "spec.nodeName=" + node.Name})
   ```
   It expects exactly one such pod; otherwise it returns an error.

3. **Command Execution**  
   Runs `cat /sys/devices/system/cpu/online` inside the pod:
   ```go
   output, err := c.ExecCommandContainer(ctx, ns, podName, containerName, "cat", "/sys/devices/system/cpu/online")
   ```
   This command lists logical CPU IDs (e.g., `0-3,8-11`).

4. **Parsing & Counting**  
   The output is parsed with a regular expression:
   ```go
   re := regexp.MustCompile(`(\d+)-(\d+)`)
   ```
   For each match the number of CPUs in that range is added to `totalCPUs`.  
   Additionally, the function counts the total number of individual CPU IDs by splitting on commas and counting ranges.

5. **Hyper‑threading Decision**  
   If the ratio of logical CPUs (`totalLogical`) to physical cores (`totalPhysical`) is greater than 1, hyper‑threading is considered enabled:
   ```go
   return totalLogical > totalPhysical, nil
   ```

### Side Effects

* No state mutation: it only reads from the cluster and parses output.
* Relies on the presence of a `cns-csi-driver` pod; if absent or multiple pods exist, it returns an error.

### How It Fits the Package

The `provider` package contains helper types (`Node`, `Pod`, etc.) and methods that perform low‑level checks against a Kubernetes cluster.  
`IsHyperThreadNode` is one such check used by higher‑level test suites to validate node configuration (e.g., ensuring that nodes meet performance or security requirements). It exemplifies the pattern of:

1. Selecting a pod on the target node.
2. Executing a command inside that pod.
3. Interpreting the result.

This method is typically called from tests that need to know whether hyper‑threading is active before running workloads that assume a specific CPU topology.
