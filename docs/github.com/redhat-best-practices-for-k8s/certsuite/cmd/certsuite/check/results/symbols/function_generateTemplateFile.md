generateTemplateFile` – results package

### Purpose
Creates a JSON template file that represents the expected shape of test results.
The function is used by the *check* command to pre‑populate a results file with all
possible test names set to an empty string, so that users can later fill in the actual
values.  
It writes the file to disk with controlled permissions (`TestResultsTemplateFilePermissions`).

### Signature
```go
func generateTemplateFile(map[string]string) error
```

| Parameter | Type | Meaning |
|-----------|------|---------|
| `testNames` | `map[string]string` | A map whose keys are the names of tests that should appear in the template. The values are ignored; they are replaced with empty strings when writing the file. |

The function returns an error if any step (encoding, I/O) fails.

### Key steps
1. **Build a JSON‑serialisable map**  
   For each test name supplied, it appends a key/value pair to a temporary slice,
   then joins them into a string that will be wrapped in a `map[string]string`.  
   The values are always set to an empty string (`""`).

2. **Encode the data**  
   A JSON encoder is created with indentation for readability.  
   The encoder writes into a `bytes.Buffer`.

3. **Write the file**  
   The buffer contents are written to `TestResultsTemplateFileName`
   using `os.WriteFile`, applying `TestResultsTemplateFilePermissions`.

### Dependencies
| Dependency | Role |
|------------|------|
| `encoding/json.NewEncoder` | Serialises the map to JSON |
| `bytes.Buffer` | Holds intermediate JSON data |
| `os.WriteFile` | Persists the template to disk |

The function also references package‑level constants:
* `TestResultsTemplateFileName`
* `TestResultsTemplateFilePermissions`

### Side effects
* Creates/overwrites a file on disk (`results.json` by default).
* The file’s permissions are set according to the constant.
* No global state is modified.

### How it fits in the package
The *results* sub‑package provides tooling for handling test results.  
`generateTemplateFile` is called during command initialisation (e.g., when a user runs `certsuite check results --template`) to give them a scaffold file that lists every possible test name. This helps standardise result reporting and prevents typos or missing fields in the final output.
