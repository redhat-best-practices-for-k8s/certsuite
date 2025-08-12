updateCrUnderTest`

| | |
|-|-|
|**File**|`pkg/provider/provider.go:414`|
|**Package**|`provider` (internal helper)|
|**Exported?**|No – used only within the package|

### Purpose
`updateCrUnderTest` is a small transformation helper that takes a slice of **autodiscover.ScaleObject**s (the type returned by the discovery routine that scans a cluster for Scale objects such as Deployments, DaemonSets, etc.) and converts it into the package‑specific `ScaleObject` representation used elsewhere in the provider logic.  
The function is called after the autodiscovery step finishes to prepare the list of Scale objects that will be used by the test harness when applying workloads under test.

### Signature
```go
func updateCrUnderTest([]autodiscover.ScaleObject) []ScaleObject
```

- **Input**: a slice of `autodiscover.ScaleObject`.  
  Each element contains:
  * `Name` – resource name  
  * `Namespace` – namespace where the object lives  
  * `Kind` – type (Deployment, DaemonSet, etc.)  

- **Output**: a new slice of `ScaleObject` (defined in this package).  
  The returned objects preserve the same fields as the input but are wrapped in the provider’s own struct so that other code can rely on a single internal type.

### Implementation details
```go
func updateCrUnderTest(scaleObjects []autodiscover.ScaleObject) []ScaleObject {
    var res []ScaleObject
    for _, s := range scaleObjects {
        res = append(res, ScaleObject{
            Name:      s.Name,
            Namespace: s.Namespace,
            Kind:      s.Kind,
        })
    }
    return res
}
```
* The function allocates a new slice and appends each converted element.  
* No global state is read or modified – it is pure from the perspective of side‑effects.

### Dependencies & Side Effects
| Dependency | Role |
|------------|------|
| `autodiscover.ScaleObject` | Source type for conversion |
| `ScaleObject` (internal)     | Destination type |

The function has **no side effects** on globals or external state. It only constructs and returns a new slice.

### How it fits the package
1. **Discovery phase** – The provider calls an autodiscover routine that scans the cluster for scalable resources, producing `[]autodiscover.ScaleObject`.  
2. **Transformation** – `updateCrUnderTest` converts that output into the provider’s internal representation (`[]ScaleObject`).  
3. **Testing phase** – Subsequent functions (e.g., workload deployment, health checks) operate on this slice to manage and verify Scale objects.

Thus, `updateCrUnderTest` is a bridge between the discovery subsystem and the test execution logic, ensuring type consistency across the package.
