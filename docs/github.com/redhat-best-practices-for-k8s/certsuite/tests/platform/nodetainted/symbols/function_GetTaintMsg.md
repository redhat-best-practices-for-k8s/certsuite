GetTaintMsg`

| Item | Detail |
|------|--------|
| **Package** | `nodetainted` – part of CertSuite’s platform tests for node taint handling. |
| **Signature** | `func GetTaintMsg(t int) string` |
| **Exported** | Yes |

### Purpose
`GetTaintMsg` translates a numeric *taint identifier* into a human‑readable description that explains why a Kubernetes node is considered tainted in the context of CertSuite’s tests.  
The function is used by test helpers to generate clear error messages when a node fails taint‑related assertions.

### Parameters
- `t int` – The index of the taint type. It corresponds to an entry in the unexported slice `kernelTaints`.

### Return Value
- `string` – A formatted message describing the taint, e.g.:

  ```text
  "Node has taint kernel-version-mismatch: required kubelet version %s is different from node's kernel version %s"
  ```

The exact wording depends on the index provided.

### Dependencies & Side Effects
| Dependency | How it’s used |
|------------|---------------|
| `kernelTaints` (unexported slice) | Holds pre‑defined taint messages; `GetTaintMsg` indexes into this slice. |
| `fmt.Sprintf` | Used twice to interpolate the message with dynamic values (`%s`). |

No global state is modified, and the function has no external side effects beyond string construction.

### Usage Flow
```go
taintIndex := 2 // chosen by test logic
msg := nodetainted.GetTaintMsg(taintIndex)
// msg now contains a descriptive taint message used in assertions or logs
```

### Diagram (optional)

```mermaid
flowchart TD
    A[GetTaintMsg(t)] --> B{Index into kernelTaints}
    B --> C[Retrieve template string]
    C --> D[Sprintf with %s placeholders]
    D --> E[Return formatted message]
```

---

**Fit in the package:**  
`nodetainted` provides utilities for checking node taints during CertSuite platform tests. `GetTaintMsg` is a small helper that centralises taint‑message formatting, keeping test code concise and ensuring consistent wording across different checks.
