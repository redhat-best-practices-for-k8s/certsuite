NewReportObject`

**Location**

```
github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper
```

```go
func NewReportObject(reason string, typ string, isCompliant bool) *ReportObject
```

### Purpose

Creates a new **`ReportObject`** instance that represents the result of a test check.  
The object stores:

| Field | Meaning |
|-------|---------|
| `ReasonForCompliance`  | Reason when the check passes (`isCompliant == true`) |
| `ReasonForNonCompliance` | Reason when the check fails (`isCompliant == false`) |

Both fields are added through the helper method **`AddField`**.

### Parameters

| Name        | Type   | Description |
|-------------|--------|-------------|
| `reason`    | `string` | Human‑readable explanation for the result. |
| `typ`       | `string` | A short tag identifying the check type (e.g., `"NetworkPolicy"`, `"PodSecurity"`). |
| `isCompliant` | `bool` | `true` if the test passed, `false` otherwise. |

### Return value

* `*ReportObject` – a pointer to the freshly allocated object.

The returned object is fully initialized with:

- `Type` set to `typ`.
- One of the two reason fields populated based on `isCompliant`.

No other side‑effects occur.

### Key dependencies

| Called function | Purpose |
|-----------------|---------|
| `AddField` (twice) | Adds a key/value pair to the internal map that holds report metadata. |

The implementation does not reference any global variables or external state; it is fully deterministic from its arguments.

### Usage context

In the *testhelper* package, test cases build a slice of `ReportObject`s to aggregate results.  
`NewReportObject` simplifies object creation by handling:

1. Initializing the struct.
2. Assigning the appropriate reason key based on compliance status.
3. Populating the type field.

Example (simplified):

```go
// Pass case
ok := NewReportObject("All pods use read‑only rootfs", "PodSecurity", true)

// Fail case
fail := NewReportObject("Container runs as root", "PodSecurity", false)
```

These objects are later serialized or printed by other utilities in the package.

### Diagram (optional)

```mermaid
flowchart TD
    A[Call NewReportObject] --> B{isCompliant?}
    B -- true --> C[AddField(ReasonForCompliance, reason)]
    B -- false --> D[AddField(ReasonForNonCompliance, reason)]
    C & D --> E[Return *ReportObject]
```

This function is a small but essential helper that enforces consistent reporting across all tests.
