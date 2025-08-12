PrintNetTestContextMap`

```go
func PrintNetTestContextMap(m map[string]NetTestContext) string
```

### Purpose  
`PrintNetTestContextMap` produces a human‑readable representation of the entire `netcommons.NetTestContext` map. It is intended for debugging, logging or test reporting where the caller needs to inspect all configured network tests at a glance.

### Parameters  

| Name | Type | Description |
|------|------|-------------|
| `m`  | `map[string]NetTestContext` | The full mapping from test identifiers (strings) to their corresponding `NetTestContext` values. |

> **Note**: The function does not modify the map; it only reads its contents.

### Return Value  

* A single string containing a formatted list of all key/value pairs in `m`.  
  Each entry is rendered as:

```
<key>: <value.String()>
```

and entries are separated by newlines. If the map is empty, the function returns `"Empty NetTestContext Map"`.

### Implementation details  

The routine builds the output using a `strings.Builder` for efficient string concatenation:

1. **Header** – Prints the total number of tests with `len(m)`.
2. **Iteration** – For each key/value pair:
   - Writes `<key>:` followed by the result of `value.String()`.  
     The `String()` method is defined on `NetTestContext` and returns a concise summary of that test’s configuration.
3. **Separator** – Each entry ends with a newline (`\n`).

All operations are pure; there are no side effects beyond the returned string.

### Dependencies  

| Dependency | Role |
|------------|------|
| `len` | Determines how many entries exist in the map for the header line. |
| `strings.Builder.WriteString` | Appends raw strings to the builder. |
| `fmt.Sprintf` | Formats numeric values and struct fields into readable text. |
| `NetTestContext.String()` | Provides a string representation of individual test contexts. |

No global variables are accessed; the function is fully self‑contained.

### How it fits the package  

The `netcommons` package centralizes networking utilities for CertSuite tests.  
- **`NetTestContext`** represents configuration for a single network test (interfaces, IP families, ports, etc.).  
- The map passed to `PrintNetTestContextMap` is typically produced by test setup functions or collected during runtime.  
- This helper turns that raw data into an easily consumable log entry, aiding developers in diagnosing failures or verifying expected test setups.

```mermaid
flowchart TD
    A[Call PrintNetTestContextMap] --> B{Iterate over map}
    B --> C[Write key]
    C --> D[Call NetTestContext.String()]
    D --> E[Append to builder]
    E --> F[Return final string]
```

---

**Usage example**

```go
ctxMap := BuildAllNetTestContexts()
log.Printf("Network test setup:\n%s", netcommons.PrintNetTestContextMap(ctxMap))
```

This prints a neatly formatted snapshot of all configured network tests.
