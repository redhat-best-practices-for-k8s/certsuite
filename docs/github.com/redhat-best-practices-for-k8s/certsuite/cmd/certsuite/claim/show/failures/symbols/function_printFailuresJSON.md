printFailuresJSON`

### Purpose
`printFailuresJSON` is a helper that serialises the list of failed test suites into pretty‑printed JSON and writes it to standard output.  
It is invoked by the *show failures* subcommand when the user selects the `json` output format.

### Signature
```go
func printFailuresJSON([]FailedTestSuite)
```

| Parameter | Type                | Description                         |
|-----------|---------------------|-------------------------------------|
| `failedSuites` | `[]FailedTestSuite` | A slice of failed test suite objects that the command has collected from a claim file. |

The function does **not** return any value; it terminates the program via `log.Fatalf` if JSON marshaling fails.

### Workflow
1. **Marshal**  
   Uses `json.MarshalIndent` to convert the `failedSuites` slice into an indented JSON string (`indent = "  "`).  

2. **Handle error**  
   If marshaling fails, it logs a fatal error message that includes the format name `"JSON"` and exits.

3. **Output**  
   On success, prints the resulting JSON to stdout with `log.Printf("%s\n", string(b))`.

### Dependencies
- `encoding/json.MarshalIndent` – serialises Go structs into formatted JSON.
- `log.Fatalf` – logs an error message and aborts execution if marshaling fails.
- `log.Printf` – writes the final JSON string to standard output.

### Side effects
- **Writes** to standard output (stdout).  
- **Exits** the program on marshal failure (`Fatalf`).  
- Does not modify its input slice or any global state.

### Integration in the package
The `failures` package implements the `show failures` command for Certsuite. The command parses flags such as `--output-format`, and when that flag equals `"json"` it calls `printFailuresJSON`.  
Other output formats (`text`, `invalid`) are handled by separate helper functions, but they share the same signature pattern.

---

#### Suggested Mermaid diagram (optional)

```mermaid
flowchart TD
    subgraph Command
        A[Parse flags] --> B{output-format}
        B -- "json" --> C[printFailuresJSON]
        B -- "text" --> D[printFailuresText]
    end

    subgraph printFailuresJSON
        C1[MarshalIndent(failedSuites)] --> C2{err?}
        C2 -- yes --> E[log.Fatalf]
        C2 -- no --> F[log.Printf(json)]
    end
```

This diagram shows the decision point on output format and the internal steps of `printFailuresJSON`.
