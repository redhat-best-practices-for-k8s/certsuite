DecodeKernelTaintsFromLetters`

| Item | Detail |
|------|--------|
| **Package** | `nodetainted` (github.com/redhat-best-practices-for-k8s/certsuite/tests/platform/nodetainted) |
| **Signature** | `func DecodeKernelTaintsFromLetters(letters string) []string` |
| **Exported?** | Yes |

---

### Purpose
Converts a compact, letter‑based representation of kernel taints into the canonical *key=value:effect* strings used by Kubernetes.

The function is invoked after retrieving the short taint code from a node’s `nodetainted` test output. The input string may contain:

- Single letters that map to predefined taint keys (e.g., `"M"` → `node.kubernetes.io/memory-pressure`)  
- Optional suffixes (`"!"`, `"?"`, `"^"`) indicating the taint effect:  
  - `!` → NoSchedule  
  - `?` → PreferNoSchedule  
  - `^` → NoExecute  

The function expands each letter into its full representation, preserving order.

---

### Inputs
| Parameter | Type   | Description |
|-----------|--------|-------------|
| `letters` | `string` | A concatenated string of taint letters and optional effect markers. Example: `"M!E?"`. |

---

### Outputs
| Return Value | Type      | Description |
|--------------|-----------|-------------|
| `[]string`   | Slice of strings | Each element is a full taint string (`key=value:effect`). Order matches the input sequence. If an unrecognized letter appears, it is ignored. |

---

### Key Dependencies

| Dependency | Role |
|------------|------|
| `kernelTaints` (unexported global) | Maps single‑character letters to their corresponding key/value strings. |
| `runCommand` (unused in this function but part of the same file) | Not used directly; present for context in the package. |

---

### Algorithm Overview

1. **Iterate over `letters`.**  
   The function walks each rune in the input string.
2. **Identify taint keys.**  
   For any letter that exists as a key in `kernelTaints`, create a base taint string using its value (`fmt.Sprintf("%s=%s", key, val)`).
3. **Handle effect markers.**  
   If the next rune is one of `!`, `?`, or `^`, append the corresponding effect to the base taint (`:NoSchedule`, `:PreferNoSchedule`, or `:NoExecute`).  
   The marker consumes its own iteration; the function continues with subsequent letters.
4. **Collect results.**  
   Valid taints are appended to a slice that is returned at the end.

If an unrecognized letter or effect marker appears, it is silently skipped (no error).

---

### Side‑Effects
* None. The function is pure: no global state is mutated and no external calls are made beyond string formatting.

---

### Usage Context

Within the `nodetainted` test suite, nodes may advertise kernel taints using a compact notation in their annotations or status fields.  
`DecodeKernelTaintsFromLetters` translates that notation into the form expected by Kubernetes API objects (`Node.Spec.Taints`).  
The resulting slice is then compared against the actual node taints to verify that the system behaves as configured.

---

### Example

```go
// Input: "M!E?"  -> Memory pressure NoSchedule, Egress NoExecute
taints := DecodeKernelTaintsFromLetters("M!E?")
fmt.Println(taints)
// Output:
//   [node.kubernetes.io/memory-pressure=NoSchedule node.egress=NoExecute]
```

---

### Mermaid Flow (Optional)

```mermaid
flowchart TD
  A[Input string] --> B{Iterate each rune}
  B -->|Letter in kernelTaints| C[Create base taint]
  B -->|Effect marker (!, ?, ^)| D[Append effect to last taint]
  B -->|Unrecognized| E[Skip]
  C --> F[Add to result slice]
  D --> F
  F --> G{Next rune}
  G --> B
  F --> H[Return slice]
```

---

This function is a small, well‑defined helper that bridges the test suite’s compact taint encoding with Kubernetes’ expected taint format.
