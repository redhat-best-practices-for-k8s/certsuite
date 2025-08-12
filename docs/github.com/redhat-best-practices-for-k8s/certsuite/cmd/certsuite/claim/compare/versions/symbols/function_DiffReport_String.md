DiffReport.String() string`

### Purpose
`String` is a receiver method on the `DiffReport` type that serialises the report into a human‑readable string.  
It is used by the command‑line tool to display differences between two claim files (or directories) in a concise, tabular format.

### Signature
```go
func (r DiffReport) String() string
```

- **Receiver** – `DiffReport` holds the results of a comparison: lists of added/removed/modified fields and any error messages.
- **Returns** – A single string containing the formatted report ready for printing or logging.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `String()` (called twice inside the method) | The first call produces the header row, the second renders each line of the report. These are likely helper functions defined in the same package that format rows based on the data in `DiffReport`. |

No external packages or globals are referenced; all work is confined to local helpers and the `DiffReport` value itself.

### Side Effects
- None. The method only reads from the receiver and constructs a string; it does not modify any state or write to files/IO streams.

### How It Fits in the Package
The **versions** package implements comparison logic for claim files, producing a `DiffReport`.  
`String()` is the canonical way to convert that report into text so that:

1. The command‑line tool (`certsuite`) can print it to stdout.
2. Tests or other consumers can capture the output as a string.

It is the public interface for inspecting the results of a claim comparison.
