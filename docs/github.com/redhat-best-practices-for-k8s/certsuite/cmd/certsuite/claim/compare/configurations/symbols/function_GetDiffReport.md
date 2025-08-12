GetDiffReport`

**File:** `configurations.go` – line 45  
**Package:** `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/configurations`

---

## Purpose
`GetDiffReport` produces a summary of differences between two *certification* configuration sets.  
The function is the public entry point used by higher‑level tooling (e.g., CLI commands or tests) to obtain a `DiffReport` that captures added, removed, and modified configuration items.

## Signature
```go
func GetDiffReport(left, right *claim.Configurations) *DiffReport
```

| Parameter | Type                     | Description |
|-----------|--------------------------|-------------|
| `left`    | `*claim.Configurations` | The baseline configuration set. |
| `right`   | `*claim.Configurations` | The target configuration set to compare against `left`. |

### Return Value
- `*DiffReport`: A pointer to a struct that aggregates the comparison results.  
  If either input is `nil`, the function will still return a report (often empty).

## Key Dependencies

| Dependency | Role |
|------------|------|
| `Compare` | Core comparison logic; returns a slice of `ComparisonResult`. |
| Built‑in `len` | Determines how many items were added or removed by comparing slice lengths. |

> *Note:* The function does not modify its inputs; it only reads from them.

## Implementation Flow

1. **Invoke `Compare(left, right)`**  
   Retrieves a slice of per‑item comparison results (`ComparisonResult`) indicating status (e.g., added, removed, unchanged).

2. **Count Additions & Removals**  
   * Added items are identified by comparing the length of `right` against the number of matches found in `Compare`.  
   * Removed items are derived similarly from `left`.

3. **Populate `DiffReport`**  
   The report includes:
   - Total added/removed counts.
   - The raw comparison slice for further inspection or display.

4. **Return Report**  
   A pointer to the constructed `DiffReport` is returned to the caller.

## Side Effects & Constraints

- No global state is modified; the function is pure from an external‑view perspective.
- If inputs are `nil`, it behaves gracefully by treating them as empty configurations.
- Relies on `claim.Configurations` and related types defined elsewhere in the `claim` package.

## Usage Context
The function sits at the heart of the *comparison* sub‑package, providing a convenient API for other components (e.g., CLI commands or unit tests) to obtain diff information without delving into the lower‑level comparison mechanics. It abstracts away slice handling and counting logic, presenting callers with a ready‑to‑consume report.

---

### Suggested Mermaid Diagram
```mermaid
flowchart TD
    A[GetDiffReport] --> B{Inputs}
    B --> C[left]
    B --> D[right]
    A --> E[Call Compare(left,right)]
    E --> F[ComparisonResult[]]
    A --> G[Count additions/removals]
    G --> H[Build DiffReport]
    H --> I[Return *DiffReport]
```

This diagram visualizes the data flow from input configurations through comparison to the final report.
