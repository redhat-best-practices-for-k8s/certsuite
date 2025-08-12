DecodeKernelTaintsFromBitMask`

```go
func DecodeKernelTaintsFromBitMask(bitmask uint64) []string
```

### Purpose
Transforms a 64‑bit integer that represents kernel taint flags into the corresponding human‑readable taint messages.

In this package the kernel exposes its current taints as a bit mask (`uint64`).  
Each bit position has a known meaning (e.g. `0x0001` = “Kernel taint: bad memory”).  
The function iterates over the predefined mapping (`kernelTaints`) and builds a slice of
the messages for all bits that are set in *bitmask*.

### Parameters
| Name   | Type   | Description |
|--------|--------|-------------|
| `bitmask` | `uint64` | The kernel taint bit mask returned by the system. |

### Return value
| Value | Type      | Description |
|-------|-----------|-------------|
| `[]string` | Slice of strings | Each element is a human‑readable description of a set taint. If no bits are set an empty slice is returned. |

### Key dependencies
* **`kernelTaints`** – A package‑level array that pairs each bit position with its symbolic name.
  The function loops over this array to determine which bits to test.
* **`GetTaintMsg(bit)`** – Utility that maps a bit value (e.g., `0x0001`) to the corresponding message string.  
  This is called for every set bit found in *bitmask*.

### Side effects
None. The function only reads global data and constructs a new slice; it does not modify any state or call external commands.

### How it fits the package
`nodetainted` tests whether a node’s kernel taints are present when certain operations (e.g., running a pod) should be blocked.  
Other helpers in the package gather the raw bit mask via `runCommand`, then use this function to translate that mask into a list of taint messages that can be asserted against expected values.

---

#### Suggested Mermaid diagram

```mermaid
flowchart TD
    A[Input: bitmask] --> B{For each kernelTaint}
    B --> C{Bit set?}
    C -- Yes --> D[GetTaintMsg(bit)]
    D --> E[Append to slice]
    C -- No --> F[Skip]
    E --> G[Return slice]
```
