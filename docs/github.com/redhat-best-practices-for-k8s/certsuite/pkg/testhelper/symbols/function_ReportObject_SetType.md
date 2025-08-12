ReportObject.SetType`

```go
func (r *ReportObject) SetType(aType string) *ReportObject
```

### Purpose  
`SetType` is a simple setter that records the type of a `ReportObject`.  
The *type* represents what kind of Kubernetes object or test artefact the
report entry refers to (e.g. `"pod"`, `"deployment"`, `"operator"`).

### Parameters  
| Name   | Type   | Description |
|--------|--------|-------------|
| `aType` | `string` | The desired value for the `ObjectType` field of the receiver.

### Return value  
* Returns a pointer to the same `ReportObject` that was mutated.  
  This allows chaining, e.g.:

```go
obj := (&ReportObject{}).SetType("pod").SetName("nginx")
```

### Side‑effects  
* Mutates the receiver’s `ObjectType` field in place.  
* No other state is changed and no external resources are accessed.

### Dependencies & Context  

| Item | Relationship |
|------|--------------|
| `ReportObject` struct | The method belongs to this type; it expects the struct to have an exported field named `ObjectType`. |
| None else | The function does not reference any other global variables or types. |

The `testhelper` package is a helper library for generating test reports.
`SetType` is part of that public API, allowing callers to annotate report
entries with their object type before further populating the rest of the
report fields.

### Usage in the Package  
Typical usage pattern:

```go
r := &ReportObject{}
r.SetType("pod")
r.Name = "my-pod"
```

or via method chaining:

```go
(&ReportObject{}).SetType("deployment").SetName("web-app")
```

Because it returns a pointer to the modified object, callers can build
reports in a fluent style without extra variable assignments.
