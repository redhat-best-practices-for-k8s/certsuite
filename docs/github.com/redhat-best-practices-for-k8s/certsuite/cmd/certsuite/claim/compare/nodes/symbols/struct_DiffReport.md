DiffReport` – Summary of Node‑Level Claim Differences  

| Element | Description |
|---------|-------------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/nodes` |
| **Purpose** | Holds a concise view of differences between two claim files at the node level.  It aggregates per‑section diffs (CNI, CSI, Hardware) and a list of individual node reports (`Nodes`). |

### Fields

| Field   | Type          | Meaning |
|---------|---------------|---------|
| `CNI`    | `*diff.Diffs` | Diff between the CNI sections of both claim files. |
| `CSI`    | `*diff.Diffs` | Diff between the CSI sections of both claim files. |
| `Hardware` | `*diff.Diffs` | Diff between the Hardware sections of both claim files. |
| `Nodes`  | `*diff.Diffs` | Aggregated diff of all node‑specific reports (one per node that appears in *both* claim files). |

> **Note**: Each field is a pointer to a `diff.Diffs` value, allowing the struct to represent “no differences” by being `nil`.

### Key Methods

#### `String() string`
```go
func (d DiffReport) String() string
```

* Implements `fmt.Stringer`.  
* Produces a table‑style text representation of the report.  
  * For each node that exists in both claim files, it shows the diff per section.  
  * If a node appears only in one file, it is flagged as `"not found in claim[1|2]"`.  

The method internally calls `String()` on the underlying `diff.Diffs` objects for each field.

### Construction

#### `GetDiffReport(nodes1, nodes2 *claim.Nodes) *DiffReport`
```go
func GetDiffReport(n1, n2 *claim.Nodes) *DiffReport
```

* **Inputs**: Two pointers to `claim.Nodes`, representing the parsed claim files.  
* **Process**:
  1. Calls a package‑level `Compare` helper for each section (CNI, CSI, Hardware, Nodes).  
  2. Each call returns a `diff.Diffs` object summarizing differences.  
  3. These results are stored in the corresponding fields of a new `DiffReport`.  
* **Output**: Pointer to a fully populated `DiffReport`.

### Interaction with the Package

```mermaid
flowchart TD
    A[claim.Nodes (file 1)] -->|Compare| B[diff.Diffs CNI]
    A -->|Compare| C[diff.Diffs CSI]
    A -->|Compare| D[diff.Diffs Hardware]
    A -->|Compare| E[diff.Diffs Nodes]

    F[claim.Nodes (file 2)] -->|Compare| B
    F -->|Compare| C
    F -->|Compare| D
    F -->|Compare| E

    B & C & D & E --> G[DiffReport]
    G --> H[String() output]
```

* The `GetDiffReport` function is the entry point that orchestrates these comparisons.  
* The resulting `DiffReport` can be printed, logged, or further processed by other parts of the certsuite toolchain.

### Side Effects & Dependencies

| Aspect | Details |
|--------|---------|
| **External packages** | Relies on `github.com/redhat-best-practices-for-k8s/certsuite/pkg/diff` for diff representation and string formatting. |
| **No I/O** | Purely computational; does not read/write files or network resources. |
| **Thread safety** | The struct is immutable after construction; safe to share across goroutines. |

### Summary

`DiffReport` is the central data structure that packages per‑section node differences into a single, printable object. It is created by `GetDiffReport`, which compares two claim files and populates each field with a `diff.Diffs`. The `String()` method then formats these diffs for human consumption, marking missing nodes appropriately. This struct enables downstream tooling (e.g., CLI output, tests) to consume a uniform view of node‑level discrepancies.
