getContainers`

| | |
|---|---|
| **Package** | `provider` |
| **Visibility** | Unexported (`func getContainers([]*Pod) []*Container`) |
| **Purpose** | Aggregate all container objects from a slice of pod descriptors, applying the package‑wide exclusion list. |

#### Function Signature
```go
func getContainers(pods []*Pod) []*Container
```

- **Input**  
  - `pods`: A slice containing pointers to `Pod` structures that represent Kubernetes pods discovered by the provider.

- **Output**  
  - Returns a new slice of pointers to `Container` objects.  
    The returned containers are all those that belong to the input pods, except any whose names appear in `ignoredContainerNames`.

#### Key Steps
1. Create an empty result slice (`containers := []*Container{}`).
2. Iterate over each pod in the provided slice.
3. For each pod, iterate over its contained `Containers` field.
4. If a container’s name is **not** present in the global map/array `ignoredContainerNames`, append it to the result slice.
5. Return the accumulated slice.

#### Dependencies
- **Types**  
  - `Pod`: defined elsewhere in the provider package; contains a slice of `Container` objects.  
  - `Container`: the container representation used throughout the provider logic.

- **Globals**  
  - `ignoredContainerNames`: a set (or map) of container names that should be omitted from all analyses. This list is populated during provider initialization and is consulted here to filter out system or side‑car containers that are not relevant for certificate validation.

#### Side Effects
- None. The function performs pure computation: it does **not** modify any input pods, the global state, or perform I/O.

#### Integration in the Package
`getContainers` is a helper used by various filter and validator functions to gather all containers that need to be examined for certificate compliance. By centralising the filtering logic (via `ignoredContainerNames`) it ensures consistent behaviour across the provider’s checks. Because it is unexported, only other files within the `provider` package can invoke it, keeping the public API focused on higher‑level operations.

#### Suggested Mermaid Diagram
```mermaid
flowchart TD
  A[Input: []*Pod] --> B{Iterate Pods}
  B --> C{Iterate Containers}
  C -->|Not ignored| D[Append to result]
  C -->|Ignored| E[Skip]
  D --> F[Return []*Container]
```

This succinctly shows the filtering loop and how containers are collected.
