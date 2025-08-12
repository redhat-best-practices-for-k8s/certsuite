createNodes` – Build a map of nodes for the provider

**File:** `pkg/provider/provider.go:649`  
**Package:** `provider`

### Purpose
`createNodes` transforms a slice of Kubernetes node objects (`[]corev1.Node`) into a **map keyed by node name** that contains enriched `Node` structures used internally by CertSuite.  
The map is later consumed by the rest of the provider logic (e.g., to decide which nodes need connectivity checks, to retrieve machine‑config data, or to annotate nodes for reporting).

### Inputs
| Parameter | Type | Description |
|-----------|------|-------------|
| `nodes` | `[]corev1.Node` | The raw node objects returned from the Kubernetes API. |

### Outputs
| Return value | Type | Description |
|--------------|------|-------------|
| `map[string]Node` | Map whose keys are node names and values are enriched `Node` structs (defined elsewhere in the package). | Provides quick lookup of a node by name and carries additional metadata such as role, labels, machine‑config status, etc. |

### Key Steps & Dependencies

1. **Cluster type detection**  
   - Calls `IsOCPCluster()` to check if the cluster is OpenShift.  
     *If true*, it logs a warning via `Warn` that “OpenShift clusters are not yet supported for this functionality.”

2. **Node filtering & enrichment**  
   For each node in the slice:
   - Extracts basic info (name, labels, taints, capacity).
   - Determines if the node is a master or worker by checking its labels against `MasterLabels` and `WorkerLabels`.
   - Calls `getMachineConfig(node)` to fetch the machine‑config status.  
     *If this call fails*, it logs an error with `Error`.

3. **Map construction**  
   - Adds each enriched node to a new map (`nodeMap`) keyed by its name.

4. **Logging**  
   - Uses `Info` to log how many nodes were processed.
   - Any errors during machine‑config retrieval are logged via `Error`.

### Side Effects
- No state is mutated outside the returned map; the function is pure with respect to package globals.
- Logging functions (`Warn`, `Error`, `Info`) produce side effects on the logger.

### How it fits the package
`createNodes` is a low‑level helper invoked by higher‑level provider initialization routines.  
It centralises node parsing and enrichment so that subsequent steps (e.g., connectivity checks, reporting) can operate on a consistent data structure without re‑parsing raw Kubernetes objects.

---

#### Suggested Mermaid diagram

```mermaid
flowchart TD
  A[Input: []corev1.Node] --> B{Is OCP?}
  B -- Yes --> C[Warn unsupported]
  B -- No --> D[Iterate nodes]
  D --> E[Extract metadata]
  E --> F[getMachineConfig(node)]
  F --> G[Add to map]
  G --> H[Return nodeMap]
```

This diagram illustrates the decision flow and the main data transformations performed by `createNodes`.
