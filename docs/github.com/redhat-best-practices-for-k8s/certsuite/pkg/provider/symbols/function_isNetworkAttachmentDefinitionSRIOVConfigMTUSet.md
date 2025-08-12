isNetworkAttachmentDefinitionSRIOVConfigMTUSet`

**File:** `pkg/provider/pods.go` (line 291)  
**Package:** `provider`

### Purpose
Determines whether a **NetworkAttachmentDefinition** (NAD) that uses the SR‑I/O‑V CNI plugin has an explicit MTU value set in its configuration.  

In OpenShift / Kubernetes environments, many tests expect an MTU of 1500 for SR‑I/O‑V interfaces. This helper parses the NAD JSON payload and checks the `mtu` field inside the `sriov` plugin block.

### Signature
```go
func isNetworkAttachmentDefinitionSRIOVConfigMTUSet(jsonStr string) (bool, error)
```

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `jsonStr` | `string` | Raw JSON representation of a NetworkAttachmentDefinition. |

| Return | Type   | Description |
|--------|--------|-------------|
| `bool` | true if an `mtu` field exists in the SR‑I/O‑V plugin block; false otherwise. |
| `error` | Any parsing or validation error encountered during processing. |

### How It Works
1. **Unmarshal**  
   The function starts by unmarshalling `jsonStr` into a generic map (`map[string]interface{}`) using Go’s `encoding/json`.  
2. **Navigate the Plugin Array**  
   * It extracts the `"plugins"` array and iterates over each element.  
3. **Identify SR‑I/O‑V Plugin**  
   For each plugin object, it checks if the `"type"` key equals `"sriov"`.  
4. **Check MTU Field**  
   When an SR‑I/O‑V plugin is found, it looks for a numeric `"mtu"` entry.  
   * If present, it returns `true, nil`.  
   * If absent, the function continues searching other plugins.
5. **Return Result**  
   After inspecting all plugins:  
   * If no SR‑I/O‑V plugin or no MTU was found → `false, nil`.  
6. **Error Handling**  
   Any JSON unmarshalling error or type assertion failure results in an `error` returned via `fmt.Errorf`.

### Dependencies
- **Standard library**
  - `encoding/json` – for unmarshalling the NAD JSON.
  - `fmt` – used to construct error messages (`Errorf`) and debug logs.
- **Logging**  
  The function calls a package‑level `Debug` helper (likely from a logger) when encountering parsing issues.

### Side Effects
None. The function is read‑only: it never mutates the input string or any global state.

### Package Context
Within the `provider` package, this helper supports pod and node validation tests that need to confirm SR‑I/O‑V configurations are correctly set. It is used by higher‑level routines that iterate over NADs discovered in a cluster, ensuring compliance with expected networking policies.
