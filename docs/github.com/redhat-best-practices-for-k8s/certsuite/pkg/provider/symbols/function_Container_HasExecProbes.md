Container.HasExecProbes()`

| Aspect | Detail |
|--------|--------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider` |
| **Receiver** | `c Container` – a struct representing a container definition in a pod. |
| **Signature** | `func (c Container) HasExecProbes() bool` |

### Purpose
`HasExecProbes` determines whether the container defines any *exec*‑based lifecycle probes (`livenessProbe`, `readinessProbe`, or `startupProbe`).  
In certsuite, containers with exec probes are treated specially during health‑check testing because the probe command may affect node state or require privileged execution.

### Inputs
- The method operates on a **value receiver** `c Container`.  
  It accesses the container’s embedded Kubernetes pod specification fields:
  - `LivenessProbe`
  - `ReadinessProbe`
  - `StartupProbe`

No external inputs are required; all data is read from the receiver.

### Outputs
- Returns a single `bool`:
  - **true** – at least one of the probes contains an exec action (`Exec` field non‑nil).
  - **false** – none of the probes use an exec command (either no probe or only HTTP/TCPSocket probes).

### Key dependencies & side effects
- **No external globals** are read; the function is pure.  
- Relies on the Kubernetes API types (`v1.Probe`) to inspect `Exec` fields.
- No state mutation occurs.

### How it fits the package
The `provider` package models OpenShift/Kubernetes objects (nodes, pods, containers).  
During test execution, certsuite iterates over all containers in a pod and uses `HasExecProbes()` to decide whether to:

1. Skip certain connectivity checks that would interfere with exec probes.
2. Apply special handling for privileged probe commands.

Thus, this method is a small but essential utility that enables higher‑level logic to be aware of container health‑check configurations without duplicating probe inspection code across the package.
