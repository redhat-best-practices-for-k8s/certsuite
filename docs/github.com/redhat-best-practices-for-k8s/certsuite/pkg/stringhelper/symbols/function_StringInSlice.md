StringInSlice`

```go
func StringInSlice(slice []T, target T, ignoreCase bool) (found bool)
```

### Purpose  
`StringInSlice` determines whether a given value (`target`) exists in a slice of the same type (`slice`).  
It is a generic helper that works with any comparable type `T`, but it is primarily used for strings.  

### Parameters  

| Name        | Type   | Description |
|-------------|--------|-------------|
| `slice`     | `[]T`  | The collection to search. |
| `target`    | `T`    | Value being looked up. |
| `ignoreCase`| `bool` | If true, the comparison is case‑insensitive (only relevant when `T` is `string`). |

### Return value  

* `found` (`bool`) – `true` if `target` appears in `slice`; otherwise `false`.

### How it works  
1. **Trim** any leading/trailing whitespace from both `slice` elements and the `target`.  
2. If `ignoreCase` is set, convert each element and the target to a common case (lower‑case) before comparison.  
3. Iterate over the slice and compare each processed element with the processed target using the built‑in equality operator (`==`).  
4. Return `true` on first match; if no match is found, return `false`.

The function relies only on standard library helpers such as `strings.TrimSpace`, `strings.Contains`, and type conversion via `string()` when needed.

### Side effects  
- None. The function performs read‑only operations on its arguments and returns a value without modifying external state or globals.

### Package context  
`StringInSlice` lives in the **stringhelper** package (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/stringhelper`).  
This package groups small, reusable string utilities that are used throughout CertSuite for configuration parsing, flag handling, and output formatting. `StringInSlice` is a core helper for membership checks in user‑supplied lists (e.g., command‑line flags or config files).
