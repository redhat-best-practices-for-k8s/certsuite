getBitPosFromLetter`

| Item | Detail |
|------|--------|
| **Signature** | `func getBitPosFromLetter(letter string) (int, error)` |
| **Purpose** | Translate a single‑character taint letter into the corresponding kernel taint bit index. Kernel taints are encoded as letters in `/proc/sys/kernel/taint`. Each letter maps to a bit position in a 32‑bit mask (base 0). This helper resolves that mapping so other code can manipulate the numeric bit mask directly. |
| **Inputs** | `letter` – the taint identifier supplied by the caller, expected to be one of the characters defined in the package’s internal `kernelTaints` slice. |
| **Outputs** | *`int`* – zero‑based index of the bit that represents the provided letter.<br>*`error`* – returned when the input is empty or not found in `kernelTaints`. The error message includes the offending value for easier debugging. |
| **Key Dependencies** | - `len` (builtin) to validate non‑empty string.<br>- `strings.Contains` (via the imported `"strings"` package) to search the `kernelTaints` slice.<br>- `fmt.Errorf` to create descriptive error values. <br>Global variables used elsewhere in the package (`kernelTaints`, `runCommand`) are **not** accessed by this function, keeping it pure. |
| **Side‑Effects** | None – purely functional; does not modify global state or interact with the system. |
| **Package Context** | In `nodetainted` the test harness parses the kernel taint string (e.g., `"D"`, `"M"`) and needs to set/unset bits in a 32‑bit integer. This helper bridges the human‑readable letter representation and the bit mask used by the test logic, enabling other functions such as `applyTaintMask` or `removeTaintMask` to work with numeric indices. |
| **Example Usage** | ```go\nidx, err := getBitPosFromLetter(\"D\")\nif err != nil { /* handle */ }\nmask := 1 << idx // bit mask for the \"D\" taint\n``` |

### Suggested Mermaid Flow (optional)

```mermaid
flowchart TD
    A[Input letter] --> B{Is non‑empty?}
    B -- No --> C[Return error]
    B -- Yes --> D{Letter in kernelTaints?}
    D -- No --> E[Return error]
    D -- Yes --> F[Find index of letter]
    F --> G[Return (index, nil)]
```

This diagram visualises the decision path: validation → lookup → result.
