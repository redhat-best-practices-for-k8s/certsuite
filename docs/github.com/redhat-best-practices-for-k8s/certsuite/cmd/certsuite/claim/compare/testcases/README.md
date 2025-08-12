## Package testcases (github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/testcases)



### Structs

- **DiffReport** (exported) — 4 fields, 1 methods
- **TcResultDifference** (exported) — 3 fields, 0 methods
- **TcResultsSummary** (exported) — 3 fields, 0 methods

### Functions

- **DiffReport.String** — func()(string)
- **GetDiffReport** — func(claim.TestSuiteResults, claim.TestSuiteResults)(*DiffReport)

### Call graph (exported symbols, partial)

```mermaid
graph LR
  DiffReport_String --> Sprintf
  DiffReport_String --> Sprintf
  DiffReport_String --> Sprintf
  DiffReport_String --> Sprintf
  DiffReport_String --> Sprintf
  DiffReport_String --> Sprintf
  DiffReport_String --> Sprintf
  DiffReport_String --> Sprintf
  GetDiffReport --> getTestCasesResultsMap
  GetDiffReport --> getTestCasesResultsMap
  GetDiffReport --> getMergedTestCasesNames
  GetDiffReport --> append
  GetDiffReport --> getTestCasesResultsSummary
  GetDiffReport --> getTestCasesResultsSummary
```

### Symbol docs

- [struct DiffReport](symbols/struct_DiffReport.md)
- [struct TcResultDifference](symbols/struct_TcResultDifference.md)
- [struct TcResultsSummary](symbols/struct_TcResultsSummary.md)
- [function DiffReport.String](symbols/function_DiffReport_String.md)
- [function GetDiffReport](symbols/function_GetDiffReport.md)
