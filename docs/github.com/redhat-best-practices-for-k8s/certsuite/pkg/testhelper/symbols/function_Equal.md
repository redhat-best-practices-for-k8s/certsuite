Equal` – Report Object Comparator

### 📌 Purpose
`Equal` is a helper that checks whether two slices of `*ReportObject` contain the same elements in the same order.  
It is used by tests to assert that the output produced by a rule or validator matches an expected set of report objects.

> **Note:** The function performs a *deep* comparison of each element, so it will consider field values, nested structs, and pointer contents when determining equality.

### ⚙️ Signature
```go
func Equal(a []*ReportObject, b []*ReportObject) bool
```

| Parameter | Type                 | Description |
|-----------|----------------------|-------------|
| `a`       | `[]*ReportObject`   | First slice to compare. |
| `b`       | `[]*ReportObject`   | Second slice to compare. |

| Return | Type  | Description |
|--------|-------|-------------|
| `bool` | true if the slices are equal, false otherwise. |

### 🔗 Dependencies
- **Standard library**  
  - `len`: Used three times to quickly reject unequal lengths.
  - `reflect.DeepEqual`: Performs element‑by‑element deep equality checks.

No other functions or globals from the package influence this function.

### 🧩 How It Works (Pseudo‑code)

```text
if len(a) != len(b):
    return false

for i in range(len(a)):
    if !DeepEqual(a[i], b[i]):
        return false

return true
```

- **Length check**: If the slices differ in length, they cannot be equal.
- **Element comparison**: Uses `reflect.DeepEqual` to compare each corresponding pair of pointers. This accounts for all exported fields within `ReportObject`.

### 📦 Placement in the Package
The `testhelper` package bundles various utilities that simplify writing tests for CertSuite rules and validators.  
`Equal` sits alongside other helpers such as:

- `CreateReportObject`
- `GetExpectedReports`
- `GenerateJSON`

It is deliberately *read‑only* – it does not modify its arguments, making it safe to use in concurrent test scenarios.

### 🔎 When to Use
- **Unit tests** that compare the actual list of reports returned by a rule against a statically defined expected slice.
- **Integration tests** where order matters (e.g., when rules are deterministic and produce reports in a known sequence).

If only the *contents* matter and order is irrelevant, you might combine `Equal` with sorting or use a different helper that ignores ordering.

### 🛠️ Example
```go
expected := []*ReportObject{
    CreateReportObject(1, "PASS", "RuleA"),
    CreateReportObject(2, "FAIL", "RuleB"),
}

actual := rule.Execute()

if !Equal(expected, actual) {
    t.Errorf("reports differ")
}
```

---

#### Mermaid Diagram (Optional)
```mermaid
flowchart TD
  A[Input Slice a] --> B{len(a)==len(b)?}
  C[Input Slice b] --> B
  B -- no --> D[Return false]
  B -- yes --> E[Loop over indices i]
  E --> F{DeepEqual(a[i],b[i])?}
  F -- no --> D
  F -- yes --> G[Next i]
  G --> H{i < len(a)-1?}
  H -- no --> I[Return true]
```

This visualizes the straightforward equality logic employed by `Equal`.
