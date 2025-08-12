GetHwInfoAllNodes`

> **Purpose**  
> Collect hardware‑level diagnostics for every node in the current test environment and return them as a map keyed by node name.

## Signature
```go
func GetHwInfoAllNodes() (map[string]NodeHwInfo)
```
* Returns a mapping from node names (`string`) to `NodeHwInfo` structs that hold parsed hardware information.
* No error is returned – failures are logged via the package’s `Error` helper and the corresponding entry in the map will contain an empty or partially‑filled `NodeHwInfo`.

## Dependencies

| Called function | Role |
|-----------------|------|
| `GetTestEnvironment()` | Provides a list of node names that belong to the current test environment. |
| `GetClientsHolder()` | Gives access to a client that can execute commands on nodes via SSH or an API. |
| `getHWJsonOutput(client, cmd)` | Runs a system‑information command (e.g., `lscpu`, `lsblk`) and returns its JSON output. |
| `getHWTextOutput(client, cmd)` | Runs the CNI plugins command (`cni-plugins`) and returns plain text. |
| `Error(msg string, args ...interface{})` | Logs diagnostics failures; used for each node that cannot be queried or when parsing fails. |

The helper `make([]byte, 0)` is used internally to build output buffers.

## Algorithm Overview

1. **Environment Setup**  
   * Retrieve the current test environment’s node list.
   * Obtain a client holder capable of running commands on those nodes.

2. **Iterate Nodes**  
   For each node name:
   * Initialize an empty `NodeHwInfo` instance.
   * Execute hardware queries in parallel (or sequentially – code not shown) using the above helpers:
     - CPU (`lscpu`)
     - Disk layout (`lsblk`)
     - PCI devices (`lspci`)
     - Network interfaces (`ip`)
     - CNI plugins list
   * Each command’s output is parsed into a struct field of `NodeHwInfo`.  
     Parsing errors trigger an `Error` log but do not abort the loop.

3. **Collect Results**  
   Store the populated `NodeHwInfo` in the result map under its node name.

4. **Return Map**  
   The function returns after all nodes have been processed, providing a complete view of hardware across the test cluster.

## Side Effects

* Logs errors for any node that cannot be queried or parsed.
* Does **not** modify global state; all data is local to the function and its return value.

## Package Context

`GetHwInfoAllNodes` lives in `pkg/diagnostics`.  
The package aggregates various diagnostic utilities (e.g., network, storage, security checks). This function specifically supplies a snapshot of low‑level hardware characteristics that other diagnostics may consume or display. It is typically invoked during test setup or when generating a full diagnostics report.

---

**Mermaid diagram suggestion**

```mermaid
flowchart TD
    A[GetTestEnvironment] --> B{Node List}
    B --> C[Iterate Nodes]
    C --> D[getHWJsonOutput lscpu] --> E[Parse CPU]
    C --> F[getHWJsonOutput lsblk] --> G[Parse Disk]
    C --> H[getHWJsonOutput lspci] --> I[Parse PCI]
    C --> J[getHWJsonOutput ip]   --> K[Parse Network]
    C --> L[getHWTextOutput cni-plugins] --> M[Parse CNI]
    D & G & I & K & M --> N[Populate NodeHwInfo]
    N --> O{Add to Map}
    O --> P[Return map[string]NodeHwInfo]
```

This diagram visualizes the per‑node workflow and how command outputs feed into the `NodeHwInfo` struct.
