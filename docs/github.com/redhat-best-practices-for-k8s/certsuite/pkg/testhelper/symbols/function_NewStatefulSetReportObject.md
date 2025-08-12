NewStatefulSetReportObject`

### Purpose
Creates a **`ReportObject`** that represents the outcome of a compliance check on a Kubernetes StatefulSet.  
The returned object is ready to be serialized (e.g., JSON) and sent to the CertSuite server.

### Signature
```go
func NewStatefulSetReportObject(namespace, statefulSetName, reason string, compliant bool) *ReportObject
```

| Parameter | Type   | Meaning |
|-----------|--------|---------|
| `namespace`      | `string` | Namespace of the StatefulSet. |
| `statefulSetName`| `string` | Name of the StatefulSet being evaluated. |
| `reason`         | `string` | Text explaining why the StatefulSet passed or failed compliance. |
| `compliant`      | `bool`   | `true` if the StatefulSet meets all rules; otherwise `false`. |

### Implementation Flow
1. **Instantiate** a generic report object via `NewReportObject()`.
2. **Populate common fields**:
   - `Namespace`, `Name`, and `ReasonForCompliance/NonCompliance` are added using `AddField`.
3. **Set the compliance status** (`Compliant`) through `AddField`.
4. **Return** the fully‑built `*ReportObject`.

### Dependencies
- **`NewReportObject()`** – returns a base report object.
- **`AddField(key, value)`** – attaches key/value pairs to the report.

No other global variables or types are accessed directly in this function.

### Side Effects & Constraints
- The function is pure: it only constructs and returns a new `ReportObject`.  
- It does **not** modify any external state (e.g., Kubernetes objects, logs).  
- Caller must handle the returned pointer; if `nil` is returned by `NewReportObject`, an error would propagate.

### How It Fits the Package
The `testhelper` package supplies helpers to generate compliance reports for various Kubernetes resources.  
`NewStatefulSetReportObject` specializes that logic for StatefulSets, mirroring similar constructors for Deployments, DaemonSets, etc.  

These objects are later marshaled into JSON and sent as part of a CertSuite test run payload.
