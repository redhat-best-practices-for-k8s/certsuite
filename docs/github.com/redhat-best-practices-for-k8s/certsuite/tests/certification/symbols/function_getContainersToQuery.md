getContainersToQuery`

| Aspect | Detail |
|--------|--------|
| **Package** | `certification` (github.com/redhat-best-practices-for-k8s/certsuite/tests/certification) |
| **Visibility** | Unexported (`func getContainersToQuery`) – used only inside the test suite. |
| **Signature** | `func(*provider.TestEnvironment) map[provider.ContainerImageIdentifier]bool` |

### Purpose
Builds a *set* of container images that should be queried for certification status during a test run.

The function receives the current testing environment (`*provider.TestEnvironment`) and returns a mapping from image identifiers to a boolean flag.  
The returned `map` acts like a **hash set** – the keys are the images to query, and all values are simply `true`.  This makes subsequent look‑ups (`if _, ok := map[image]; ok { … }`) inexpensive.

### Inputs
| Parameter | Type | Role |
|-----------|------|------|
| `env` | `*provider.TestEnvironment` | Holds information about the current test environment (e.g., which operators are present, whether Helm charts were installed). The function inspects this object to decide which images need querying. |

### Output
- A map whose keys are `provider.ContainerImageIdentifier`.  
  Each key represents a container image that must be checked for certification status.

The boolean value is always `true` and is not used for any logic; the map is only used as a set.

### Key Dependencies & Side Effects
| Dependency | How it’s used |
|------------|---------------|
| `make` (built‑in) | Creates the result map (`map[provider.ContainerImageIdentifier]bool{}`). No other side effects. |

No global state is read or modified; the function purely transforms its argument into a new data structure.

### Context within the Package
In the certification test suite, tests need to know which container images have been deployed (via operators or Helm charts) so they can query each image’s certification status.  
`getContainersToQuery` encapsulates that logic:

1. It inspects the `TestEnvironment` for installed operators and chart releases.
2. For every relevant component it adds its image identifier to the returned map.

Other parts of the test suite then iterate over this map, calling a certification validator (e.g., `validator.QueryStatus`) for each image.

```mermaid
graph TD
    A[Test] --> B{getContainersToQuery(env)}
    B --> C[Map[ImageID → true]]
    C --> D[Loop over images]
    D --> E[validator.QueryStatus(image)]
```

### Summary
`getContainersToQuery` is a small, pure helper that turns the current test environment into a set of container image identifiers ready for certification status querying. It has no side effects and depends only on Go’s `make`.
