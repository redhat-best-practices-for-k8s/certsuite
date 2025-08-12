checkResults` – Command handler for the *results* sub‑command

```go
func checkResults(cmd *cobra.Command, args []string) error
```

| Item | Details |
|------|---------|
| **Purpose** | Validate that the current test run produced the expected results.  It is invoked by the `certsuite check results` command and performs a diff between the real outcomes (stored in a SQLite DB) and the expected outcomes defined in the policy bundle. |
| **Inputs** | * `cmd`: the Cobra command instance that called this function.  It is used to read flags such as the path to the test‑results database, the name of the policy bundle, and whether to write a template file. <br>* `args`: positional arguments – unused in the current implementation (always empty). |
| **Outputs** | * Returns an `error` only when something goes wrong during execution (e.g., DB access failure or template generation error).  A non‑zero exit status is issued through `cmd.Exit(...)` if mismatches are found. |

### Flow of the function

```mermaid
flowchart TD
    A[Start] --> B{Get flags}
    B --> C{DB path?}
    C --> D[getTestResultsDB]
    D --> E{Error?} -->|yes| F[Return err]
    E --> G{Read expected results}
    G --> H{Mismatches?}
    H -->|none| I[Print “All tests passed”]
    H -->|some| J[print mismatch details]
    J --> K[Exit(1)]
```

1. **Retrieve command flags**  
   * `--db` → database file that contains the test results.  
   * `--bundle` → path to the policy bundle from which expected results are extracted.  
   * `--template` (bool) → whether to generate a template of the results file.

2. **Open the results database** via `getTestResultsDB`.  If opening fails, return the error.

3. **Generate or load the expected test‑results JSON** by calling `getExpectedTestResults`.  
   * When `--template` is set, a new template file named `<bundle>_test_results.json` is created with the appropriate permissions (`0644`) using `generateTemplateFile`.

4. **Compare**:  
   * The function iterates over the rows returned by the DB, appending any mismatch between the real and expected results to a slice.
   * If no mismatches are found it prints “All tests passed” and exits successfully.

5. **Mismatch handling** – When mismatches exist:
   * `printTestResultsMismatch` prints detailed differences (which test failed or was missing).
   * The command terminates with `cmd.Exit(1)` to signal failure to the caller/CI system.

### Dependencies

| Called function | Responsibility |
|-----------------|----------------|
| `getTestResultsDB` | Opens SQLite DB, returns `*sql.DB`. |
| `Errorf`, `Println` | Logging helpers from Cobra. |
| `generateTemplateFile` | Creates a JSON template file for expected results. |
| `getExpectedTestResults` | Reads expected outcomes from the bundle. |
| `printTestResultsMismatch` | Formats and prints mismatched test entries. |

### Side‑effects

* Writes to standard output (success or mismatch details).  
* May create a `<bundle>_test_results.json` file if `--template` is used.  
* Calls `cmd.Exit(1)` on failure, causing the program to terminate with exit status 1.

### Package context

The `results` package lives under `cmd/certsuite/check/results`.  
It implements the **check** command’s *results* sub‑command and relies on the shared test‑result database format defined elsewhere in the repo.  The function is unexported because it is only invoked internally by Cobra when the user runs:

```bash
certsuite check results --db <path> --bundle <policy> [--template]
```

This command is essential for CI pipelines to assert that a policy bundle’s tests pass against the current Kubernetes cluster.
