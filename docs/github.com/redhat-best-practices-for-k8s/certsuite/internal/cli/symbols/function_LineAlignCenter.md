LineAlignCenter`

| Aspect | Details |
|--------|---------|
| **Package** | `internal/cli` |
| **Signature** | `func LineAlignCenter(s string, width int) string` |
| **Exported?** | ✅ Yes – used by other parts of the CLI to format status messages. |

### Purpose
`LineAlignCenter` takes a text line and a desired output width, then returns a new string padded with spaces so that the original text is centered within the specified width.  
It is primarily used when printing banners or status bars where the text must be visually balanced.

### Parameters
| Name | Type   | Meaning |
|------|--------|---------|
| `s`  | `string` | The content to center. |
| `width` | `int` | Target column width (including the padded spaces). |

> **Note**: If `len(s)` is greater than `width`, the function returns `s` unchanged.

### Return Value
A new string of length `width`.  
If centering requires an odd number of padding spaces, the extra space goes to the right side to keep the text visually centered.

### Implementation Details
```go
func LineAlignCenter(s string, width int) string {
    if len(s) >= width { return s }

    // Compute left/right padding.
    totalPad := width - len(s)
    leftPad  := totalPad / 2
    rightPad := totalPad - leftPad

    // Build padded line: "%<leftPad>s%s%<rightPad>s"
    return fmt.Sprintf("%*s%s%*s", leftPad, "", s, rightPad, "")
}
```
- Uses `fmt.Sprintf` twice to generate the left and right spaces.
- Relies on Go’s built‑in `len()` for byte length (works fine for ASCII banners used in this project).

### Dependencies
- **Standard library**: `fmt`, `strings`.  
  No external packages or global state are accessed.

### Side Effects
None – pure function. It performs no I/O, logging, or modification of globals.

### Role in the Package
`LineAlignCenter` is a small helper that keeps the CLI’s visual output tidy.  
It is called from functions that print banners (e.g., `PrintBanner`) and progress indicators where text must be horizontally centered within a fixed terminal width (`lineLength` global). Because it is pure, it can be unit‑tested in isolation.

---

**Mermaid diagram suggestion**

```mermaid
flowchart TD
    A[Input string s] --> B{len(s) < width?}
    B -- No --> C[Return s]
    B -- Yes --> D[Compute totalPad = width - len(s)]
    D --> E[leftPad = totalPad/2, rightPad = totalPad-leftPad]
    E --> F[fmt.Sprintf("%*s%s%*s", leftPad, "", s, rightPad, "")]
    F --> G[Return padded string]
```

This diagram visualises the decision path and padding logic used by `LineAlignCenter`.
