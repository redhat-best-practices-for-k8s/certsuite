generateZipFileName` (unexported)

### Purpose
Creates a deterministic file name for the zip archive that contains all test results.  
The function is used by the archiving logic in `internal/results/archiver.go` to
store a snapshot of the current run under a timestamp‑based name.

### Signature
```go
func generateZipFileName() string
```

- **No parameters** – the function derives everything from the current time.
- **Returns**: a string that represents the file path (including directory and
  extension) for the zip archive.

### Implementation details

| Step | Code snippet | Explanation |
|------|--------------|-------------|
| 1 | `time.Now()` | Fetches the current wall‑clock time. |
| 2 | `t.Format("2006-01-02T15-04-05")` | Formats the timestamp in a sortable, human‑readable form (ISO‑8601 without colons). |
| 3 | `filepath.Join("results", fmt.Sprintf("%s.zip", formattedTime))` | Builds the final path: `<outputDir>/YYYY-MM-DDTHH-MM-SS.zip`. The prefix `"results"` comes from the package’s default output directory. |

The function does **not** write any files; it merely returns a string.

### Dependencies

| Dependency | Role |
|------------|------|
| `time.Now` | Current time for timestamp. |
| `time.Time.Format` | Formats the timestamp. |
| `fmt.Sprintf` | Builds the file name suffix (`".zip"`). |
| `filepath.Join` | Normalises path separators across operating systems. |

No external packages or global variables are accessed.

### Side‑effects
None – pure function. The returned string is used by callers to open a new zip writer.

### Package context

Within the `results` package, archiving happens in two steps:

1. **Generate a timestamped name** (`generateZipFileName`)  
2. **Write results into that archive** (other functions read embedded HTML/JS templates and write test output).

Thus `generateZipFileName` is the entry point for naming the result artifacts, ensuring each run produces a uniquely identifiable zip file without overwriting previous data.

### Suggested Mermaid diagram

```mermaid
flowchart TD
    A[Current Run] --> B{Generate Zip Name}
    B --> C{{Time.Now()}}
    C --> D[Format Timestamp]
    D --> E[Build Path: results/<timestamp>.zip]
    E --> F[Return String]
```

This diagram visualises the single‑path flow from a test run to the resulting file name.
