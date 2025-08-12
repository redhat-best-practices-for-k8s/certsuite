Container` – Representation of a Kubernetes container inside CertSuite

The **`Container`** type is the central data holder used by CertSuite’s provider package to
examine, test, and report on individual containers that run in a cluster.

| Field | Type | Purpose |
|-------|------|---------|
| `ContainerImageIdentifier` | `ContainerImageIdentifier` | Parsed information about the image (registry, repo, tag).  Used by pre‑flight checks to resolve image metadata. |
| `Namespace` | `string` | Namespace in which the pod that owns this container lives. |
| `NodeName` | `string` | Node name where the pod is scheduled. |
| `Podname` | `string` | Name of the owning pod. |
| `PreflightResults` | `PreflightResultsDB` | Stores results of pre‑flight checks performed on this container (e.g., run‑as‑non‑root, read‑only root FS). |
| `Runtime` | `string` | Container runtime used (`docker`, `containerd`, etc.).  Determines how to query the runtime for UID. |
| `Status` | `corev1.ContainerStatus` | Runtime status object returned by the API (image ID, state, ready flag, etc.). |
| `UID` | `string` | The *runtime* UID of the container, e.g. a Docker container ID or CRI‑ID.  Populated lazily via `GetUID`. |
| `embedded:*corev1.Container` | `*corev1.Container` | Embeds the Kubernetes container spec (`Image`, `Command`, `Args`, security context, probes, …).  The embedded field gives direct access to the spec fields without a name prefix. |

> **Why embed?**  
> The provider code often needs both the *spec* (what the user requested) and the *status* (what actually ran).  Embedding `*corev1.Container` allows callers to write `c.Image` instead of `c.Container`.  

## Key Methods

| Method | Signature | Role |
|--------|-----------|------|
| `GetUID() (string, error)` | `(string, error)` | Lazily parses `Status.ContainerID` into a runtime‑specific UID.  If the ID is missing or malformed, an error is returned.  Used by tests that need to identify containers in logs or metrics. |
| `HasExecProbes() bool` | `bool` | Checks whether any of the container’s liveness/readiness probes use the `exec` action.  Useful for detecting probe‑based sidecar injection. |
| `HasIgnoredContainerName() bool` | `bool` | Returns true if the container name is on the ignored list (e.g., Istio proxies).  The helper uses `IsIstioProxy()` and a global `ignoredContainers` slice. |
| `IsContainerRunAsNonRoot(*bool) (bool, string)` | `(bool, string)` | Evaluates `securityContext.runAsNonRoot`.  If the value is set to false, returns false with an explanatory message; otherwise true.  The boolean pointer allows callers to override defaults from pod level. |
| `IsContainerRunAsNonRootUserID(*int64) (bool, string)` | `(bool, string)` | Similar to above but checks `runAsUser`.  A user ID of 0 is considered insecure. |
| `IsIstioProxy() bool` | `bool` | Detects if the container name matches known Istio sidecar names (`istio-proxy`, `istiod`). |
| `IsReadOnlyRootFilesystem(*log.Logger) bool` | `bool` | Checks `securityContext.readOnlyRootFilesystem`.  Logs a warning if the flag is unset. |
| `IsTagEmpty() bool` | `bool` | Returns true when the image tag is empty (i.e., using `:latest`).  Useful for discouraging latest tags in production. |
| `SetPreflightResults(map[string]PreflightResultsDB, *TestEnvironment) error` | `error` | Populates `PreflightResults` by running pre‑flight checks via the provider’s test environment.  Handles Docker config resolution, insecure flags, and writes results into a database‑like map. |
| `String() string` | `string` | Human‑readable short description: `"containerName@imageTag"`. |
| `StringLong() string` | `string` | Verbose description including namespace, node, pod, image, status, and UID. |

## How it fits the package

1. **Creation** – `NewContainer()` simply allocates an empty struct; container objects are later populated in `getPodContainers`.
2. **Population** – `getPodContainers` iterates over a `corev1.Pod`, calling helpers to:
   * Build `ContainerImageIdentifier`
   * Resolve runtime UID via `GetRuntimeUID`
   * Skip containers on the ignored list
3. **Filtering** – The test environment uses helper methods such as `GetGuaranteedPodContainersWithExclusiveCPUs` that call container methods like `IsReadOnlyRootFilesystem`.
4. **Pre‑flight checks** – Each container runs a suite of pre‑flight tests (via `SetPreflightResults`) to verify security best practices.
5. **Reporting** – The string helpers enable concise log output and result summaries.

Overall, the `Container` struct is the *single point* where spec, status, runtime metadata, and test results converge, allowing CertSuite to reason about a container’s compliance in a Kubernetes cluster.
