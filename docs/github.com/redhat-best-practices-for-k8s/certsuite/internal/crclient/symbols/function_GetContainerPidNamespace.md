GetContainerPidNamespace`

| Feature | Details |
|---------|---------|
| **Package** | `crclient` (`github.com/redhat-best-practices-for-k8s/certsuite/internal/crclient`) |
| **Signature** | `func GetContainerPidNamespace(container *provider.Container, env *provider.TestEnvironment) (string, error)` |
| **Exported** | ✅ |

### Purpose

Retrieves the PID namespace identifier of a container that is running inside a Kubernetes pod.  
The function first obtains the target pod’s context via the test environment, then asks the kube‑client for the process ID of the container. Finally it executes `cat /proc/self/ns/pid` from within the container to read the namespace path.

### Inputs

| Parameter | Type | Role |
|-----------|------|------|
| `container` | `*provider.Container` | Describes which container’s PID namespace is needed (name, image, etc.). |
| `env` | `*provider.TestEnvironment` | Holds configuration and helper methods for communicating with the Kubernetes cluster. |

### Outputs

| Return | Type | Meaning |
|--------|------|---------|
| `string` | Namespace path (e.g., `/proc/<pid>/ns/pid`) | The namespace identifier of the container’s PID namespace. |
| `error` | error | Non‑nil if any step fails: context retrieval, PID extraction, command execution, or parsing. |

### Core Steps

1. **Get pod context**  
   ```go
   ctx := GetNodeProbePodContext(env)
   ```  
   Retrieves a *context.Context* that carries cancellation / timeout information for downstream client calls.

2. **Find container’s PID**  
   ```go
   pid, err := GetPidFromContainer(ctx, container, env)
   ```  
   Uses the CRClient to query the Kubernetes API and discover which host‑process ID is running the container.

3. **Execute `cat /proc/self/ns/pid` inside container**  
   ```go
   cmd := []string{"sh", "-c", fmt.Sprintf("cat /proc/%s/ns/pid", pid)}
   output, err := ExecCommandContainer(ctx, env.GetClientsHolder(), container.PodName, container.Name, cmd)
   ```  
   The command is run inside the pod via the kube‑client’s exec API. The output contains the full path to the PID namespace.

4. **Return parsed result**  
   The function trims whitespace and returns the string; if any step errors out, a descriptive wrapped error is returned.

### Dependencies & Side‑Effects

| Dependency | What it does |
|------------|--------------|
| `GetNodeProbePodContext` | Provides context for Kubernetes client calls. |
| `GetPidFromContainer` | Calls the CRClient API to map container name → host PID. |
| `ExecCommandContainer` | Executes a command inside a pod via kube‑client. |
| `GetClientsHolder` | Supplies the necessary client interfaces (e.g., REST, Exec). |

*Side‑effects:*  
The function only performs read operations against the cluster and executes a harmless shell command in the target container. No state is mutated.

### How It Fits the Package

`crclient` offers high‑level wrappers around Kubernetes CRUD and exec operations for test environments. `GetContainerPidNamespace` is a utility that ties together:

1. **Kubernetes client interaction** – to locate the container’s host PID.
2. **Exec capability** – to interrogate the container’s filesystem.

It is used by tests that need to compare or manipulate process namespaces, ensuring that containers are isolated as expected.

### Suggested Mermaid Diagram

```mermaid
flowchart TD
  A[GetNodeProbePodContext] --> B[GetPidFromContainer]
  B --> C{PID found?}
  C -- yes --> D[ExecCommandContainer("cat /proc/<pid>/ns/pid")]
  D --> E[Return namespace string]
  C -- no --> F[Error: PID not found]
```

This diagram illustrates the decision flow and external calls involved in obtaining a container’s PID namespace.
