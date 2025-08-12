ObjectSpec` – A lightweight key/value container for JSON output

### Purpose
`ObjectSpec` is a minimal helper type used by the **failures** sub‑package of *certsuite* (located at  
`github.com/redhat-best-practices-for-k8s/certsuite/cmd/certsuite/claim/show/failures`).  
Its role is to accumulate arbitrary key/value pairs that describe a test failure or diagnostic
message and then serialize them into JSON for downstream consumption (e.g., printing to the console,
writing to a file, or sending over an API).

The type intentionally avoids using a map so that the order of fields can be preserved.  
This ordering is useful when generating human‑readable logs where the sequence conveys context
(e.g., first the error code, then details, then remediation steps).

---

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `Fields` | `[]struct{Key, Value string}` | A slice that stores each field as a small struct containing a key and its associated value. The order in the slice is the order of insertion. |

> **Note**: The inner struct has no exported fields; only `ObjectSpec`’s methods manipulate it.

---

### Methods

| Method | Signature | What it does | Key interactions |
|--------|-----------|--------------|-------------------|
| `AddField` | `func (o *ObjectSpec) AddField(key, value string)` | Appends a new key/value pair to the `Fields` slice. | Uses Go’s built‑in `append`. No side effects beyond mutating the receiver. |
| `MarshalJSON` | `func (o ObjectSpec) MarshalJSON() ([]byte, error)` | Implements `json.Marshaler`. Builds a JSON object from the stored fields. It iterates over `Fields`, formats each pair as `"key":"value"`, and concatenates them into a single JSON string. Returns the byte slice and any error (never returns an error in current implementation). | Calls `len` to pre‑allocate buffer size, `fmt.Sprintf` to format key/value pairs. No external dependencies. |

> **Side effect**: `MarshalJSON` does not modify the receiver; it only reads from `Fields`.

---

### How it fits into the package

- **Data flow**:  
  1. Other code in *failures* creates an `ObjectSpec`, e.g., when building a failure report.  
  2. It calls `AddField` repeatedly to add details such as `"error_code":"E1234"`, `"message":"Invalid cert"`, etc.  
  3. When the report needs to be emitted, `MarshalJSON` is called (directly or via `json.Marshal`) to produce a JSON representation that preserves insertion order.

- **Dependencies**: The struct itself has no external dependencies beyond the Go standard library (`fmt`, `encoding/json`). It is deliberately lightweight so it can be reused in other parts of the CLI without pulling in heavier packages.

---

### Suggested Mermaid diagram

```mermaid
classDiagram
    class ObjectSpec {
        +[]{Key,Value} Fields
        +AddField(string,string) void
        +MarshalJSON() ([]byte,error)
    }
```

This diagram shows `ObjectSpec` as a container holding an ordered slice of key/value pairs and exposing two operations: adding fields and serializing to JSON.

---

### Summary

`ObjectSpec` is a small, order‑preserving data holder used in the *failures* sub‑package to construct structured diagnostics. Its API is intentionally simple—just add fields and marshal to JSON—making it straightforward to use anywhere test results need to be emitted in a predictable, readable format.
