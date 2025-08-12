ReportObject.AddField`

| Item | Detail |
|------|--------|
| **Signature** | `func (ro ReportObject) AddField(key string, value string) *ReportObject` |
| **Exported?** | Yes |
| **File / Line** | `pkg/testhelper/testhelper.go:377` |

### Purpose
`AddField` is a helper method that augments a `ReportObject` with an additional key‑value pair.  
The method is used throughout the test‑generation workflow to capture dynamic data (e.g., runtime metrics, compliance reasons, resource names) and attach it to a report before serialisation.

### Inputs
| Parameter | Type   | Description |
|-----------|--------|-------------|
| `key`     | `string` | The field name that will be stored in the report. |
| `value`   | `string` | The corresponding value for the field. |

Both arguments are simple strings; no validation is performed.

### Output
The method returns a pointer to the same `ReportObject`.  
Returning a pointer allows callers to chain calls:

```go
ro := NewReportObject().
    AddField("name", "pod1").
    AddField("status", "running")
```

### Internal Behaviour (Side‑effects)
1. **Key Append** – The supplied `key` is appended to the slice `ObjectFieldsKeys`.
2. **Value Append** – The supplied `value` is appended to the slice `ObjectFieldsValues`.

The slices are stored inside the receiver struct, so the state of the object changes in place.

### Dependencies
* Uses the built‑in `append` function twice.
* Relies on the `ReportObject` type having exported fields:
  * `ObjectFieldsKeys []string`
  * `ObjectFieldsValues []string`

No external packages are imported by this method.

### Context within the package
The `testhelper` package provides utilities for constructing test data and reports.  
`ReportObject` represents a generic report structure that is serialised to JSON/YAML.  
`AddField` is the primary mechanism for populating this structure with arbitrary metadata during test execution, making it central to how tests convey results back to the framework.

### Example
```go
// Build a simple compliance report
report := NewReportObject().
    AddField("resource", "pod").
    AddField("name", "nginx-pod").
    AddField("status", "compliant")

json, _ := json.MarshalIndent(report, "", "  ")
fmt.Println(string(json))
```

The resulting JSON will contain the key/value pairs in the order they were added.
