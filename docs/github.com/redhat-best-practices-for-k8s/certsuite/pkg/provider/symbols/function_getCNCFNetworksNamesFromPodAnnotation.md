getCNCFNetworksNamesFromPodAnnotation`

| Aspect | Details |
|--------|---------|
| **Signature** | `func getCNCFNetworksNamesFromPodAnnotation(annotation string) []string` |
| **Exported?** | No – helper used only inside the *provider* package. |

### Purpose
Extracts the list of network names from a pod’s CNCF‑compliant CNI annotation (`k8s.v1.cni.cncf.io/networks`).  
The annotation can be either:

1. A comma‑separated string: `<net>[,<net>...]`
2. A JSON array of objects, each containing at least a `"name"` field.

The function returns **only** the network names, discarding namespaces or other fields.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `annotation` | `string` | Raw value of the annotation to parse. It may contain whitespace and/or newlines. |

### Return Value
| Type | Description |
|------|-------------|
| `[]string` | Ordered slice of network names found in the annotation. If parsing fails or no names are present, an empty slice is returned (no error). |

### Implementation Details

1. **Trim & Detect Format**  
   * `strings.TrimSpace(annotation)` removes surrounding whitespace.  
   * If the trimmed string starts with `'['` it’s treated as JSON; otherwise it’s a simple list.

2. **JSON Path**  
   * `json.Unmarshal([]byte(...), &obj)` parses into an array of generic maps (`[]map[string]interface{}`).  
   * Each map is inspected for a `"name"` key, which is appended to the result slice after trimming whitespace.

3. **Comma‑Separated Path**  
   * `strings.Split(annotation, ",")` splits on commas.  
   * Every part is trimmed and added to the result slice.

4. **No Error Propagation** – Any unmarshalling failure or missing `"name"` keys simply results in those entries being ignored; the function always returns a slice (possibly empty).

### Dependencies
* `encoding/json.Unmarshal`
* `strings.TrimSpace`, `strings.Split`, `append`

These are standard library functions; no external packages are involved.

### Side Effects
None. The function is pure: it only reads its input and produces a new slice. No global state or I/O is touched.

### Role in the Package
The *provider* package orchestrates pod‑level checks for certsuite.  
When inspecting a pod, the provider needs to know which CNI networks are attached so that connectivity tests can be scoped appropriately. `getCNCFNetworksNamesFromPodAnnotation` supplies this information by parsing the annotation stored on each pod.

```mermaid
flowchart TD
    A[Pod] -->|has annotation| B[getCNCFNetworksNamesFromPodAnnotation]
    B --> C{Parse format}
    C -- JSON --> D[Extract names from objects]
    C -- List --> E[Split by comma]
    D & E --> F[Return []string]
```

This helper keeps the parsing logic isolated, making the rest of the provider code cleaner and easier to test.
