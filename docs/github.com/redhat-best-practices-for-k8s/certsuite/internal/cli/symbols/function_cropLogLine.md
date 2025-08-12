cropLogLine`

```go
func cropLogLine(line string, maxLen int) string
```

#### Purpose
`cropLogLine` is a small helper used by the CLI logger to keep each line of output short enough for terminal display.  
When logs are streamed in real‑time, they may contain carriage‑return (`\r`) characters that cause lines to overwrite one another.  The function removes those control characters and optionally truncates the string if it exceeds a configured maximum width.

#### Parameters
| Name   | Type   | Description |
|--------|--------|-------------|
| `line` | `string` | A raw log line received from the check process. |
| `maxLen` | `int` | The desired maximum length for the output string.  If `len(line) <= maxLen`, the original string is returned unchanged. |

#### Return Value
A single string that has:
1. All carriage‑return characters removed (`\r` → ``).
2. Truncated to `maxLen` characters if it was longer.

The function never modifies global state or writes to I/O; it simply returns a processed string.

#### Key Dependencies
* **Standard Library** – uses `strings.ReplaceAll` to strip `\r` and `len` to measure the string length.
* No other package variables are accessed, making the function pure and easy to test in isolation.

#### Side‑Effects
None.  The function is deterministic and has no observable effects beyond its return value.

#### Usage Context
Within the `cli` package this helper is called by the log‑sniffing goroutine that reads from `CliCheckLogSniffer`.  
When a line is received, it is first passed through `cropLogLine` before being sent to the user’s terminal.  This keeps output clean and prevents overly long lines from scrolling past or wrapping unexpectedly.

```go
for {
    select {
    case raw := <-CliCheckLogSniffer:
        cleaned := cropLogLine(raw, lineLength)
        fmt.Println(cleaned)   // printed to stdout
    ...
```

By centralising the cropping logic in a dedicated function, the rest of the CLI code can remain agnostic about terminal width handling and focus on higher‑level formatting.
