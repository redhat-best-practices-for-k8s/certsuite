ExecCommandContainerNSEnter`

| Item | Details |
|------|---------|
| **Package** | `crclient` (`github.com/redhat-best-practices-for-k8s/certsuite/internal/crclient`) |
| **Exported** | Yes |
| **Signature** | `func ExecCommandContainerNSEnter(cmd string, container *provider.Container) (string, error)` |

### Purpose
Executes a shell command inside the namespace of a running Kubernetes container by using `nsenter`.  
The function first obtains the PID of the target process within the container, then uses the probe pod’s kube‑client to run `nsenter` in that namespace. It retries the operation until it succeeds or exhausts a configurable number of attempts.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `cmd` | `string` | The command to be executed inside the container (e.g., `"cat /etc/os-release"`). |
| `container` | `*provider.Container` | Reference to the target container; must contain its pod name, namespace and PID. |

### Returns
| Value | Type | Description |
|-------|------|-------------|
| `string` | The standard output of the executed command (trimmed). |
| `error` | Any error encountered during the process. |

### Workflow

```mermaid
flowchart TD
    A[GetTestEnvironment] --> B[GetNodeProbePodContext]
    B --> C[GetClientsHolder]
    C --> D[GetPidFromContainer]
    D --> E{pid found?}
    E -- no --> F[Errorf("...")]
    E -- yes --> G[Itoa(pid)]
    G --> H[ExecCommandContainer(nsenter cmd)]
    H --> I{success?}
    I -- success --> J[Return output]
    I -- fail --> K[Sleep(RetrySleepSeconds)]
    K --> L[RetryAttempts loop]
```

1. **Environment & Context**  
   * `GetTestEnvironment()` provides the current test configuration.  
   * `GetNodeProbePodContext()` fetches context needed to run commands on the node probe pod.

2. **Client Setup**  
   * `GetClientsHolder()` gives a client set capable of executing commands inside pods.

3. **PID Retrieval**  
   * `GetPidFromContainer(container)` looks up the process ID of the container’s main process using Docker inspect or CRI APIs.

4. **Command Construction & Execution**  
   * Build an `nsenter` command that attaches to the PID’s namespace:  
     ```bash
     nsenter -t <pid> -n -- <cmd>
     ```
   * Execute this via `ExecCommandContainer`, which runs it in the probe pod.

5. **Retry Logic**  
   * If execution fails, wait `RetrySleepSeconds` and retry up to `RetryAttempts`.  
   * On success, trim trailing newline and return output; otherwise propagate error.

### Dependencies

| Dependency | Role |
|------------|------|
| `GetTestEnvironment` | Retrieves test config (e.g., namespace). |
| `GetNodeProbePodContext` | Provides context for executing commands on node probe pod. |
| `Errorf` | Wraps errors with context. |
| `GetClientsHolder` | Supplies Kubernetes client interface. |
| `GetPidFromContainer` | Determines the target PID in container’s process tree. |
| `Itoa` | Converts integer PID to string for command line. |
| `ExecCommandContainer` | Runs arbitrary commands inside a pod. |
| `Sleep` | Implements retry delay. |

### Side‑Effects & Constraints

* Requires **node‑level access**: the probe pod must have privileges to run `nsenter`.  
* Relies on the container having a single main process whose PID can be retrieved reliably.  
* The function is *blocking* and may wait up to `(RetryAttempts × RetrySleepSeconds)` seconds before giving up.

### How it Fits the Package

`crclient` provides utilities for interacting with Kubernetes CRI environments (containers, pods, nodes).  
`ExecCommandContainerNSEnter` extends this by allowing tests to probe inside container namespaces without needing direct node SSH access. It is typically used in integration tests that need to inspect runtime state or configuration of a running container.

---
