filterDPDKRunningPods`

| Aspect | Detail |
|--------|--------|
| **Package** | `provider` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider`) |
| **Visibility** | Unexported (internal helper) |
| **Signature** | `func filterDPDKRunningPods(pods []*Pod) []*Pod` |

### Purpose
`filterDPDKRunningPods` scans a list of pods and returns only those that are currently running DPDK‑enabled containers.  
The function checks each container in every pod for the presence of the string `"dpdk"` (case sensitive) in its name or image, then verifies that the container is actually **running** by executing `ps -C dpdk` inside it. Pods with no matching containers are discarded.

### Inputs & Outputs
- **Input:** slice of pointers to `Pod` objects (`[]*Pod`).  
  Each `Pod` contains a list of containers (via `pod.Spec.Containers`) and the pod’s status.
- **Output:** a new slice containing only those pods that host at least one running DPDK container.

### Key Steps
1. **Early exit** – if the input slice is empty, return an empty slice immediately.
2. **Iterate over pods** – for each pod:
   - Inspect every container in `pod.Spec.Containers`.
   - If a container’s name or image contains `"dpdk"`, proceed to check runtime status.
3. **Check container state** –  
   - Build a command that greps for the process `dpdk` inside the container (`ps -C dpdk`).  
   - Use `ExecCommandContainer(ctx, pod.Namespace, pod.Name, containerName, cmd...)` (from the provider’s client holder) to run this command.  
   - If the command succeeds and returns non‑empty output, the container is considered running.
4. **Collect qualifying pods** – once a pod has at least one running DPDK container, it is appended to the result slice.

### Dependencies & Side‑Effects
| Dependency | Role |
|------------|------|
| `GetClientsHolder` | Provides access to the Kubernetes client and exec utilities. |
| `NewContext`, `Sprintf` | Build execution context and command string. |
| `ExecCommandContainer` | Executes a shell command inside a pod container; may log errors via `Error`. |
| `Contains` | Checks for substring `"dpdk"` in names/images. |
| `append` | Builds the result slice. |

The function has no visible side‑effects beyond logging possible execution errors. It only reads pod data and performs exec calls.

### Integration with the Package
This helper is used by higher‑level filtering logic that prepares a set of pods for DPDK‑specific tests. By isolating the runtime check into `filterDPDKRunningPods`, the provider keeps the main test orchestration code clean while reusing common Kubernetes client interactions.

```mermaid
flowchart TD
    A[Input Pods] --> B{Iterate each Pod}
    B --> C{Container contains "dpdk"?}
    C -- yes --> D[Exec ps -C dpdk in container]
    D --> E{Command succeeds?}
    E -- yes --> F[Add pod to result]
    C -- no --> G[Skip container]
    E -- no --> H[Skip pod]
    B --> I[Return filtered Pods]
```

> **Note:** The function assumes that the Kubernetes client holder is correctly initialized (`GetClientsHolder()`), and it relies on the standard `ExecCommandContainer` helper to run commands inside containers.
