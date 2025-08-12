containerPortNameFormatCheck`

| | |
|-|-|
| **File** | `tests/manageability/suite.go:94` |
| **Package** | `github.com/redhat-best-practices-for-k8s/certsuite/tests/manageability` |
| **Visibility** | unexported (private to the package) |
| **Signature** | `func containerPortNameFormatCheck(name string) bool` |

### Purpose
The function validates whether a Kubernetes *container port name* follows the naming convention required by the CertSuite test suite.  
In Kubernetes, a container port name must match the pattern:

```
<protocol>[-<suffix>]
```

where `<protocol>` is one of:
`grpc`, `grpc-web`, `http`, `http2`, `tcp`, `udp`.

The optional `<suffix>` may be any string chosen by the application.  
If the supplied `name` satisfies this pattern, the function returns `true`; otherwise it returns `false`.

### Parameters
| Name | Type | Description |
|------|------|-------------|
| `name` | `string` | The container port name to validate. |

### Return Value
| Type | Meaning |
|------|---------|
| `bool` | `true` if the format is valid, otherwise `false`. |

### Implementation Notes
* The function uses only the standard library `strings.Split` to separate the string on `"-"`.
* It then checks whether the first segment (the protocol part) exists in the package‑level slice `allowedProtocolNames` (declared at line 89).  
  *If the protocol is not one of the allowed names, the function immediately returns `false`.*
* If the protocol is allowed, the function accepts any optional suffix and therefore always returns `true`.

### Dependencies
| Dependency | Role |
|------------|------|
| `strings.Split` | Splits the port name on hyphens to isolate the protocol part. |
| `allowedProtocolNames` | Slice of strings containing valid protocol identifiers. |

### Side‑effects
None – the function is pure and only performs a logical check.

### Usage Context
The function is used by the test suite when validating container specifications in Kubernetes manifests or runtime objects. It ensures that any container port name supplied to tests conforms to the naming rules expected by CertSuite, thereby preventing false negatives caused by improperly formatted names.

### Example
```go
valid := containerPortNameFormatCheck("http")
fmt.Println(valid) // true

invalid := containerPortNameFormatCheck("ftp")
fmt.Println(invalid) // false
```

---

#### Mermaid diagram – Flow of `containerPortNameFormatCheck`

```mermaid
flowchart TD
    A[Input: name] --> B{Split on "-"}
    B --> C[protocol = first segment]
    C --> D{protocol ∈ allowedProtocolNames}
    D -- yes --> E[Return true]
    D -- no  --> F[Return false]
```

This succinctly captures the decision tree of the function.
