WrapLines`

| Attribute | Value |
|-----------|-------|
| **Package** | `internal/cli` |
| **Exported** | ✅ |
| **Signature** | `func WrapLines(s string, max int) []string` |

### Purpose
`WrapLines` is a utility that takes an arbitrary multiline string and breaks it into an array of lines no longer than the supplied *maximum* width (`max`).  
The function preserves word boundaries – words are not split in the middle unless they exceed `max`. Empty lines are preserved. The returned slice can be used for pretty‑printing status messages, logs or help text that must fit within a fixed terminal width.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `s`  | `string` | The source string to wrap; may contain newline characters (`\n`). |
| `max`| `int`    | Maximum allowed length of each line. If a word is longer than `max`, it will be placed on its own line even though the line exceeds `max`. |

### Return value
- `[]string`: A slice where each element is one wrapped line.

### Algorithm (step‑by‑step)
1. **Split input into paragraphs**  
   ```go
   lines := strings.Split(s, "\n")
   ```
   Each paragraph may contain multiple words separated by spaces.

2. **Allocate result slice** – pre‑size it to the number of original lines for efficiency:
   ```go
   wrapped := make([]string, 0, len(lines))
   ```

3. **Process each paragraph**  
   For every `line` in `lines`:
   - If the line is empty (`len(line) == 0`) → append an empty string to `wrapped`.
   - Otherwise split the paragraph into words:
     ```go
     words := strings.Fields(line)
     ```
   - Build a new wrapped line word‑by‑word, ensuring that adding a word does not exceed `max`.  
     *If it would overflow*:  
       - Append the current buffer to `wrapped` (unless empty).  
       - Start a new buffer with the word.  
     *If it fits*: append the word to the buffer (with a space if needed).

4. **Flush any remaining buffer** after finishing a paragraph.

5. Return the populated `wrapped` slice.

### Key Dependencies
- Standard library functions: `strings.Split`, `strings.Fields`, `len`, `append`.
- No external packages or global state are accessed; `WrapLines` is pure and deterministic.

### Side Effects & Thread Safety
- **None** – purely functional.  
- Does not modify its arguments (string values are immutable).  
- Safe for concurrent use.

### Integration in the Package
Within `internal/cli`, this helper is used by various UI functions that need to display wrapped text in a terminal with limited width, such as status banners or log previews. It keeps the rest of the CLI code focused on business logic while delegating line‑wrapping concerns to a single, well‑tested utility.

---

#### Suggested Mermaid Diagram (optional)

```mermaid
flowchart TD
    A[Input String] --> B{Split by "\n"}
    B --> C[Paragraph]
    C --> D{Empty?}
    D -- Yes --> E[Append empty line]
    D -- No --> F{Split into words}
    F --> G[Build wrapped line]
    G --> H{Word fits?}
    H -- Yes --> I[Add to buffer]
    H -- No --> J[Flush buffer, start new]
    J --> G
    I --> G
    G --> K[End paragraph]
    K --> L[Return slice]
```

This visual aid clarifies the control flow for developers new to the package.
