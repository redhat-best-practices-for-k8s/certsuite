buildCSV`

| Aspect | Detail |
|--------|--------|
| **Location** | `cmd/certsuite/claim/show/csv/csv.go:138` |
| **Signature** | `func (*claim.Schema, string, map[string]claimschema.TestCaseDescription) [][]string` |
| **Purpose** | Convert a claim schema into a CSV‑ready 2‑D slice of strings. The resulting rows contain the original claim data plus three derived columns:<br>• **Remediation** – the remediation text for each test case.<br>• **Mandatory/Optional** – whether the test case is mandatory or optional.<br>• **CNFType** – the type of CNF (Cloud Native Function) that the claim targets. |
| **Inputs** | 1. `s *claim.Schema` – The parsed claim file. <br>2. `cnfName string` – Name of the CNF to which the claim applies; used when adding the CNFType column.<br>3. `testCaseMap map[string]claimschema.TestCaseDescription` – A lookup that maps test‑case identifiers to their descriptions (containing remediation and requirement level). |
| **Outputs** | `[][]string`: each inner slice represents a CSV row, starting with the header when `addHeaderFlag` is true. The first three columns are the original claim fields; the next three are the added metadata. |
| **Key Steps** | 1. Iterate over every test case in the schema (`s.TestCases`).<br>2. For each, look up its description in `testCaseMap`. If missing, default to empty strings.<br>3. Append a new row containing:<br>   * original claim data<br>   * remediation from the map<br>   * `"Mandatory"` or `"Optional"` based on the description’s `Requirement` field<br>   * the supplied `cnfName`<br>4. If `addHeaderFlag` is set, prepend a header row with column names. |
| **Dependencies** | • `claim.Schema` – provides claim structure.<br>• `claimschema.TestCaseDescription` – holds remediation and requirement info.<br>• Global flag `addHeaderFlag` controls inclusion of the header. No other external packages are used directly in this function. |
| **Side Effects** | None: the function only constructs a slice; it does not modify the schema, map, or any global state. |
| **Package Context** | Part of the `csv` sub‑package under `cmd/certsuite/claim/show`. The package exposes a CLI command (`CSVDumpCommand`) that calls `buildCSV` to generate CSV data for display or export. This function is central to transforming internal claim representations into a flat, tabular form suitable for downstream tools or human consumption. |
