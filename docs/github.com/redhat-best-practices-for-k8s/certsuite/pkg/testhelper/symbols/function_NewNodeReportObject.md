NewNodeReportObject`

```go
func NewNodeReportObject(name string, reason string, compliant bool) *ReportObject
```

### Purpose  
Creates a `*ReportObject` that represents the outcome of evaluating a single Kubernetes **node**.  
The function is part of the `testhelper` package and is used by test suites to build structured compliance reports.

### Parameters  

| Name   | Type   | Description |
|--------|--------|-------------|
| `name` | `string` | The node name that is being reported on. |
| `reason` | `string` | A human‑readable explanation of why the node passed or failed the check. |
| `compliant` | `bool` | `true` if the node meets the policy, `false` otherwise. |

### Return value  

* `*ReportObject` – a fully populated report object that contains:
  * The node name in its `Name` field.
  * A single field with key `ReasonForCompliance` or `ReasonForNonCompliance` (depending on `compliant`) and the supplied reason as its value.
  * The compliance status stored in the object’s internal state.

### Implementation details

1. **Create base report**  
   ```go
   ro := NewReportObject()
   ```
   This calls the helper that initializes a `ReportObject` with empty fields, status set to unknown and an empty field map.

2. **Set node name**  
   ```go
   ro.Name = name
   ```

3. **Add compliance reason**  
   The function chooses the correct key based on `compliant`:
   ```go
   if compliant {
       ro.AddField(ReasonForCompliance, reason)
   } else {
       ro.AddField(ReasonForNonCompliance, reason)
   }
   ```
   `AddField` inserts a key/value pair into the report’s field map.

4. **Return**  
   The fully populated object is returned to the caller.

### Dependencies

| Dependency | Role |
|------------|------|
| `NewReportObject()` | Provides an empty `ReportObject`. |
| `AddField(key, value)` | Adds a key/value pair to the report’s field map. |

No global variables or side‑effects are used beyond these helpers.

### How it fits the package

* The `testhelper` package supplies utilities for constructing compliance reports that can be serialised and compared in tests.
* `NewNodeReportObject` is one of several convenience constructors (there are similar functions for pods, operators, etc.) that pre‑populate a report with a specific entity type and its status.  
* Tests call this function to create deterministic report objects, which are then marshalled into JSON or used by assertion helpers.

### Example usage

```go
// In a test:
nodeReport := NewNodeReportObject("worker-1", "CPU limits missing", false)

// nodeReport now contains:
//   Name:      "worker-1"
//   Fields:    map[string]string{"ReasonForNonCompliance": "CPU limits missing"}
//   Status:    FAILED (implied by the field key)
```

The returned object can be marshalled or compared against expected values in test assertions.
