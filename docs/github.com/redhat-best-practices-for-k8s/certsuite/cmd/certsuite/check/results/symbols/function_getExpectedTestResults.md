getExpectedTestResults` – Package *results*

| Aspect | Details |
|--------|---------|
| **Signature** | `func getExpectedTestResults(file string) (map[string]string, error)` |
| **Visibility** | Unexported – used only inside the package. |

#### Purpose
The function reads a JSON file that lists expected test results for CertSuite checks and converts it into an in‑memory map.

* The input path (`file`) must point to a JSON file containing an array of objects, each with at least `name` (test identifier) and `expectedResult` fields.  
* The output is a `map[string]string` where the key is the test name and the value is its expected result string (e.g., `"pass"`, `"fail"`).

The map is used by the command logic to compare actual results against expectations when generating the final report.

#### Inputs
| Parameter | Type   | Description |
|-----------|--------|-------------|
| `file`    | `string` | File system path to the JSON file holding expected test results. |

#### Outputs
| Return | Type      | Meaning |
|--------|-----------|---------|
| `map[string]string` | Map of test name → expected result string. |
| `error` | Non‑nil if: <br>• The file cannot be read (e.g., does not exist, permission denied).<br>• JSON parsing fails.<br>• Any test entry lacks required fields. |

#### Key Dependencies
1. **`ioutil.ReadFile`** – reads the entire file into memory.
2. **`encoding/json.Unmarshal`** – parses the JSON payload into a slice of structs (`testResultJSON` is inferred from usage).
3. **`fmt.Errorf`** – constructs descriptive error messages for read or parse failures.
4. **Standard library `make`** – allocates the result map.

No external packages are involved; all operations use Go’s standard library.

#### Side Effects
* None on program state—only I/O (file reading) and local variable allocation.
* Errors propagate up to the caller for handling or termination of the command.

#### How It Fits the Package
The `results` package implements the CLI sub‑command that compares live test outputs with expected outcomes.  
- The **expected results** are loaded once via `getExpectedTestResults`.  
- Subsequent logic (not shown here) iterates over actual check results, looks up each in this map, and marks them as pass/fail/miss/skip accordingly.

This function is a small, pure helper that keeps file‑I/O separate from the command’s business logic.
