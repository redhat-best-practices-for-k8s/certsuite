isNetworkAttachmentDefinitionConfigTypeSRIOV`

| Item | Detail |
|------|--------|
| **Signature** | `func isNetworkAttachmentDefinitionConfigTypeSRIOV(cniCfg string) (bool, error)` |
| **Visibility** | Unexported – used only inside the *provider* package. |
| **Purpose** | Determines whether a given CNI configuration JSON contains any plugin of type `"sriov"`.  The function is meant to be called with the raw `config` field from a Kubernetes `NetworkAttachmentDefinition` (NAD). |

### How it works

1. **Unmarshal the JSON**  
   The input string is unmarshaled into an interface (`map[string]interface{}`) using `json.Unmarshal`.  
   *If the string is not valid JSON, an error is returned.*

2. **Inspect top‑level fields**  
   - If the map contains a `"type"` key, its value is compared against `"sriov"`.  
     - Match → return `(true, nil)`.
     - No match → continue to next step.
   - If the map contains a `"plugins"` key (the multi‑plugin form), it must be an array of objects.  
     Each element is examined: if any has `"type":"sriov"`, the function returns `true`.

3. **Return result**  
   *If no `"sriov"` plugin is found, return `(false, nil)`.*  
   Any error during unmarshaling or type assertions results in an `(false, err)` pair.

### Key dependencies

| Dependency | Role |
|------------|------|
| `encoding/json.Unmarshal` | Parses the CNI config string. |
| `fmt.Errorf`, `log.Debugf` (from the package’s logger) | Produce human‑readable error and debug messages. |

### Side effects

*None.*  
The function only reads its input, logs diagnostic information via the package logger, and returns a boolean flag plus an optional error.

### Context within the *provider* package

The provider package is responsible for gathering runtime data about a Kubernetes cluster (nodes, pods, etc.) and exposing it to the CertSuite framework.  When evaluating whether a pod or node has SR‑IOV networking enabled, the provider needs to inspect the CNI configuration attached to that resource.

`isNetworkAttachmentDefinitionConfigTypeSRIOV` is called by higher‑level helpers such as `checkPodCniPlugins()` (not shown in the snippet) to decide if SR‑IOV-related checks should run.  It abstracts away the JSON parsing logic so that other parts of the package can simply ask “does this CNI config involve SR‑IOV?” without dealing with the details of the two possible CNI formats.

### Example usage

```go
cniJSON := `{
    "cniVersion": "0.4.0",
    "name": "sriov-network",
    "plugins": [
        {"type":"sriov","device":"eth1"},
        {"type":"firewall"}
    ]
}`
hasSRIOV, err := isNetworkAttachmentDefinitionConfigTypeSRIOV(cniJSON)
if err != nil {
    log.Errorf("invalid CNI config: %v", err)
}
fmt.Println(hasSRIOV) // true
```

This function is a small but crucial building block for CertSuite’s networking validation logic.
