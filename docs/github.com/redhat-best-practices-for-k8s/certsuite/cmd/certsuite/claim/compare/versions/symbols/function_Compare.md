Compare` – Diffing two *official claim* version trees

```go
func Compare(a, b *officialClaimScheme.Versions) *DiffReport
```

The function lives in the `github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/versions` package.  
It is the entry point for computing a diff between two **Official Claim** schemas that are represented by the `Versions` struct (see `officialClaimScheme.Versions`). The returned `*DiffReport` contains all differences that were detected.

---

## Purpose

*   Serialize each input schema to JSON.
*   Unmarshal the JSON back into generic `map[string]interface{}` values – this normalises the structure and removes type information that is irrelevant for a structural diff.
*   Call the internal recursive comparison routine (`compare`) on those maps.
*   Return a populated `DiffReport` (or terminate the process with an error if any step fails).

This function is used by command‑line tools to report changes between two claim schema versions, e.g. when generating migration guides or validating compatibility.

---

## Inputs

| Parameter | Type | Description |
|-----------|------|-------------|
| `a` | `*officialClaimScheme.Versions` | First schema tree (the “source”). |
| `b` | `*officialClaimScheme.Versions` | Second schema tree (the “target”). |

Both are pointers to the same struct type defined in the *Official Claim* package. The function does **not** modify these inputs.

---

## Output

| Return | Type | Description |
|--------|------|-------------|
| `*DiffReport` | pointer to a `DiffReport` struct (defined in this package) | Holds lists of added, removed and modified nodes between the two schemas. A nil report indicates no differences were found. |

If any error occurs during marshalling/unmarshalling or the recursive comparison, the function logs the error via `log.Fatalf` and exits the program – it never returns an error value.

---

## Key Dependencies

| Dependency | Role |
|------------|------|
| `encoding/json.Marshal` / `json.Unmarshal` | Convert the strongly‑typed `Versions` structs to generic JSON maps. |
| `log.Fatalf` | Abort on any serialization/deserialization failure. |
| `compare` (internal helper) | Performs the actual structural diff between two `map[string]interface{}` objects. |

The function relies on the *Official Claim* schema definitions being serialisable to JSON, which is guaranteed by that package.

---

## Side Effects

*   The function **exits** the process if any error occurs (via `log.Fatalf`).
*   It does not modify its inputs or any global state.
*   The only observable side effect is the creation of a `DiffReport`.

---

## How it fits the package

The `versions` package provides tooling for comparing two versions of the official claim specification.  
`Compare` is the public API that callers use; all other helpers (`compare`, `addNode`, `removeNode`, etc.) are private and used internally to build the report.

Typical usage in a command‑line tool:

```go
old, _ := loadVersions("v1.yaml")
new, _ := loadVersions("v2.yaml")

report := versions.Compare(old, new)
printReport(report)
```

This keeps the comparison logic isolated from the rest of the application while still exposing a clean interface.
