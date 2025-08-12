ObjectSpec.AddField` – Documentation

#### Purpose
`AddField` is a helper method on the `ObjectSpec` struct that records an additional field to be inspected when displaying claim failures.  
It is used by the CLI command that prints failure details in various output formats (plain text, JSON, etc.). By adding a field name and its label, the command can later iterate over these entries to extract values from the underlying data structure.

#### Signature
```go
func (o *ObjectSpec) AddField(name string, label string)
```

| Parameter | Type   | Description |
|-----------|--------|-------------|
| `name`    | `string` | The internal field identifier that will be looked up in the object’s data map. |
| `label`   | `string` | A human‑readable label used when rendering output (e.g., column headers or JSON keys). |

The method returns **nothing** (`void`). It mutates the receiver.

#### Side Effects
- Appends a new entry to `o.Fields`, an internal slice of field descriptors.  
  ```go
  o.Fields = append(o.Fields, Field{name: name, label: label})
  ```
- No external state is modified; it only changes the state of the `ObjectSpec` instance.

#### Dependencies
| Dependency | Role |
|------------|------|
| `append` (built‑in) | Adds a new element to the slice. |

No other package variables or functions are referenced directly within this method.

#### Usage Context
Within the **failures** command (located at `cmd/certsuite/claim/show/failures`), an `ObjectSpec` is created for each type of claim failure that can be displayed.  
During initialization, fields relevant to that object type are registered via `AddField`. Later, when rendering results, the command iterates over these fields to extract and format data according to the chosen output format (`outputFormatFlag`).  

```go
spec := &ObjectSpec{}
spec.AddField("status", "Status")
spec.AddField("reason", "Reason")
// ...
```

#### Relationship to Package
`AddField` is a small but essential part of the *failures* sub‑command.  
It encapsulates the mapping between internal data keys and user‑facing labels, allowing the rest of the command logic (parsing flags, validating `outputFormatFlag`, generating JSON or text output) to remain agnostic about the underlying object schema.

---

**Mermaid diagram suggestion**

```mermaid
flowchart TD
    subgraph CLI
        A[User runs certsuite claim show failures]
        B[parseFlags() → claimFilePathFlag, testSuitesFlag, outputFormatFlag]
        C[loadSpec() → ObjectSpec with fields added via AddField]
        D[displayFailures(spec)]
    end

    subgraph AddField
        E1[Call: spec.AddField("status", "Status")]
        E2[Call: spec.AddField("reason", "Reason")]
    end

    A --> B --> C --> D
    C -->|fields populated by| E1 & E2
```

This diagram illustrates how `AddField` feeds the field list into the rendering pipeline of the failures command.
