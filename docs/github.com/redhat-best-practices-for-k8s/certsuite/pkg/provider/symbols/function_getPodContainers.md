getPodContainers`

| | |
|---|---|
| **Signature** | `func(pod *corev1.Pod, isDaemon bool) []*Container` |
| **Visibility** | unexported (internal helper) |
| **File/Line** | `provider.go:424` |

### Purpose
Collects the containers that should be examined for a given pod.  
The function filters out containers that are ignored by name, enriches each container with its image source information and logs warnings if the pod is empty.

### Parameters

| Name | Type | Description |
|------|------|-------------|
| `pod` | `*corev1.Pod` | The pod from which to extract container definitions. |
| `isDaemon` | `bool` | Indicates whether the pod is a DaemonSet (`true`) or not (`false`). This flag is used only for logging purposes when a pod has no containers. |

### Return Value
- A slice of pointers to `Container` structs (see *containers.go*).  
  Each element represents a container that should be considered during runtime image checks.

### Key Steps

1. **Check empty pod** – If the pod contains no containers (`len(pod.Spec.Containers) == 0`) a warning is logged using `Warn`. The log message distinguishes between normal pods and DaemonSets.
2. **Iterate over declared containers** – For each container in `pod.Spec.Containers`:
   - Skip it if its name matches any string in the global slice `ignoredContainerNames` (checked by `HasIgnoredContainerName`).  
     This allows tests to ignore system‑provided or sidecar containers that are not relevant for image source validation.
   - Build a `Container` value using `buildContainerImageSource`, which attaches runtime UID information via `GetRuntimeUID`.
   - Append the resulting container to the output slice.

### Dependencies

| Function | Role |
|----------|------|
| `len` | Check pod container count. |
| `Warn` | Emit log messages for empty pods or ignored containers. |
| `HasIgnoredContainerName` | Determine if a container should be skipped. |
| `GetRuntimeUID` | Retrieve the runtime UID of the pod (used inside `buildContainerImageSource`). |
| `buildContainerImageSource` | Construct the final `Container` struct with image source metadata. |

### Side‑Effects
- Logs warnings via `Warn`; no other state is modified.
- No mutation of the input `pod`.

### Package Context
`getPodContainers` is a helper used by higher‑level provider functions that iterate over all pods in a cluster (e.g., when performing image source validation or connectivity checks). By isolating container extraction and filtering logic here, the rest of the package can focus on orchestration while this function guarantees consistent container selection across different pod types.
