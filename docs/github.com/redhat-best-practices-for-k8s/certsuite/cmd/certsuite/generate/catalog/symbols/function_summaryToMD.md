summaryToMD`

*Location:* `cmd/certsuite/generate/catalog/catalog.go:287`  
*Visibility:* unexported (internal helper)

### Purpose
Converts a **catalog summary** (`catalogSummary`) into a Markdown string that can be printed to the console or written to a file.  
The function is used by the *generate* sub‑command when the user selects the `--output markdown` flag.

### Signature
```go
func summaryToMD(summary catalogSummary) string
```
| Parameter | Type | Description |
|-----------|------|-------------|
| `summary` | `catalogSummary` | A struct that contains a list of classification sections, each with a name and a slice of entries. |

The function returns the entire Markdown representation as a single string.

### Algorithm overview
1. **Header** – Starts with a level‑2 header (`## Summary`) followed by a horizontal rule.
2. **Sections** – For each `catalogSection` in `summary.sections`:  
   * Write the section name as a level‑3 header.  
   * Build a Markdown table:  
     * The first column lists the entry names (e.g., “Certificate”, “Key”).  
     * The second column contains the corresponding value, rendered as a Go format string (`%s` style).  
     * Each row is added to an internal slice of strings and joined with newlines.
3. **Return** – Concatenates all generated lines into one Markdown block.

### Key dependencies
| Dependency | Role |
|------------|------|
| `fmt.Sprintf` | Formats table rows (`"| %s | %s |\n"`). |
| `make`, `append` | Builds slices of strings for the header, section headers, and table rows. |
| `len` | Determines how many sections/rows to process. |
| `strings.Join` | Concatenates slice elements into a single string with newline separators. |

### Side effects
* No global state is modified.
* The function only reads from its argument and standard library helpers.

### Integration in the package
`summaryToMD` is called by the command execution flow when generating catalog output:

```go
if markdownFlag {
    md := summaryToMD(summary)
    fmt.Println(md) // or write to file
}
```

Thus it bridges internal data structures (`catalogSummary`) with user‑facing Markdown documentation.
