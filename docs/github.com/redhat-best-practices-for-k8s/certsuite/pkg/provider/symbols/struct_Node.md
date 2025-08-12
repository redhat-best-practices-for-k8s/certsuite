Node` – a lightweight wrapper around a Kubernetes node

| Field | Type | Purpose |
|-------|------|---------|
| `Data` | `*corev1.Node` | The raw K8s API object that holds all information about the host (labels, annotations, spec, status). |
| `Mc`   | `MachineConfig` | Holds the machine‑config data for this node when it is part of an OpenShift cluster. It is populated during provider initialisation (`createNodes`). |

> **Why a wrapper?**  
> The provider uses `Node` to expose convenient predicates and helpers (e.g., *IsControlPlaneNode*, *GetRHELVersion*) that operate on the raw node data while keeping the API surface small and testable. It also allows marshaling the node back into JSON for logs or diagnostics.

---

### Key methods

| Method | Signature | What it does | Dependencies |
|--------|-----------|--------------|---------------|
| `MarshalJSON` | `func() ([]byte, error)` | Serialises the underlying `corev1.Node` (`Data`) to JSON. | Uses standard `encoding/json.Marshal`. |
| `IsControlPlaneNode`, `IsWorkerNode` | `func() bool` | Inspect the node’s labels/taints (via `StringInSlice`) to decide if it belongs to the control‑plane or worker pool. | Relies on helper `StringInSlice`. |
| `IsCSCOS`, `IsRHCOS`, `IsRHEL`, `IsRTKernel` | `func() bool` | Checks the node’s OS flavour by searching the `osimage` label/annotation for known substrings (e.g., `"rhcos"`, `"rhel"`). Uses `strings.Contains` and `strings.TrimSpace`. |
| `GetCSCOSVersion`, `GetRHCOSVersion`, `GetRHELVersion` | `func() (string, error)` | Pulls the version string from the node’s OS image label. For RHCOS it also trims to a short form via `GetShortVersionFromLong`. Returns an error if the node is not of the expected OS type. |
| `HasWorkloadDeployed` | `func([]*Pod) bool` | Scans a list of pods and returns true when at least one belongs to this node (`pod.Spec.NodeName == n.Data.Name`). |
| `IsHyperThreadNode` | `func(*TestEnvironment) (bool, error)` | Executes the command `cat /proc/cpuinfo | grep processor | wc -l` inside a pod that is scheduled on this node. Compares the reported thread count against the expected value stored in `TestEnvironment`. Returns an error if the exec fails or the counts differ. |

---

### How it fits into **provider**

* The provider creates a map of all nodes (`createNodes`) and stores each as a `Node` instance.
* Tests access nodes via methods such as `GetBaremetalNodes()` or by iterating over the internal node map.
* Node predicates are used throughout the test suite to filter for control‑plane, worker, RHEL/RHCOS/CSCOS machines, etc.
* The helper `IsHyperThreadNode` allows tests that need to assert CPU topology (e.g., ensuring hyper‑threading is enabled or disabled).

---

### Typical usage

```go
// Grab the node map from the environment
nodes := env.GetNodes()

for _, n := range nodes {
    // Only run on worker nodes running RHCOS
    if !n.IsWorkerNode() || !n.IsRHCOS() { continue }

    ver, err := n.GetRHCOSVersion()
    if err != nil { /* handle error */ }
    fmt.Printf("Worker %s runs RHCOS v%s\n", n.Data.Name, ver)
}
```

---

### Summary

`Node` is a thin abstraction that:

1. **Hides** the raw Kubernetes node API from most of the test logic.
2. Provides **intuitive predicates** for common checks (role, OS type, kernel).
3. Exposes **version extraction** helpers used by the test cases.
4. Keeps **serialization** straightforward via `MarshalJSON`.

Its design keeps the provider codebase clean and makes it easy to add new node‑specific utilities without touching the rest of the system.
