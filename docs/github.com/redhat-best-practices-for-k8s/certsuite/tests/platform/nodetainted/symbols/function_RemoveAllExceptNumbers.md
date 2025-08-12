RemoveAllExceptNumbers`

| Attribute | Value |
|-----------|-------|
| **Package** | `nodetainted` (github.com/redhat-best-practices-for-k8s/certsuite/tests/platform/nodetainted) |
| **Exported?** | Yes |
| **Signature** | `func RemoveAllExceptNumbers(s string) string` |
| **File / Line** | `/Users/deliedit/dev/certsuite/tests/platform/nodetainted/nodetainted.go:131` |

### Purpose
`RemoveAllExceptNumbers` sanitises a string by stripping every character that is not a digit (`0‑9`).  
The cleaned string is returned for use in tests where numeric values are extracted from command output or log files.

### Parameters
| Name | Type   | Description |
|------|--------|-------------|
| `s`  | `string` | Input text that may contain letters, punctuation, whitespace, etc. |

### Return Value
* **`string`** – the input string with all non‑numeric characters removed.

### Implementation Details
```go
func RemoveAllExceptNumbers(s string) string {
    re := regexp.MustCompile(`[^0-9]`)
    return re.ReplaceAllString(s, "")
}
```
1. `regexp.MustCompile` compiles a regular expression that matches any character **not** in the range `0‑9`.  
2. `ReplaceAllString` replaces every match with an empty string, effectively deleting those characters.

The function relies on Go’s standard library `regexp` package; no global state is read or mutated.

### Side Effects
* None – pure functional transformation; input string remains unchanged and the function has no observable effect outside its return value.

### Package Context
In the **nodetainted** test suite, this helper is used to normalise command‑line output before asserting against expected numeric values (e.g., kernel version numbers).  
It complements `runCommand` (which executes shell commands) and `kernelTaints` (a slice of taint strings) by providing a lightweight way to isolate the numeric component of any textual data.  

### Suggested Mermaid Diagram
```mermaid
flowchart TD
    A[Input string] --> B{Regex: [^0-9]}
    B --> C[Replace matches with ""] --> D[Output string]
```
This diagram illustrates how non‑digits are removed to produce the final numeric string.
