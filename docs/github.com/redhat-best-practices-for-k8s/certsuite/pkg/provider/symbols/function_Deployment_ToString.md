Deployment.ToString`

```go
func (d Deployment) ToString() string
```

### Purpose  
`ToString` returns a human‑readable description of a `Deployment` object. It is used by the provider when printing diagnostic information or logging deployment status. The output contains the deployment’s **namespace** and **name**, formatted as:

```
<Namespace>/<Name>
```

This string representation is stable across program runs because it relies only on immutable fields of the `Deployment` struct.

### Receiver  
- `d Deployment`: a value (not pointer) representing a Kubernetes deployment.  
  The struct holds at least two exported fields used by this method:
  - `Namespace string`
  - `Name      string`

### Return Value  
- `string`: formatted as `"namespace/name"`.

### Dependencies  
- Uses the standard library function **`fmt.Sprintf`** to format the output.
- Relies on the struct’s exported fields; no other global variables or functions are referenced.

### Side‑Effects  
None. The method is pure: it only reads from the receiver and returns a string.

### Package Context  
The `provider` package models Kubernetes objects for certsuite’s test engine.  
`Deployment.ToString` provides a convenient way to identify deployments in logs, reports, or user interfaces. It complements other similar methods (e.g., on services, pods) that expose readable identifiers across the provider model.
