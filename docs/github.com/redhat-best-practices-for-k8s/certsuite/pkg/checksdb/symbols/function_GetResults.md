GetResults`

```go
func GetResults() map[string]claim.Result
```

### Purpose
`GetResults` aggregates the latest test results that have been stored in the **checks database** and returns them as a simple, thread‑safe lookup table.

The function is used by other packages (e.g. reporting or metrics) to obtain a snapshot of all checks that were executed during the current run, without exposing the internal structures (`ChecksGroup`, `Check`) that hold the data.

### Inputs / Outputs
| Direction | Type                         | Description |
|-----------|------------------------------|-------------|
| **Input** | *none*                       | The function reads from package‑level globals; no caller‑supplied arguments. |
| **Output** | `map[string]claim.Result`   | A map keyed by the fully‑qualified check name (`<group>.<check>`). Each value is a `claim.Result`, the type that represents a single test outcome (passed, failed, skipped, aborted or error). |

### Key Dependencies
| Dependency | Role |
|------------|------|
| `dbByGroup` | Holds all `ChecksGroup` instances keyed by group name.  Each `ChecksGroup` contains a slice of `Check` objects, each of which stores its own `Result`. |
| `dbLock`    | A `sync.Mutex` protecting concurrent access to the database structures.  `GetResults` locks this mutex while it reads from `dbByGroup`. |

### Side‑Effects
* **No side‑effects** – the function only reads state; it does not modify any global variable or lock.
* The returned map is a *shallow copy*: keys are strings and values are copies of the underlying `claim.Result` objects.  Mutating the returned map will not affect the internal database.

### How It Fits the Package
The checks package maintains a mutable in‑memory store (`dbByGroup`) that receives results from individual test executions.  
Other parts of the application (e.g., reporting, CLI output) need a stable view of those results.  `GetResults` provides this view:

```go
results := checksdb.GetResults()
for name, r := range results {
    fmt.Printf("%s: %s\n", name, r.Status())
}
```

Because the function is public, it becomes the canonical way for consumers to read test outcomes without exposing or depending on the internal `ChecksGroup`/`Check` structures.

### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[Caller] -->|calls| B(GetResults)
    B --> C{Lock dbLock}
    C --> D[Read dbByGroup]
    D --> E[Build map[string]claim.Result]
    E --> F[Return map]
```

This diagram shows the read‑only path through the database, highlighting that the function holds a lock while it copies results.
