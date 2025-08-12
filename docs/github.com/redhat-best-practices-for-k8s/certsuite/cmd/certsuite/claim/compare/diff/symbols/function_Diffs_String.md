Diffs.String` – Human‑readable diff report

| Element | Description |
|---------|-------------|
| **Package** | `diff` (github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/compare/diff) |
| **Receiver type** | `Diffs` – a struct that holds the computed differences between two claim files. |
| **Signature** | `func (d Diffs) String() string` |

---

### Purpose
Implements the `fmt.Stringer` interface for `Diffs`.  
The method returns a single string that represents the comparison result as a neatly formatted table:

```
<name>: Differences
FIELD                           CLAIM 1     CLAIM 2
/jsonpath/to/field1             value1      value2
...

<name>: Only in CLAIM 1
/jsonpath/to/field/in/claim1/only

<name>: Only in CLAIM 2
/jsonpath/to/field/in/claim2/only
```

`<name>` is replaced with `d.Name`.  
The table columns for **FIELD** and **CLAIM 1** have widths that adapt to the longest path/value found.

---

### Inputs & Outputs
| Input | Type | Notes |
|-------|------|-------|
| `d` (receiver) | `Diffs` | Contains: <br>• `Name string` – identifier used in headings.<br>• `Differences []FieldDiff` – paths that differ with their two values.<br>• `OnlyInClaim1 []string`, `OnlyInClaim2 []string` – paths present only in one claim. |

| Output | Type | Description |
|--------|------|-------------|
| string | The formatted report. | No other side effects. |

---

### Key Dependencies
* **Standard library**
  * `fmt.Sprint`, `fmt.Sprintf`: build the table lines.
  * `len` (built‑in): calculate width of longest field/value for dynamic column sizing.

No external packages are used; the method relies solely on data already stored in the receiver.

---

### Side Effects
* Pure function – does **not** modify `d` or any global state.
* The only effect is the creation and return of a string representation.

---

### How it Fits in the Package
The `diff` package computes differences between two claim files.  
Other functions (e.g., `CompareClaims`) populate a `Diffs` instance; this method provides a convenient, human‑readable summary that can be printed to stdout or logged.  

Typical usage:

```go
differences := diff.CompareClaims(claim1, claim2)
fmt.Println(differences.String())
```

---

### Suggested Mermaid Diagram (optional)

```mermaid
graph TD
  A[Claim 1] -->|diffs| B(Diffs)
  C[Claim 2] -->|diffs| B
  B --> D[String()]
  D --> E[Console output]
```

This visualises that two claim inputs produce a `Diffs` value, whose `String()` method yields the final report.
