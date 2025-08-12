DiffReport.String` – Human‑Readable Report

| Item | Details |
|------|---------|
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/configurations` |
| **Receiver type** | `DiffReport` (value receiver) |
| **Signature** | `func (d DiffReport) String() string` |
| **Purpose** | Produce a concise, readable representation of the differences found between two configuration objects. The method is intended for display in logs or command‑line output rather than for programmatic consumption. |

### Inputs & Outputs

* **Input** – A `DiffReport` value (`d`).  
  It holds information about which configuration items differ (e.g., lists of added, removed, or changed keys).  
  The method does not mutate the receiver.

* **Output** – A single string containing a formatted diff report.  
  If there are no differences, the output is `"No differences found."`.  
  Otherwise it concatenates the string representations of each difference category (added, removed, modified) using the `String()` methods defined on those underlying slice or struct types.

### Key Dependencies

| Dependency | How It’s Used |
|------------|---------------|
| `d.Added.String()` | Generates a string for added configuration items. |
| `d.Removed.String()` | Generates a string for removed configuration items. |
| `d.Modified.String()` | Generates a string for modified configuration items. |

These sub‑methods are defined on the corresponding types (`Added`, `Removed`, `Modified`) within the same package and return a formatted slice representation.

### Side Effects

* **None** – The method only reads data; it does not alter the `DiffReport` or any global state.

### Package Context

The `configurations` package compares two Kubernetes configuration objects to identify discrepancies.  
`DiffReport.String()` is the primary way callers present those findings to users, e.g., in CLI tools or log files. It encapsulates all formatting logic so that other parts of the program can simply call `report.String()` and receive a ready‑to‑display string.

### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[DiffReport] -->|String()| B[String output]
    A --> C[Added]
    A --> D[Removed]
    A --> E[Modified]
    C -->|String()| B
    D -->|String()| B
    E -->|String()| B
```

This diagram shows the `DiffReport` delegating to its component types for string conversion, then assembling the final output.
