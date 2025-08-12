DiffReport.String` – Table‑style diff renderer

| Item | Details |
|------|---------|
| **Package** | `nodes` (`github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/nodes`) |
| **Receiver type** | `DiffReport` (struct holding per‑node comparison results) |
| **Signature** | `func (d DiffReport) String() string` |
| **Exported?** | Yes – implements the `fmt.Stringer` interface. |

### Purpose
The method turns a `DiffReport` into a human‑readable table that lists, for each node appearing in both claim files, whether its attributes match or differ.  
If a node exists only in one of the two claims, it is annotated with *“not found in claim[1|2]”*.

### Inputs / Outputs
- **Input** – The `DiffReport` instance (`d`).  
  It must already contain the per‑node comparison data (likely maps or slices that were populated during the claim‑file parsing stage).
- **Output** – A single string containing a formatted table.  
  Each row corresponds to one node, with columns for:
  1. Node name / identifier
  2. Status of presence in each claim (`found`/`not found`)
  3. Result of attribute comparison (e.g., `match`, `differs`, or detailed diff).

### Key Dependencies & Calls
The method calls the standard library’s `fmt.Stringer`‑compatible helpers, such as:

- `String()` from `fmt` – used to format values.
- Likely `strings.Builder` for efficient string concatenation (not shown but typical).
  
No external globals are referenced; all data comes from the receiver.

### Side Effects
The method is **pure**: it does not mutate the `DiffReport`. It only reads its fields and returns a new string. Therefore, calling it repeatedly will yield the same result unless the underlying report changes.

### Package Context
`nodes` contains logic for comparing Kubernetes claim files at the node level.  
- Other functions in this package build the `DiffReport` by walking two claims’ node lists.  
- Once constructed, callers (e.g., CLI commands) invoke `String()` to display results in a tabular format or redirect it to a file.

### Suggested Mermaid Diagram
```mermaid
flowchart TD
    A[Claim File 1] -->|parse nodes| B[DiffReport.nodes1]
    C[Claim File 2] -->|parse nodes| D[DiffReport.nodes2]
    B & D --> E[Compare each node]
    E --> F[Populate DiffReport.entries]
    F --> G[DiffReport.String() → table output]
```

This method is the final step in presenting comparison results to users.
