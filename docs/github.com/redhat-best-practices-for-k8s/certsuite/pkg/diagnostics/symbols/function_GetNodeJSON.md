GetNodeJSON` – Package Diagnostics

| Item | Detail |
|------|--------|
| **Signature** | `func GetNodeJSON() map[string]interface{}` |
| **Exported?** | Yes |
| **Purpose** | Retrieve a structured representation of the current Kubernetes node set, analogous to the output of `oc get nodes -o json`. |

### Overview
`GetNodeJSON` acts as a thin wrapper around the *Test Environment* helper (`GetTestEnvironment`).  
It asks the environment for the raw node list (as returned by the OpenShift client), then marshals and unmarshals that data into an untyped `map[string]interface{}`. This map can be consumed directly by callers or further processed into a typed struct.

### Workflow
1. **Fetch Raw JSON**  
   Calls `GetTestEnvironment()` which returns a JSON string describing the node list.
2. **Round‑Trip Conversion**  
   - `json.Marshal` turns the raw string into bytes (no state change).  
   - `json.Unmarshal` decodes those bytes back into a generic map (`map[string]interface{}`).
3. **Return Value**  
   The resulting map is returned to the caller.

### Dependencies
- **GetTestEnvironment** – Provides the node list as JSON text.
- **encoding/json** – Used for marshaling/unmarshaling.
- **logrus** (via `Error`) – Logs any serialization errors; no panic or error return path, so callers must rely on log output.

### Side Effects
- Writes to the logger if marshaling/unmarshaling fails.  
- No other state changes; function is pure aside from logging.

### Integration in Package
`diagnostics` aggregates various system‑level checks (CPU, memory, PCI devices, etc.). `GetNodeJSON` supplies node metadata needed by higher‑level diagnostic routines and tests that require introspection of the cluster’s nodes. It mirrors the standard OpenShift command output, making it convenient for tools that expect that format.

### Suggested Diagram
```mermaid
flowchart TD
    A[Caller] --> B(GetNodeJSON)
    B --> C{GetTestEnvironment}
    C --> D[Raw JSON string]
    D --> E[json.Marshal → bytes]
    E --> F[json.Unmarshal → map[string]interface{}]
    F --> G[Return to Caller]
```

---
