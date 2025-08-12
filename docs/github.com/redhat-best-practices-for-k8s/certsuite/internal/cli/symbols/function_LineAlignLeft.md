LineAlignLeft`

```go
func LineAlignLeft(s string, width int) string
```

### Purpose  
`LineAlignLeft` formats a text line so that it is left‑aligned within a fixed column width.  
It is used by the CLI renderer to keep status messages and progress bars neatly aligned in the terminal.

### Parameters  

| Name   | Type  | Description |
|--------|-------|-------------|
| `s`    | `string` | The raw text that should be displayed. |
| `width` | `int` | Desired column width (in runes). If the string is shorter than this value, it will be padded with spaces; if longer, it will be returned unchanged. |

### Return Value  

- A new `string` whose visual length is at least `width`.  
  *If `len(s) < width`*, the result consists of `s` followed by enough spaces to reach `width`.  
  *Otherwise* the original string is returned.

### Key Dependency  

The function uses the standard library’s `fmt.Sprintf` to build the padded string:

```go
return fmt.Sprintf("%-*s", width, s)
```

`%-*s` formats a left‑aligned string with a dynamic field width (`width`). No other packages or global variables are referenced.

### Side Effects  

None. The function is pure: it only returns a new string and does not modify any external state.

### Package Context  

`LineAlignLeft` lives in the `cli` package, which provides command‑line rendering utilities for CertSuite.  
It is typically invoked by higher‑level formatting helpers that compose status bars, progress indicators, or log messages so that they line up vertically regardless of varying content lengths. The function plays a small but essential role in keeping the CLI output readable and aesthetically consistent.
