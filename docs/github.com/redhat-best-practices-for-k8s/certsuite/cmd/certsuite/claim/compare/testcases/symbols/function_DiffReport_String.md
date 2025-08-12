DiffReport.String() string`

| Item | Details |
|------|---------|
| **Package** | `testcases` – part of the CertSuite claim comparison tool (`github.com/redhat-best-practices-for-k8s/certsuite/...`). |
| **Receiver** | `r DiffReport` – a struct holding per‑claim test case results. |
| **Signature** | `func (r DiffReport) String() string` |
| **Exported** | Yes (`String`) – implements `fmt.Stringer`. |

### Purpose
Render a human‑readable summary of two claim test result sets side by side:

1. **Summary table** – counts of *passed*, *skipped* and *failed* cases for each claim.
2. **Diff table** – list of individual test cases that have differing outcomes between the two claims, showing the status in each claim.

The output is a plain‑text string suitable for console printing or logging.

### Inputs / State Used
| Field | Role |
|-------|------|
| `r.Cases1` | Map of test case name → result for CLAIM‑1. |
| `r.Cases2` | Map of test case name → result for CLAIM‑2. |
| `r.Summary1`, `r.Summary2` | Pre‑computed counters (`passed/failed/skipped`) for each claim (populated elsewhere). |

The method reads these fields; it does **not** modify any state.

### Outputs
A single string containing:

```
Test cases summary table:
STATUS   # in CLAIM-1  # in CLAIM-2
passed    X            Y
skipped   A            B
failed    C            D

Test cases with different results table:
TEST CASE NAME                 CLAIM-1   CLAIM-2
foo-test                       passed    failed
bar-test                       skipped   passed
...
```

The exact formatting uses `fmt.Sprintf` to align columns.

### Key Dependencies
* Standard library: `fmt.Sprintf`, `len`.
* No external packages.
* Relies on the struct fields being correctly populated before calling.

### Side Effects
None – purely read‑only. It only formats and returns a string.

### How it fits the package

The `testcases` package is responsible for comparing two sets of claim test results.  
- Other functions build the `DiffReport` by iterating over test cases, counting statuses, and recording differences.  
- `String()` provides the final presentation layer: once a `DiffReport` is ready, calling `fmt.Println(report)` or similar will output the formatted tables.

A minimal Mermaid diagram showing data flow:

```mermaid
flowchart TD
    A[Claim‑1 Test Cases] -->|count| B[Summary1]
    A -->|list| C[Cases1 Map]
    D[Claim‑2 Test Cases] -->|count| E[Summary2]
    D -->|list| F[Cases2 Map]
    B & E --> G[DiffReport{Summary1, Summary2, Cases1, Cases2}]
    G --> H[String() → formatted string]
```

The `String` method is the only public way to convert a `DiffReport` into a displayable form.
