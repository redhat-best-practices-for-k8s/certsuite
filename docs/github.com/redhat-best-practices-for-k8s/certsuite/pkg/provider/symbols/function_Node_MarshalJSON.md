Node.MarshalJSON`

### Purpose
`MarshalJSON` implements the `json.Marshaler` interface for the `Node` type.  
It serialises a `Node` instance into its JSON representation so that it can be
encoded to or written in any context that expects JSON (e.g., API responses,
log output, file persistence).

### Receiver & Signature
```go
func (n Node) MarshalJSON() ([]byte, error)
```
- **Receiver** – `Node` (passed by value).  
  The method operates on a copy of the node, ensuring no mutation of the caller’s instance.
- **Return values** –  
  - `[]byte`: the JSON‑encoded byte slice.  
  - `error`: non‑nil if the underlying `encoding/json.Marshal` fails.

### Implementation Details
```go
func (n Node) MarshalJSON() ([]byte, error) {
    return json.Marshal(n)
}
```
- Delegates to the standard library’s `json.Marshal`.  
- No additional fields are added or omitted; it simply forwards the struct as‑is.
- Because `Node` is a value receiver, concurrent callers cannot interfere with each other.

### Dependencies
| Dependency | Role |
|------------|------|
| `encoding/json.Marshal` | Performs the actual conversion to JSON. |

No global variables, constants, or external packages are accessed by this method.

### Side‑effects
- **None** – the method is read‑only; it does not modify the node or any other state.

### Context within the Package
* `provider/nodes.go` defines the `Node` struct (not shown here) and includes this marshaller to satisfy interfaces that require JSON output.  
* The package `github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider` is responsible for representing Kubernetes resources; `MarshalJSON` allows a node’s data to be serialized when:
  * Sending diagnostic information over the network.
  * Writing test results to disk.
  * Logging detailed state in debug or audit trails.

---

#### Quick Usage
```go
node := Node{ /* fields */ }
data, err := json.Marshal(node) // internally calls node.MarshalJSON()
if err != nil { log.Fatal(err) }

fmt.Println(string(data)) // JSON representation of the node
```

This method is intentionally lightweight and relies on Go’s standard JSON encoder to maintain compatibility with the rest of the ecosystem.
