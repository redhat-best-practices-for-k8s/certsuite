GetScaleCrUnderTest`

| Aspect | Detail |
|--------|--------|
| **Package** | `autodiscover` (github.com/redhat-best-practices-for-k8s/certsuite/pkg/autodiscover) |
| **Exported?** | ✅ |
| **Signature** | `func GetScaleCrUnderTest(names []string, crds []*apiextv1.CustomResourceDefinition) []ScaleObject` |

### Purpose
Collects *scale* custom‑resource objects that are currently being used by the test suite.  
In CertSuite a “scale” CR is any Custom Resource Definition (CRD) that has a `Scale` subresource defined in its OpenAPI schema. The function returns a slice of `ScaleObject`, each describing one such CR and the namespace(s) where it is present.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `names` | `[]string` | Names of CRDs that should be considered for scaling tests. If empty, all discovered CRDs are examined. |
| `crds` | `[]*apiextv1.CustomResourceDefinition` | List of all CRDs present in the cluster (obtained earlier by discovery logic). |

### Return Value
`[]ScaleObject` – a slice where each element contains:
- The CRD name and group/version.
- The namespace(s) in which an instance of that CR exists.
- Any additional metadata required for later test stages.

The slice may be empty if no scaleable CRs are found or if errors occur during discovery.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `GetClientsHolder()` | Retrieves a shared client set used to query the API server. |
| `List`, `Namespace`, `Resource` (client-go helpers) | Build and execute list requests for CR instances across namespaces. |
| `getCrScaleObjects(crd)` | Parses a CRD’s OpenAPI schema to determine if it declares a `scale` subresource. |
| Logging utilities (`Info`, `Debug`, `Warn`, `Fatal`) | Emit diagnostic messages; may terminate the process on unrecoverable errors. |

### Algorithm (high‑level)

1. **Client Acquisition** – obtain a Kubernetes client from `GetClientsHolder`.  
2. **Filter CRDs** – iterate over the supplied `crds`; for each:
   * If `names` is non‑empty, skip any CRD whose name isn’t in that list.
   * Call `getCrScaleObjects(crd)` to see if a scale subresource exists.  
3. **Namespace Enumeration** – for each qualifying CRD:
   * List all namespaces where an instance of the CR exists (using the client’s dynamic interface).  
   * Record these namespaces in a `ScaleObject`.  
4. **Error Handling** – any fatal error during listing causes `Fatal` to log and exit; non‑fatal errors are logged via `Warn`.  
5. **Return** – the slice of all discovered `ScaleObject`s.

### Side Effects
- **Logging**: Emits info/debug/warn/fatal messages based on progress and failures.
- **No state mutation**: The function only reads from clients and CRD objects; it does not modify cluster resources or package globals.

### How It Fits the Package
`autodiscover` is responsible for introspecting a Kubernetes cluster to identify components relevant to CertSuite tests.  
`GetScaleCrUnderTest` specifically targets CRDs that support scaling, which are needed by the *scale* test suite (e.g., validating that scale subresources expose correct metrics). By filtering CRDs and collecting their namespaces, it supplies the rest of the test harness with the concrete objects to exercise.

---

#### Suggested Mermaid diagram

```mermaid
flowchart TD
    A[GetClientsHolder] --> B[Iterate over CRDs]
    B --> C{names filter?}
    C -- yes --> D[getCrScaleObjects(crd)]
    C -- no  --> D
    D --> E{scale subresource?}
    E -- no --> F[Skip]
    E -- yes --> G[List namespaces of CR instances]
    G --> H[Append to ScaleObject slice]
    H --> I[Return result]
```

This diagram visualises the main decision points and data flow within `GetScaleCrUnderTest`.
