FilterArray`

| Symbol | Details |
|--------|---------|
| **Package** | `arrayhelper` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/arrayhelper`) |
| **Signature** | `func FilterArray(list []string, f func(string) bool) []string` |
| **Exported** | Yes |

### Purpose
`FilterArray` is a small utility that extracts all elements from a slice of strings that satisfy a user‑supplied predicate.  
It is the canonical “filter” helper for this repository and is used throughout `certsuite` wherever string slices need to be trimmed based on dynamic conditions (e.g., filtering out ignored certificates, selecting only specific labels, etc.).

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `list` | `[]string` | The source slice from which elements are selected. It may be empty or nil; the function handles both cases gracefully. |
| `f` | `func(string) bool` | A predicate that receives each element of `list`. If it returns `true`, the element is kept in the result; otherwise it is discarded. |

### Return Value
- `[]string`: a new slice containing only those elements for which `f` returned `true`.  
  The returned slice is always non‑nil (it will be empty if no elements match).

### Implementation Notes & Dependencies
```go
func FilterArray(list []string, f func(string) bool) []string {
    result := make([]string, 0, len(list))
    for _, v := range list {
        if f(v) {
            result = append(result, v)
        }
    }
    return result
}
```
- **`make`**: Allocates the underlying array with a capacity equal to `len(list)` to avoid repeated reallocations.  
- **Loop over `list`**: Standard range iteration; no mutation of the input slice occurs.  
- **Predicate call (`f(v)`)**: The function supplied by the caller decides inclusion logic.  
- **`append`**: Adds matched elements to `result`. Because we pre‑allocate with `len(list)` capacity, most calls will be O(1).

### Side Effects
- No global state is read or modified.
- Does not alter the original `list`; it only reads from it.
- The returned slice may share the underlying array with the input if no elements are filtered out (since capacity is set to `len(list)`). This is harmless because the function never writes back to that backing array.

### Usage Pattern
```go
// Keep only strings starting with "cert"
filtered := FilterArray(allStrings, func(s string) bool {
    return strings.HasPrefix(s, "cert")
})
```

### Placement in Package
`arrayhelper` provides tiny helpers that avoid boilerplate across the codebase.  
`FilterArray` sits at the core of these helpers—any other function that needs to prune a slice can delegate to it, keeping the package focused and easy to test.

### Mermaid Flow (Optional)
```mermaid
flowchart TD
    A[Input list] --> B{Iterate}
    B -->|f(v)=true| C[Append v to result]
    B -->|f(v)=false| D[Skip]
    C --> E[Return result]
```

---

**TL;DR:**  
`FilterArray` is a pure, reusable filter for string slices that returns all elements satisfying a user‑supplied predicate. It’s safe, side‑effect free, and part of the core utilities in `arrayhelper`.
