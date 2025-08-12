getCrScaleObjects`

| | |
|-|-|
| **Package** | `autodiscover` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/autodiscover`) |
| **Signature** | `func getCrScaleObjects(crs []unstructured.Unstructured, csv *apiextv1.CustomResourceDefinition) []ScaleObject` |
| **Exported** | No (private helper) |

### Purpose
Collects Kubernetes *Scale* objects for a set of Custom Resources (`crs`) that belong to the same CustomResourceDefinition as `csv`.  
The function is used during automatic discovery of metrics‑scalable resources, ensuring that only CRs which have an accompanying Scale subresource are returned.

### Parameters

| Name | Type | Description |
|------|------|-------------|
| `crs` | `[]unstructured.Unstructured` | Slice of raw Custom Resource objects. They may be from any namespace and represent instances of a CRD. |
| `csv` | `*apiextv1.CustomResourceDefinition` | The CSV (ClusterServiceVersion) that defines the CRD type to filter on. Its group, version, and kind are used to match the CRs. |

### Return Value

| Type | Description |
|------|-------------|
| `[]ScaleObject` | A slice of `ScaleObject` values – one for each matching CR that exposes a Scale subresource. The function may return an empty slice if none are found or if an error occurs while retrieving the scale objects. |

### Key Steps & Dependencies

1. **Client Acquisition**  
   Uses `GetClientsHolder()` (from the package’s client‑factory) to obtain a `dynamic.Interface` capable of interacting with arbitrary resources.

2. **CRD Identification**  
   For each CR in `crs`, its GroupVersionKind is compared against that of `csv`. Only matching CRs are processed further.

3. **Namespace & Name Extraction**  
   Calls `GetName()` and `GetNamespace()` on the unstructured object to determine where the Scale resource lives.

4. **Scale Retrieval**  
   Invokes the dynamic client's `Scales(namespace).Get(name, metav1.GetOptions{})` to fetch the Scale subresource (a standard API that returns current replicas).

5. **Error Handling**  
   If any call fails (`TODO`, `Fatal`) the function logs a fatal error and exits – indicating that scale discovery is critical for the rest of the process.

6. **Result Construction**  
   For each successful fetch, it appends a `ScaleObject` (a local struct defined elsewhere in the package) to the result slice.

### Side Effects

* The function may terminate the program via `Fatal` if fetching a Scale object fails; this is intentional because missing scale information degrades discovery accuracy.
* It does **not** modify any of its input arguments or global state beyond reading from the dynamic client.

### How It Fits the Package

`autodiscover` aims to automatically locate all relevant resources for certificate and TLS analysis.  
Scale objects are essential for determining which workloads can be autoscaled (e.g., Deployments, StatefulSets, CRs with a Scale subresource). `getCrScaleObjects` bridges the gap between raw CR instances and their scaling metadata, feeding downstream logic that aggregates metrics or applies policies based on replica counts.

---

#### Suggested Mermaid Flow

```mermaid
flowchart TD
    A[Start] --> B{Iterate over crs}
    B -->|match CSV GVK| C[Extract name & namespace]
    C --> D[Get Scale via dynamic client]
    D -->|success| E[Append to result slice]
    D -->|failure| F[Fatal error (terminate)]
    E --> B
    F --> G[End]
```
