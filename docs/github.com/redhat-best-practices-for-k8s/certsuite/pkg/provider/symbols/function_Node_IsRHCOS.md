Node.IsRHCOS` – Detect if a node runs **Red Hat CoreOS**

### Purpose
`IsRHCOS` is a method on the `Node` type that answers the simple question:

> *Does this Kubernetes node run Red Hat CoreOS (RHCOS)?*

The answer influences which tests are executed or skipped.  
For example, certain networking or storage tests may be irrelevant on RHCOS and
are therefore guarded by this check.

### Signature
```go
func (n Node) IsRHCOS() bool
```

* **Receiver** – `Node` is the struct that represents a Kubernetes node in
  *certsuite*.  
* **Return value** – `true` if the node’s OS image contains “rhcos”, otherwise
  `false`.

### How it works

1. **Read the node image label**  
   The method accesses `n.Status.NodeInfo.ContainerRuntimeVersion`.  
   This string typically looks like:
   ```
   containerd://rhel-8.7-rhcos-2023.04.11-00
   ```

2. **Trim whitespace** – `strings.TrimSpace` is called to remove any accidental
   leading/trailing spaces.

3. **Search for “rhcos”** – `strings.Contains` checks whether the trimmed string
   contains the substring `"rhcos"`.  
   The comparison is case‑sensitive and simple; it does not attempt a full
   regex match or version parsing.

4. **Return the result** – If the substring is found, the method returns `true`,
   otherwise `false`.

### Dependencies

| Dependency | Role |
|------------|------|
| `strings.Contains` | Substring search in the image string |
| `strings.TrimSpace` | Clean up whitespace before searching |

No global variables or other package state are touched.

### Side‑effects
None. The method is purely functional: it reads from the node’s data and returns a boolean.

### Package context

* **Location** – `pkg/provider/nodes.go`, line 59.  
* **Related code** – Other helpers such as `IsCentOSStream` or checks for specific labels use similar patterns.  
* **Usage** – Higher‑level test selection logic (e.g., in the provider’s test runner) calls `node.IsRHCOS()` to decide whether to skip or run certain tests.

---

#### Mermaid diagram suggestion

```mermaid
flowchart TD
    A[Node.Status.NodeInfo.ContainerRuntimeVersion] -->|TrimSpace| B[String]
    B -->|Contains("rhcos")| C{Is RHCOS?}
    C -->|Yes| D[Return true]
    C -->|No | E[Return false]
```

This visual helps developers see the simple two‑step decision process performed by `IsRHCOS`.
