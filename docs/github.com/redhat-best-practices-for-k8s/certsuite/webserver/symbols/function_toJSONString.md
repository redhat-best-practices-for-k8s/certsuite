toJSONString`

| Item | Detail |
|------|--------|
| **Package** | `webserver` (`github.com/redhat-best-practices-for-k8s/certsuite/webserver`) |
| **Signature** | `func toJSONString(map[string]string) string` |
| **Visibility** | Unexported (used only inside this package) |

### Purpose
`toJSONString` serialises a map of string keys and values into a human‑readable JSON string.  
It is used when the web server needs to send configuration data or status information to a browser client in JSON format.

### Parameters
- `m map[string]string`: A key/value mapping that will be encoded as JSON. The map’s contents are expected to be simple strings; any non‑string values would cause an encoding error (which is ignored by the helper).

### Return Value
- `string`: The UTF‑8 JSON representation of `m`. If marshaling fails, it returns an empty string.

### Implementation Details
```go
func toJSONString(m map[string]string) string {
    b, err := json.MarshalIndent(m, "", "  ")
    if err != nil {
        return ""
    }
    return string(b)
}
```
* The function uses `json.MarshalIndent` from the standard library to produce pretty‑printed JSON with two‑space indentation.  
* Errors are swallowed; callers must handle an empty string as “encoding failed”.

### Dependencies
| Dependency | Role |
|------------|------|
| `encoding/json.MarshalIndent` | Converts the map into indented JSON bytes. |
| `string()` conversion | Turns the byte slice into a Go string for return. |

### Side Effects
None – it only reads from its argument and returns a new value; no global state is modified.

### Context in Package
- **Other helpers**: `toJSONString` is part of a small set of utilities that convert internal data structures into forms suitable for HTTP responses (e.g., generating JSON for AJAX calls).  
- **Usage**: It is called by handlers that respond with configuration data, such as the `/config` endpoint or WebSocket message handlers.  

### Example
```go
// Suppose we have a map of environment variables to expose:
env := map[string]string{
    "API_URL":   "https://api.example.com",
    "LOG_LEVEL": "debug",
}
jsonStr := toJSONString(env)
// jsonStr now contains:
// {
//   "API_URL":   "https://api.example.com",
//   "LOG_LEVEL": "debug"
// }
```

---

#### Mermaid Diagram (Optional)

```mermaid
flowchart TD
    A[Handler] --> B{Prepare map}
    B --> C[toJSONString(map)]
    C --> D[Return JSON string]
```
This diagram illustrates how a handler builds a `map[string]string`, passes it to `toJSONString`, and obtains the JSON string that is sent back to the client.
