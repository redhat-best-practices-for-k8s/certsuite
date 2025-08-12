getMatchingTestIDs` – Package *info*

| Item | Details |
|------|---------|
| **Location** | `cmd/certsuite/info/info.go:124-140` |
| **Signature** | `func getMatchingTestIDs(filter string) ([]string, error)` |

### Purpose
Given a user‑supplied filter expression (a string that may contain label selectors or other test metadata predicates), this helper returns the list of internal test identifiers that satisfy the expression.  
It is used by the CLI command that displays information about available tests.

### Inputs & Outputs

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `filter`  | `string` | A filter expression understood by CertSuite’s label evaluator (e.g., `"category=web,platform=k8s"`). If empty, all test IDs are returned. |

| Return | Type | Description |
|--------|------|-------------|
| `[]string` | list of matching test identifiers (`CheckID`s) |
| `error` | non‑nil if parsing the filter or loading the database fails. |

### Key Dependencies

| Called function | Role |
|-----------------|------|
| `InitLabelsExprEvaluator()` | Sets up the expression engine used to evaluate label predicates against tests. |
| `LoadInternalChecksDB()` | Loads the internal test database (metadata + labels). |
| `FilterCheckIDs()` | Applies the filter expression to the loaded DB and returns the subset of IDs that match. |
| `Errorf()` | Wraps errors with context before returning them. |

### Side‑Effects & Assumptions

* The function **does not modify** global state or the database; it only reads from the internal checks DB.
* It assumes that `InitLabelsExprEvaluator` and `LoadInternalChecksDB` succeed; if either fails, an error is returned immediately.
* If the filter string is syntactically invalid for the evaluator, the function returns a wrapped error.

### How it fits the package

The *info* command provides users with introspection on CertSuite tests.  
`getMatchingTestIDs` is the core routine that turns a user’s textual filter into concrete test identifiers, which are then used to fetch detailed metadata or display summaries. It bridges CLI input → evaluation engine → database lookup, keeping the rest of the command logic clean and focused on presentation.

---

#### Suggested Mermaid Diagram (for internal docs)

```mermaid
graph TD
  A[User enters filter] --> B[getMatchingTestIDs]
  B --> C{InitLabelsExprEvaluator}
  B --> D{LoadInternalChecksDB}
  B --> E[FilterCheckIDs]
  E --> F[Return []string, error]
```
This diagram visualises the call flow and key sub‑steps performed by `getMatchingTestIDs`.
