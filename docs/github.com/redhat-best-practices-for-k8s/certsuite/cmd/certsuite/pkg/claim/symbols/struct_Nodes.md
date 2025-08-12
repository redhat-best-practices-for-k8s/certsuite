Nodes` – High‑level representation of node information in the *claim* package

| Element | Description |
|---------|-------------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/pkg/claim` |
| **Location** | Defined at line 60 of `claim.go`. |

### Purpose
The `Nodes` struct aggregates diverse data that a *certsuite* claim may need about the cluster nodes.  
It is used as an intermediate container when serializing/deserializing node‑related claims, and serves as the source of truth for other claim components (e.g., `NodeSummary`, `CniNetworks`, etc.) that consume node information.

### Fields
| Field | Type | Typical content |
|-------|------|-----------------|
| `NodesHwInfo` | `interface{}` | Raw hardware metadata extracted from nodes (CPU, memory, NICs). The concrete type is usually a slice or map of custom structs. |
| `NodesSummary` | `interface{}` | A condensed view of node status (ready/not‑ready, taints, labels). Often a struct with aggregated counts. |
| `CniNetworks` | `interface{}` | Information about the Container Network Interface networks attached to nodes (CIDR ranges, plugin names). |
| `CsiDriver` | `interface{}` | Data about CSI drivers installed on the cluster and their node‑level capabilities. |

> **Note:** All fields are `interface{}` because the claim format is flexible; concrete types are injected by other parts of the program (e.g., JSON unmarshaling or internal helpers). The struct itself does not enforce any schema.

### Dependencies
- **JSON marshaler/unmarshaler** – When a claim file is read, `encoding/json` populates these fields with concrete data structures.
- **Other claim types** – Functions that generate or validate node‑related claims (e.g., `GenerateNodeSummaryClaim`) reference this struct to fetch the needed information.

### Side Effects
The struct is *pure*—it holds data and does not perform actions. Any mutation happens outside of the struct (e.g., after unmarshaling). Because fields are exported, other packages can read or replace them freely, but they should maintain the expected type contract for downstream consumers.

### Role in the Package
`Nodes` acts as a **central repository** for all node‑level data that claims might reference. It is passed around to:

1. **Claim generators** – Assemble claim JSON from raw cluster state.
2. **Validators** – Cross‑check claim consistency against live cluster metrics.
3. **Reporters** – Output human‑readable summaries of node status.

By keeping all node data in one struct, the package achieves a clear separation between *data acquisition* (e.g., via Kubernetes API calls) and *claim processing*, improving maintainability and testability.

---

#### Suggested Mermaid diagram

```mermaid
classDiagram
    class Nodes {
        +NodesHwInfo interface{}
        +NodesSummary interface{}
        +CniNetworks interface{}
        +CsiDriver interface{}
    }
    class ClaimGenerator
    class Validator
    class Reporter
    Nodes --> ClaimGenerator : provides data
    Nodes --> Validator : validates against live state
    Nodes --> Reporter : outputs summaries
```

This diagram visualizes how `Nodes` feeds into the main claim‑handling workflows.
