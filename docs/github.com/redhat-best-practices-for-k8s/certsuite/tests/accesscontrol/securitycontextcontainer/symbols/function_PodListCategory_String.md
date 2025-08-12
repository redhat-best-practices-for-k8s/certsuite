PodListCategory.String`

| Item | Detail |
|------|--------|
| **Receiver** | `c PodListCategory` – the enum value that represents a category of pod list tests. |
| **Signature** | `func (c PodListCategory) String() string` |
| **Exported?** | Yes – it is part of the public API of the `securitycontextcontainer` test package. |

### Purpose
Converts a `PodListCategory` value into a human‑readable string that describes which set of security‑context tests are being executed for a pod list.  
The function is used by the test harness when printing logs or generating reports, ensuring that the output reflects the correct category name rather than an opaque numeric value.

### Inputs
* `c` – The enum instance on which the method is called.  
  It can be any of the predefined constants (e.g., `CategoryID1`, `CategoryID2`, …).  
  If an undefined or unknown value is passed, it falls back to `"Undefined"` because that constant is defined in the same file.

### Output
* A string representation such as:
  * `"OK/OKNOK"`
  * `"OK/NOK"`
  * `"Undefined"`

The actual mapping between enum values and strings is hard‑coded inside the method using a `switch` on `c`.  

```go
func (c PodListCategory) String() string {
    switch c {
    case CategoryID1:
        return fmt.Sprintf("%s/%s", OKString, OKNOKString)
    case CategoryID2:
        return fmt.Sprintf("%s/%s", OKString, NOKString)
    ...
    default:
        return "Undefined"
    }
}
```

### Key Dependencies
* **`fmt.Sprintf`** – used to concatenate the status strings with a slash separator.
* **Constants** – `OKString`, `OKNOKString`, `NOKString` (and their numeric counterparts) defined elsewhere in the file.  
  These provide the actual text for each category.

### Side Effects
None. The function is pure: it only reads its receiver and constant values, returning a new string without modifying any global state or performing I/O.

### Package Context
The `securitycontextcontainer` package contains a suite of tests that validate how Kubernetes pods enforce security contexts.  
Each test case is grouped into categories (`PodListCategory`) to indicate the expected outcome (e.g., “OK” or “NOK”) for different security‑context configurations.  
`String()` supplies a readable description of these categories, which is crucial for:

* **Test output** – developers can quickly see which category failed or passed.
* **Logging** – logs include category names instead of raw enum values.
* **Reporting tools** – external systems that consume test results rely on the string representation.

By keeping this method simple and free of side effects, the package ensures deterministic behavior across all test executions.
