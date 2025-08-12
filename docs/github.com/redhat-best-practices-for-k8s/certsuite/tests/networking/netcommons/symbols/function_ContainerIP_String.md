ContainerIP.String` – Human‑readable representation of a container’s IP configuration  

```go
// String displays the ContainerIP data structure.
func (c ContainerIP) String() string
```

### Purpose
`String` is the standard *stringer* implementation for the `ContainerIP` type.  
It converts an instance of `ContainerIP` into a concise, human‑readable string that can be logged or printed in tests.

### Inputs & Outputs
| Parameter | Type         | Description |
|-----------|--------------|-------------|
| `c` (receiver) | `ContainerIP` | The IP configuration to format. |

**Return value**

- `string`: A formatted representation of the `ContainerIP`.  
  Example: `"eth0/IPv4 10.244.1.5"`.

### Key Dependencies
| Dependency | Role |
|------------|------|
| `fmt.Sprintf` | Builds the final string using a format verb (`%s`). |
| `c.StringLong()` | Provides the detailed description of the IP that is embedded in the short form. |

> **Note**: The implementation delegates the heavy lifting to `StringLong`, which returns a more verbose representation (including interface name, IP version, address, and any reserved ports). `String` simply wraps that output.

### Side‑Effects
- None.  
  The method is pure; it does not modify the receiver or any global state.

### Relationship within `netcommons`
* `ContainerIP` represents a container’s network configuration (interface name, IP version, address, etc.).  
* `StringLong` gives a full description useful for debugging.  
* `String` offers a compact form that is typically used in test output and logs.

```mermaid
flowchart TD
    A[ContainerIP] -->|String()| B[String]
    B --> C[StringLong]
```

This method enables the package to satisfy Go’s `fmt.Stringer` interface, allowing `ContainerIP` values to be printed with `%v`, `%s`, or in any format that relies on a string representation.
