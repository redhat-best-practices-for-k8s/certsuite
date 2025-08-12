showInfo` – Display Test Information

### Purpose
`showInfo` is the core handler for the `info` sub‑command of **certsuite**.  
It gathers user‑requested options, looks up matching test IDs, and prints either:

* a list of tests (short description) when the `--list` flag is set, or  
* detailed case information boxes for each matched test.

### Signature
```go
func showInfo(cmd *cobra.Command, args []string) error
```
* `cmd`: the cobra command instance that invoked this function.  
  Used to read flags and provide context for error messages.  
* `args`: positional arguments supplied by the user (unused in current logic).  

The function returns an `error` so Cobra can surface problems back to the CLI.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `GetString`, `GetBool` | Read command‑line flags (`--file`, `--list`). |
| `Flags()` | Access the flag set for this command. |
| `getMatchingTestIDs` | Resolve a user‑supplied pattern (or file) into concrete test IDs. |
| `printTestList` | Render a plain list of tests when `--list` is used. |
| `getTestDescriptionsFromTestIDs` | Retrieve full test descriptions needed for the detailed view. |
| `adjustLineMaxWidth`, `printTestCaseInfoBox` | Format and display each test case in a nicely wrapped box. |

### Flow

1. **Parse flags**  
   * `--file` → `filename`: path to a file containing test IDs or patterns.  
   * `--list`  → `listFlag`: whether to print only the list of matching tests.

2. **Find matching test IDs**  
   Calls `getMatchingTestIDs(filename)` which returns a slice of strings.  
   If no matches, an error is returned via `Errorf`.

3. **Handle `--list`**  
   When set, `printTestList(matchingTestIDs)` prints a simple list and the function exits.

4. **Retrieve full descriptions**  
   For each ID in `matchingTestIDs`, call `getTestDescriptionsFromTestIDs(ids)` to obtain detailed test objects.

5. **Validate results**  
   If no descriptions are returned, an error is reported.

6. **Format output**  
   * `adjustLineMaxWidth()` calculates a suitable maximum line width for the terminal.  
   * For each description, `printTestCaseInfoBox(description)` outputs a formatted box containing all relevant test details.

7. **Return nil on success** or an error if any step fails.

### Side Effects
* Reads files (if `--file` is provided).  
* Writes to stdout (via the print helpers).  
* Modifies global `lineMaxWidth` used by formatting functions.

### Relationship to the Package
The `info` package provides a CLI command that lets users inspect the test suite.  
`showInfo` is the executor behind that command, orchestrating flag parsing, data retrieval, and pretty‑printing. It ties together helper utilities (`getMatchingTestIDs`, `printTestCaseInfoBox`) with Cobra’s command framework to expose this functionality to end‑users.
