PointerToString`

| Item | Details |
|------|---------|
| **Package** | `stringhelper` (`github.com/redhat-best-practices-for-k8s/certsuite/pkg/stringhelper`) |
| **Signature** | `func PointerToString[T any](p *T) string` |
| **Exported?** | Yes – intended for external callers. |

### Purpose
`PointerToString` provides a safe, human‑readable representation of the value pointed to by an arbitrary pointer type.  
It is primarily used in log traces when printing Kubernetes resources that expose fields as pointers (e.g., `*bool`, `*int`).  

- If the incoming pointer is **nil**, it always returns `"nil"`.  
- Otherwise, it converts the dereferenced value to a string using Go’s standard formatting (`fmt.Sprint`).

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `p` | `*T` (generic) | A pointer to any type. The function does **not** modify the underlying value. |

### Return Value
- `string`:  
  - `"nil"` when `p == nil`.  
  - The default string representation of `*p` otherwise (e.g., `"true"`, `"1984"`).

### Key Dependencies
| Dependency | Role |
|------------|------|
| `fmt.Sprint` | Performs the actual conversion from a value to its string form. |

### Side Effects
- **None**: The function only reads the pointer; it never writes or modifies any state.

### Usage Context
In the Certsuite codebase, many Kubernetes resource structs contain optional fields represented as pointers (to differentiate “unset” from zero values). When generating logs or audit trails, these pointers need to be printed in a readable way. `PointerToString` centralises this logic so that all log statements can rely on a single, consistent conversion routine.

### Example Usage
```go
var b *bool
fmt.Println(PointerToString(b)) // → "nil"

bTrue := true
fmt.Println(PointerToString(&bTrue)) // → "true"

num := 1984
fmt.Println(PointerToString(&num)) // → "1984"
```

### Suggested Mermaid Diagram
```mermaid
flowchart TD
    A[Caller] -->|p| B{Is p nil?}
    B -- yes --> C["Return \"nil\""]
    B -- no --> D[fmt.Sprint(*p)]
    D --> E["Return string value"]
```

This function is a small but essential helper that keeps log output clean and avoids repetitive nil‑checks throughout the package.
