getResultsSummary`

```go
func getResultsSummary() map[string][]int
```

`getResultsSummary` is an internal helper that aggregates the outcome statistics for all check groups in the database.  
It produces a mapping from group name to a slice of four integers:

| Index | Meaning               |
|-------|-----------------------|
| 0     | Number of **passed** checks (`PASSED`)   |
| 1     | Number of **failed** checks (`FAILED`)   |
| 2     | Number of **skipped** checks (`SKIPPED`) |
| 3     | Number of **aborted** checks (`CheckResultAborted`)* |

(*The `aborted` value is currently unused in the public API but is kept for completeness.)

### Purpose
The function is used by reporting and statistics code to provide a quick, per‑group overview of how many checks succeeded, failed or were skipped. It is invoked after all checks have run and the global `dbByGroup` map has been populated.

### Inputs / Outputs

* **Inputs** – None (reads only package‑level state).  
  It accesses the global `dbByGroup`, which maps group names to `ChecksGroup` objects that contain per‑check results.

* **Output** – A new `map[string][]int`.  
  Each key is a group name, and each value is a slice of four integers as described above. The map is freshly allocated with `make`, so callers can safely modify it without affecting the internal state.

### Key Dependencies

| Dependency | Role |
|------------|------|
| `dbByGroup` | Holds all check groups; the function iterates over its entries to collect statistics. |
| `CheckResult*` constants (`PASSED`, `FAILED`, `SKIPPED`, `CheckResultAborted`) | Used to translate per‑check results into the four integer counts. |
| `make(map[string][]int)` | Allocates the result map. |

No other globals or external packages are used.

### Side Effects

* **None** – The function only reads global state and returns a new data structure; it does not modify any package variables or perform I/O.

### Placement in the Package

`getResultsSummary` lives in `checksdb/checksdb.go`, which contains the core logic for storing, executing, and summarizing checks. It is part of the internal plumbing that powers:

* The command‑line reporting tool (`certsuite run`)  
* API endpoints that expose test results
* Any other component that needs a quick statistical snapshot of all check groups.

Because it is unexported, callers are expected to use higher‑level helpers (e.g., `PrintSummary` or `GenerateReport`) that wrap this function.
