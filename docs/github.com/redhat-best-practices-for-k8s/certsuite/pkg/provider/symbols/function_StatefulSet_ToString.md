StatefulSet.ToString` – Human‑Readable Description of a StatefulSet

### Purpose
`ToString` is a helper that returns a concise string representation of a `StatefulSet`.  
It is used by the test framework when printing diagnostic information (e.g., in logs, test failures or debugging output). The method does **not** modify the receiver; it only formats existing data.

### Signature
```go
func (ss StatefulSet) ToString() string
```

| Part | Description |
|------|-------------|
| `ss` | The `StatefulSet` value to stringify. |

### Implementation Detail
The method is implemented in `pkg/provider/statefulsets.go`.  
It simply calls the standard library’s `fmt.Sprintf`:

```go
return fmt.Sprintf("%s/%s", ss.Namespace, ss.Name)
```

So it concatenates the namespace and name of the StatefulSet separated by a slash.

### Dependencies
- **Standard Library**: uses `fmt.Sprintf`.
- **StatefulSet struct**: expects the receiver to expose exported fields `Namespace` and `Name`.  
  These fields are populated when a `StatefulSet` is constructed from Kubernetes API objects elsewhere in the package.

No global variables, constants, or other functions are referenced inside `ToString`.

### Side Effects
None. The function is pure: it reads data from its receiver and returns a new string.

### Usage Context
- **Logging / Debugging**: Whenever a StatefulSet needs to be identified in log messages, the framework calls `ToString` to get a readable identifier.
- **Test Reporting**: In failure reports or test summaries, the string is used to refer to the resource under examination.

This method fits into the *provider* package as part of the abstraction that turns raw Kubernetes objects into lightweight Go structs. The `StatefulSet` type is one of several such abstractions (others include Deployments, DaemonSets, etc.), and each provides a `ToString` for uniform reporting.
