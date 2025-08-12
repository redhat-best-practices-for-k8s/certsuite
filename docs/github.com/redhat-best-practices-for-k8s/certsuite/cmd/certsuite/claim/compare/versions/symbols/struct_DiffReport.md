DiffReport` – Summary of Version Comparisons

The **`DiffReport`** type is the public representation of a comparison between two
`officialClaimScheme.Versions`.  
It lives in the *versions* sub‑package (`compare/versions`) and is the only
exported struct that callers use to inspect the result of `Compare`.

| Field | Type | Meaning |
|-------|------|---------|
| **Diffs** | `*diff.Diffs` | The raw diff data produced by the underlying comparison logic.  
  It contains the detailed changes (added/removed/modified) between the two
  version objects. |

> **Note**: `diff.Diffs` is defined in another package (`github.com/.../diff`) and
> encapsulates the low‑level representation of differences.

## Purpose

*Encapsulate* the comparison result so that callers can:
- Inspect the raw diff tree via the `Diffs` field.
- Convert the report to a human‑readable string with `String()`.

This struct is deliberately thin; all heavy lifting happens inside
[`Compare`](#compare). It acts as an opaque handle for the comparison outcome.

## Methods

### `func (d DiffReport) String() string`

* **Input**: none – operates on the receiver.
* **Output**: a formatted string that represents the diff report.
  Internally it delegates to two functions named `String` (likely
  helpers in the same package or imported ones). The exact formatting is
  defined elsewhere; this method simply forwards the call.

### Dependencies & Side‑Effects

| Dependency | Role |
|------------|------|
| `diff.Diffs.String()` | Produces a textual representation of the diff tree. |
| `fmt.Sprintf` (or similar) | Used inside the helper functions to build the final string. |

No global state is mutated; all operations are pure with respect to
the struct’s data.

## Interaction with the Package

```mermaid
flowchart TD
    Compare -->|returns| DiffReport
    DiffReport.Diffs --> diff.Diffs
    DiffReport.String() --> diff.Diffs.String()
```

1. **`Compare(a, b *officialClaimScheme.Versions) *DiffReport`**  
   - Serializes `a` and `b` to JSON (`Marshal`).  
   - Deserializes them back (`Unmarshal`) into generic maps to prepare for comparison.  
   - Calls an internal `diff.Compare` function that returns a `*diff.Diffs`.  
   - Wraps that result in a new `DiffReport{Diffs: d}`.

2. **`DiffReport.String()`**  
   - Converts the embedded `Diffs` into a user‑friendly string, typically for
     logging or test output.

The struct therefore sits at the boundary between raw diff data and
consumer code that needs either structured access (`Diffs`) or a printable
summary (`String`).
