scenarioIDToText`

```go
func scenarioIDToText(id string) string
```

### Purpose
`scenarioIDToText` translates an internal *scenario identifier* into a more readable text description that can be shown in generated documentation or logs.  
The function is used by the catalog generation logic to convert opaque IDs (e.g., `"K8S-1234"`) into human‑friendly labels such as `"Container Runtime Security"`.

### Parameters
| Name | Type   | Description |
|------|--------|-------------|
| `id` | `string` | The raw scenario identifier that the catalog contains. |

### Return Value
* A single string containing the human‑readable name for the given scenario ID.  
  If no mapping is found, the function returns the original `id`.

### Key Dependencies & Side Effects
- **No external packages** – the implementation relies only on Go’s standard library.
- **No global state** – the function is pure; it does not modify any package‑level variables (`generateCmd`, `markdownGenerateClassification`, `markdownGenerateCmd`) or other data structures.
- **Deterministic** – given the same input, it always returns the same output.

### How It Fits the Package
The `catalog` subpackage is responsible for generating a structured catalog of certificate tests.  
During generation:

1. The catalog loader reads scenario IDs from configuration files.  
2. When rendering Markdown or other documentation formats, the loader calls `scenarioIDToText` to present each scenario with a readable label instead of its raw ID.

Because the function is unexported, it is only used internally by the package’s generation logic and never exposed to external callers.

---

#### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[Load catalog JSON] --> B{For each scenario}
    B --> C[Get scenarioID]
    C --> D[scenarioIDToText(id)]
    D --> E[Render Markdown with human‑readable name]
```

This diagram illustrates how `scenarioIDToText` sits between the raw data load and the final documentation rendering.
