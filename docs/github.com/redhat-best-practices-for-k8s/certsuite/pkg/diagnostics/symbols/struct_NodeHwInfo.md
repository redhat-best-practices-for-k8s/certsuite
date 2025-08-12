NodeHwInfo`

> **Package**: `github.com/redhat-best-practices-for-k8s/certsuite/pkg/diagnostics`  
> **File**: `diagnostics.go` (line 49)  

### Purpose
`NodeHwInfo` is a lightweight container that aggregates the raw hardware‑inspection output collected from a Kubernetes node.  
It is used by the diagnostics subsystem to expose per‑node hardware details in a structured form, which can then be rendered as JSON or plain text.

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `IPconfig` | `interface{}` | Raw output of an IP configuration command (e.g., `ip addr`). The type is left generic so that the diagnostic code can store either a string, a structured map, or any other representation that may be produced by the underlying helper. |
| `Lsblk`     | `interface{}` | Result of the `lsblk` command. Like `IPconfig`, it is kept as an empty interface to allow flexible unmarshalling of JSON or plain text. |
| `Lscpu`     | `interface{}` | Output from `lscpu`. Stored generically for the same reason. |
| `Lspci`     | `[]string` | A slice of strings, each representing a line returned by `lspci`. This field is kept as a simple slice because the PCI output does not need complex parsing in the current diagnostics logic. |

> **Note**: All fields are exported so callers can freely read them after the hardware information has been collected.

### How It Is Populated

The public function that creates instances of this struct is `GetHwInfoAllNodes` (see below). The routine performs the following steps for each node in the cluster:

1. **Obtain environment** – Calls `GetTestEnvironment` to get context needed for command execution.
2. **Create a client holder** – Invokes `GetClientsHolder`, which supplies SSH or API clients capable of running commands on nodes.
3. **Run hardware queries** – For each node it runs the following utilities:
   - `ip addr` → stored in `IPconfig`
   - `lsblk`  → stored in `Lsblk`
   - `lscpu`  → stored in `Lscpu`
   - `lspci`  → parsed into a string slice and stored in `Lspci`
4. **Collect output** – The helper functions `getHWJsonOutput` (for JSON‑capable commands) and `getHWTextOutput` (for plain text) capture the command results and assign them to the corresponding struct fields.
5. **Error handling** – Any error during a command run triggers `logrus.Error`. The node’s entry may still be created with empty or partial data, but diagnostics will report the failure.

### Interaction With Other Package Elements

| Element | Relationship |
|---------|--------------|
| `GetHwInfoAllNodes` | Returns a `map[string]NodeHwInfo` mapping node names to their hardware info. The struct is the value type of this map. |
| `getHWJsonOutput`, `getHWTextOutput` | Helper functions that capture command output and populate `NodeHwInfo`. |
| Logging (`Error`) | Used for side‑effect reporting when a command fails. |

### Usage Example

```go
hwMap := diagnostics.GetHwInfoAllNodes()
for node, info := range hwMap {
    fmt.Printf("Node: %s\n", node)
    // IP configuration (may be raw JSON or string)
    fmt.Printf("IPconfig: %+v\n", info.IPconfig)

    // PCI devices
    for _, pciLine := range info.Lspci {
        fmt.Println(pciLine)
    }
}
```

### Diagram

```mermaid
graph TD
  subgraph Cluster
    node1((NodeA))
    node2((NodeB))
  end
  node1 -->|GetHwInfoAllNodes| mapA[map["node-a"] → NodeHwInfo]
  node2 -->|GetHwInfoAllNodes| mapB[map["node-b"] → NodeHwInfo]

  classDef struct fill:#f9f,stroke:#333,stroke-width:2px;
  class mapA,mapB struct;
```

### Summary

`NodeHwInfo` is a simple, exported data holder that collects per‑node hardware inspection results. It serves as the value type for the diagnostics mapping returned by `GetHwInfoAllNodes`, enabling callers to access node‑specific IP configuration, block device layout, CPU details, and PCI devices in a uniform structure. The struct itself has no behavior; its fields are populated solely through diagnostic helper functions that run system commands on each node.
