FsDiff.GetResults`

| Aspect | Detail |
|--------|--------|
| **Purpose** | Aggregates the outcome of a filesystem‑diff test run and produces a single integer exit code that can be used by callers (e.g., CI pipelines) to determine success or failure. |
| **Receiver** | `FsDiff` – the struct that holds state for a diff operation (created in `NewFsDiff`, populated during `Run`). The method uses only the receiver’s fields, so it is pure with respect to global state. |
| **Signature** | `func (f FsDiff) GetResults() int` |
| **Inputs** | No explicit arguments; the method relies on data already stored in the `FsDiff` instance: e.g., the list of diffs collected, any errors encountered during mounting or comparison, and any logs that were written to temporary directories. |
| **Outputs** | An integer exit code:<br>• `0` – all comparisons succeeded (no differences found).<br>• Non‑zero – at least one difference or error was detected. The exact non‑zero value is not specified in the snippet, but typical practice is to return a distinct code for each failure type (e.g., 1 for file mismatch, 2 for mount error). |
| **Key Dependencies** | *None* – the method does not call external functions or packages beyond what is already imported by the package. It only inspects fields of `FsDiff`. |
| **Side Effects** | None; the method is read‑only with respect to both the receiver and global variables. It may, however, close any open temporary directories or files if such cleanup logic is embedded (not shown in the snippet). |
| **Integration into the Package** | `cnffsdiff` provides a high‑level API for comparing filesystem snapshots between two containers or node images. The typical flow is: <br>1. Construct an `FsDiff` instance.<br>2. Call its `Run()` method to perform mounts, diff operations, and log results.<br>3. After completion, call `GetResults()` to obtain a consolidated exit status that can be asserted in tests or passed back to the test harness. |

### Suggested Mermaid Diagram

```mermaid
flowchart TD
    A[FsDiff instance] -->|Run()| B[Perform diff]
    B --> C{Differences?}
    C -- yes --> D[Set error flag]
    C -- no --> E[All good]
    E --> F[GetResults()]
    D --> F
    F --> G[Return exit code]
```

This diagram illustrates how `GetResults()` consumes the state produced by `Run()` and converts it into a single integer outcome.
