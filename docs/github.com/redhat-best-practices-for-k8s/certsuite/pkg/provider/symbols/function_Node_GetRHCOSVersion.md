Node.GetRHCOSVersion` – Overview

| Aspect | Detail |
|--------|--------|
| **Purpose** | Retrieve the short RHEL‑CoreOS (RHCOS) version running on a node, if the node is indeed an RHCOS instance. |
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/pkg/provider` |
| **Exported?** | Yes – callers can invoke it directly on a `Node`. |

---

### Function Signature

```go
func (n Node) GetRHCOSVersion() (string, error)
```

* **Receiver**: `Node` – the struct representing an OpenShift node.  
  The method operates purely on this instance; no external state is modified.

* **Returns**:
  * `string`: a *short* RHCOS version string (`<major>.<minor>`), e.g. `"4.15"`.
  * `error`: non‑nil when the node isn’t an RHCOS node or the version cannot be parsed.

---

### How It Works

1. **Check Node Type**  
   Calls `n.IsRHCOS()` (see `Node.IsRHCOS`) to determine if this node runs RHCOS.  
   *If false* → return an error: `"not a rhcos node"`.

2. **Parse Version from `node.Status.NodeInfo.OSImage`**  
   The OS image string typically looks like:
   ```
   Red Hat Enterprise Linux CoreOS 4.15.0-0.nightly-2023-01-01
   ```
   * Split the string twice on `" "` to isolate the version part (`"4.15.0-0.nightly-2023-01-01"`).  
   * Trim any surrounding whitespace.

3. **Convert Long → Short Version**  
   Pass the extracted long form to `GetShortVersionFromLong`, which removes the patch, release‑candidate, and build components, returning just `<major>.<minor>` (e.g., `"4.15"`).

4. **Return Result** – if all steps succeed, return the short version; otherwise propagate any parsing error.

---

### Dependencies

| Dependency | Role |
|------------|------|
| `IsRHCOS` | Determines node type. |
| `Errorf`  | Constructs a formatted error when not an RHCOS node. |
| `strings.Split`, `TrimSpace` | Basic string manipulation to isolate the version part. |
| `GetShortVersionFromLong` | Normalises the long OS image string into a short semantic version. |

These are all local functions or standard library calls; no external packages are imported.

---

### Side‑Effects

* **None** – The method is read‑only: it reads fields from the receiver and performs pure calculations, never mutating any state.

---

### Context within `provider` Package

`Node.GetRHCOSVersion` is part of a suite of node introspection helpers.  
Other methods in this file (`IsRHCOS`, `GetLongVersionFromImage`, etc.) provide complementary information such as:

* Whether the node runs RHCOS, RHCE, or RHEL.
* The full OS image string.
* CPU and memory characteristics.

`GetRHCOSVersion` is typically used by tests that need to assert compliance rules specific to a particular RHCOS release (e.g., verifying security patch levels). It feeds into higher‑level logic in the provider that maps nodes to their OpenShift version and applies relevant test suites.
