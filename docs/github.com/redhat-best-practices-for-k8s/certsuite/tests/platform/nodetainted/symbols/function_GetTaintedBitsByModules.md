GetTaintedBitsByModules`

**Package**: `nodetainted`  
**File**: `nodetainted.go` (line 183)  
**Exported**: Yes

---

### Purpose
`GetTaintedBitsByModules` transforms a mapping of *module names* to their *taint letter strings* into a bitset representation.  
In this test suite each taint is encoded as a single ASCII character; the function converts those characters into integer bit positions and records which bits are set.

---

### Signature
```go
func GetTaintedBitsByModules(map[string]string) (map[int]bool, error)
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `moduleToLetters` | `map[string]string` | Keys are kernel module names; values are strings of taint letters that the module is known to contain. |

| Return | Type | Description |
|--------|------|-------------|
| `bits` | `map[int]bool` | For each bit position that appears in any letter, the map contains `true`. The key is the bit index (0‑based). |
| `err`  | `error` | Non‑nil if an unknown taint letter is encountered. |

---

### Dependencies

| Dependency | Role |
|------------|------|
| `getBitPosFromLetter(char)` | Converts a single taint letter to its numeric bit position. |
| `fmt.Errorf` | Generates an error when an invalid letter is found. |
| (Unused in this function) `kernelTaints`, `runCommand` | Declared elsewhere in the package; not used here but available for other helpers. |

---

### Algorithm Overview

1. **Create empty result map** – `bits := make(map[int]bool)`.
2. **Iterate over modules** – For each `(module, letters)` pair:
   * Convert the string to a slice of runes.
   * For every rune `c`:
     * Call `getBitPosFromLetter(c)` → `pos`.
     * If `pos < 0`, return an error (`fmt.Errorf("unknown taint letter: %c", c)`).
     * Otherwise, set `bits[pos] = true`.
3. **Return** the populated map and a nil error.

---

### Side Effects
* No global state is mutated.
* Only local variables are created; all data returned is owned by the caller.

---

### How It Fits the Package

`nodetainted` tests whether nodes report taint bits correctly.  
Other helpers in this package:

| Helper | Role |
|--------|------|
| `runCommand` | Executes system commands (e.g., to read `/proc/modules`). |
| `kernelTaints` | Stores known kernel‑taint mappings. |

`GetTaintedBitsByModules` is the core routine that turns human‑readable taint letters into a bitset that can be compared against the node’s reported taint bits during tests.

---

### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[Input: module→taints string] --> B{Iterate modules}
    B --> C{For each letter}
    C --> D[getBitPosFromLetter]
    D --> E[Set bits[pos] = true]
    E --> F[Return map[int]bool]
```

---
